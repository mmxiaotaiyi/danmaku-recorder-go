package client

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"dm_client_go/storage"

	"github.com/lxzan/gws"

	log "github.com/sirupsen/logrus"
)

var UID = int64(0)
var Buvid = ""
var Header http.Header

type Client struct {
	RoomInfo WsAuthBody
	connect  *gws.Conn
	handler  MsgHandler
	storage  *storage.StorageService
	done     chan struct{}
	stopOnce sync.Once
	closing  atomic.Bool
}

func (c *Client) OnOpen(socket *gws.Conn) {
	wsAuthMsg := WsAuthMessage{Body: c.RoomInfo}
	if err := c.sendAuthMsg(wsAuthMsg); err != nil {
		log.Warn("send auth msg to websocket error: ", err)
		return
	}
	go c.heartBeat()
}

func (c *Client) OnClose(socket *gws.Conn, err error) {
	log.Info("disconnected from blive: ", c.RoomInfo.Roomid)
	if c.closing.Load() {
		return
	}
	c.cancel()
	time.Sleep(2 * time.Second)
	if c.closing.Load() {
		return
	}
	c.connectLoop()
}

func (c *Client) OnPing(socket *gws.Conn, payload []byte) {
}

func (c *Client) OnPong(socket *gws.Conn, payload []byte) {
}

func (c *Client) OnMessage(socket *gws.Conn, message *gws.Message) {
	defer message.Close()
	c.handler.MsgHandler(message.Bytes())
}

func generateBuvid() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func init() {
	Header = make(http.Header)
	Header.Set("Accept-Encoding", "gzip, deflate, br")
	Header.Set("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en;q=0.3,en-US;q=0.2")
	Header.Set("Accept", "application/json, text/plain, */*")
	Header.Set("Cache-Control", "no-cache")
	Header.Set("Connection", "keep-alive")
	Header.Set("Origin", "https://live.bilibili.com")
	Header.Set("Pragma", "no-cache")
	Header.Set("Referer", "https://live.bilibili.com")
	Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36")

	// 生成buvid并设置cookie
	buvid := generateBuvid()
	timestamp := time.Now().Unix()
	randomNum := rand.Intn(9000000000) + 1000000000
	cookie := fmt.Sprintf("buvid3=%s; b_nut=%d; i-wanna-go-back=-1; _uuid=%sinfoc; buvid4=%s-%d-%d",
		buvid, timestamp, buvid, buvid, randomNum, timestamp)
	Header.Set("Cookie", cookie)
}

func (c *Client) biliChatConnect(url string) error {
	var err error
	c.connect, _, err = gws.NewClient(c, &gws.ClientOption{
		Addr:      url,
		TlsConfig: &tls.Config{InsecureSkipVerify: true},
		PermessageDeflate: gws.PermessageDeflate{
			Enabled:               true,
			ServerContextTakeover: true,
			ClientContextTakeover: true,
		},
		RequestHeader: Header,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) sendAuthMsg(wsAuthMsg WsAuthMessage) error {
	wsPackage := wsAuthMsg.GetPackage()
	err := c.connect.WriteMessage(gws.OpcodeBinary, wsPackage)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) heartBeat() {
	t := time.NewTicker(time.Second * 30)
	for {
		select {
		case <-c.done:
			return
		case <-t.C:
			if c.connect == nil {
				return
			}
			heartBeatPackage := WsHeartBeatMessage{Body: []byte{}}
			err := c.connect.WriteMessage(gws.OpcodeBinary, heartBeatPackage.GetPackage())
			if err != nil {
				log.Error("failed to send heart beat to blive: ", err)
				return
			}
		}
	}
}

func (c *Client) sendConnect() error {
	apiLiveAuth, err := getLiveRoomAuth(c.RoomInfo.Roomid)
	if err != nil {
		return err
	} else if apiLiveAuth.Code != 0 {
		log.Warnf("get live room info error: %v", apiLiveAuth.Message)
		return ErrRespCodeNotZero
	}
	if c.RoomInfo.Key == "" {
		c.RoomInfo.Key = apiLiveAuth.Data.Token
	}
	for nowSum, i := range apiLiveAuth.Data.HostList {
		u := url.URL{Scheme: "wss", Host: i.Host + ":" + strconv.Itoa(i.WssPort), Path: "/sub"}
		err = c.biliChatConnect(u.String())
		if err != nil {
			log.Warnf("connect to blive websocket error for %d time: %v\n", nowSum, err)
			if nowSum == 2 {
				return err
			}
		} else {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

func (c *Client) resetForReconnect() {
	c.done = make(chan struct{})
	c.stopOnce = sync.Once{}
}

func (c *Client) connectLoop() {
	c.resetForReconnect()
	for {
		select {
		case <-c.done:
			return
		default:
			err := c.sendConnect()
			if err != nil {
				log.Warn("connect to blive error: ", err)
				time.Sleep(5 * time.Second)
			} else {
				go c.connect.ReadLoop()
				return
			}
		}
	}
}

func (c *Client) cancel() {
	if c.done == nil {
		return
	}
	c.stopOnce.Do(func() {
		close(c.done)
	})
}

func (c *Client) Close() {
	if !c.closing.CompareAndSwap(false, true) {
		return
	}
	c.cancel()
	if c.storage != nil {
		c.storage.Stop()
	}
	if c.connect != nil {
		c.connect.WriteClose(1000, nil)
	}
}

func (c *Client) BiliChat() {
	defer func() {
		if err := recover(); err != nil {
			log.Warnf("start blive panic: %v", err)
		}
	}()

	c.storage = storage.NewStorageService(c.RoomInfo.Roomid)
	c.storage.Start()

	c.handler = MsgHandler{
		RoomId:  c.RoomInfo.Roomid,
		storage: c.storage,
	}

	c.connectLoop()
	log.Debug("start blive success: ", c.RoomInfo.Roomid)
}
