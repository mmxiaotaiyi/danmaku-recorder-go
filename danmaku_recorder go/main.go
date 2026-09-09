package go_bili_chat

import (
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"dm_client_go/message"

	"dm_client_go/client"

	log "github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

var ErrGetRoomFailed = errors.New("get room failed")
var ErrRoomAlreadyExist = errors.New("room already exist")
var ErrRoomNotExist = errors.New("room not exist")

type Handler struct {
	rooms sync.Map
}

type LiveRoom struct {
	RoomId int
	Client *client.Client
}

func setupLogFile() {
	dateStr := time.Now().Format("2006-01-02")
	logFileName := dateStr + ".log"
	fileInfo, err := os.Stat(logFileName)
	if err == nil && fileInfo.Size() > 10*1024 {
		f, err := os.Create(logFileName)
		if err != nil {
			log.Errorf("无法清空日志文件: %v", err)
			return
		}
		f.Close()
	}
	f, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Errorf("无法打开日志文件: %v", err)
		return
	}
	log.SetOutput(f)
}

func init() {
	setupLogFile()
	log.SetFormatter(&easy.Formatter{
		TimestampFormat: "01-02 15:04:05",
		LogFormat:       "[bili-live][%time%][%lvl%]: %msg% \n",
	})
	ChangeLogLevel(log.ErrorLevel)
}

func ChangeLogLevel(level log.Level) {
	log.SetLevel(level)
}

func SetHeader(header http.Header) {
	client.Header = header
}

func SetHeaderUA(ua string) {
	client.Header.Set("User-Agent", ua)
}

func SetBuvid(Buvid string) {
	client.Buvid = Buvid
}

func SetHeaderCookie(cookie string) {
	client.Header.Set("Cookie", cookie)
}

func SetUID(uid int64) {
	client.UID = uid
}

func GetNewHandler() *Handler {
	h := Handler{}
	message.GlobalHandler = &h

	return &h
}

func (h *Handler) ExistRoom(RoomId int) bool {
	if _, ok := h.rooms.Load(RoomId); ok {
		return true
	}
	return false
}

func (h *Handler) AddRawRoom(RawRoom client.WsAuthBody) error {
	if _, ok := h.rooms.Load(RawRoom.Roomid); ok {
		return ErrRoomAlreadyExist
	}
	room := &LiveRoom{
		RoomId: RawRoom.Roomid,
		Client: &client.Client{RoomInfo: RawRoom},
	}
	room.Client.BiliChat()
	h.rooms.Store(room.RoomId, room)
	return nil
}

func (h *Handler) AddRoom(RoomId int) error {
	if RoomId <= 100000 {
		RealRoomId, err := client.GetRealRoomId(RoomId)
		if err != nil {
			return err
		} else if RealRoomId == 0 {
			return ErrGetRoomFailed
		}
		log.Debug(RoomId, " is short roomid, the real roomid is: ", RealRoomId)
		RoomId = RealRoomId
	}
	return h.AddRawRoom(client.WsAuthBody{
		Roomid:   RoomId,
		Protover: 3,
		UID:      client.UID,
		Buvid:    client.Buvid,
		Type:     2,
		Platform: "web",
		Key:      "",
	})
}

func (h *Handler) DelRoom(RoomId int) error {
	if RoomId <= 100000 {
		RealRoomId, err := client.GetRealRoomId(RoomId)
		if err != nil {
			return err
		} else if RealRoomId == 0 {
			return ErrGetRoomFailed
		}
		log.Debug(RoomId, " is short roomid, the real roomid is: ", RealRoomId)
		RoomId = RealRoomId
	}
	if _, ok := h.rooms.Load(RoomId); !ok {
		return ErrRoomNotExist
	}
	c, ok := h.rooms.Load(RoomId)
	if !ok {
		return ErrRoomNotExist
	}
	defer h.rooms.Delete(RoomId)

	room := c.(*LiveRoom)
	room.Client.Close()
	return nil
}

func (h *Handler) GetRooms() []int {
	var roomIDs []int
	h.rooms.Range(func(key, value any) bool {
		roomID := key.(int)
		roomIDs = append(roomIDs, roomID)
		return true
	})
	return roomIDs
}

func (h *Handler) Wait() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	h.CloseAll()
}

func (h *Handler) CloseAll() {
	h.rooms.Range(func(key, value any) bool {
		room := value.(*LiveRoom)
		room.Client.Close()
		return true
	})

	h.rooms = sync.Map{}
}
