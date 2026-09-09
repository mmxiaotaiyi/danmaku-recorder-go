package client

import (
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/bytedance/sonic"
	log "github.com/sirupsen/logrus"
)

const (
	UserAgent      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36"
	BiliLiveApiUrl = "api.live.bilibili.com"
)

const (
	AppKey = "1d8b6e7d45233436"
	AppSec = "560c52ccd288fed045859ed18bffd973"
)

var shortRoomId2Long = make(map[int]int)

func getReq(data url.Values, getUrl string) ([]byte, error) {
	u, err := url.ParseRequestURI(getUrl)
	if err != nil {
		return nil, err
	}
	u.RawQuery = data.Encode()
	client := http.Client{}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header = Header
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	if resp.StatusCode != 200 {
		return nil, ErrRespCodeNotZero
	}

	rawData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	contentEncoding := resp.Header.Get("Content-Encoding")

	switch strings.ToLower(contentEncoding) {
	case "gzip":
		reader, err := gzip.NewReader(strings.NewReader(string(rawData)))
		if err != nil {
			log.Error("Failed to create gzip reader: ", err)
			return nil, err
		}
		defer reader.Close()
		decompressed, err := io.ReadAll(reader)
		if err != nil {
			log.Error("Failed to decompress gzip data: ", err)
			return nil, err
		}

		return decompressed, nil

	case "br":
		reader := brotli.NewReader(strings.NewReader(string(rawData)))
		decompressed, err := io.ReadAll(reader)
		if err != nil {
			log.Error("Failed to decompress brotli data: ", err)
			return nil, err
		}
		return decompressed, nil

	default:
		return rawData, nil
	}
}

func signed(params map[string]string) map[string]string {
	params["appkey"] = AppKey

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var queryParts []string
	for _, k := range keys {
		queryParts = append(queryParts, fmt.Sprintf("%s=%s", k, params[k]))
	}
	query := strings.Join(queryParts, "&")

	sign := fmt.Sprintf("%x", md5.Sum([]byte(query+AppSec)))
	params["sign"] = sign

	return params
}

func getLiveRoomAuth(roomId int) (ApiLiveAuth, error) {
	getUrl := url.URL{Scheme: "https", Host: BiliLiveApiUrl, Path: "/xlive/app-room/v1/index/getDanmuInfo"}

	params := map[string]string{
		"actionKey": "appkey",
		"build":     "6640400",
		"channel":   "bili",
		"device":    "android",
		"mobi_app":  "android",
		"platform":  "android",
		"room_id":   strconv.Itoa(roomId),
		"ts":        strconv.FormatInt(time.Now().Unix(), 10),
	}

	signedParams := signed(params)

	data := url.Values{}
	for k, v := range signedParams {
		data.Set(k, v)
	}

	s, err := getReq(data, getUrl.String())
	if err != nil {
		return ApiLiveAuth{}, err
	}

	var j ApiLiveAuth
	err = sonic.Unmarshal(s, &j)
	if err != nil {
		return ApiLiveAuth{}, err
	}
	return j, nil
}

func GetRealRoomId(roomId int) (int, error) {
	if _, ok := shortRoomId2Long[roomId]; ok {
		return shortRoomId2Long[roomId], nil
	}
	getUrl := url.URL{Scheme: "https", Host: BiliLiveApiUrl, Path: "/xlive/web-room/v1/index/getRoomPlayInfo"}
	data := url.Values{}
	data.Set("room_id", strconv.Itoa(roomId))
	s, err := getReq(data, getUrl.String())
	if err != nil {
		return 0, err
	}
	var j ApiLiveRoomId
	err = sonic.Unmarshal(s, &j)
	if err != nil {
		return 0, err
	}
	shortRoomId2Long[roomId] = j.Data.RoomID
	return j.Data.RoomID, nil
}
