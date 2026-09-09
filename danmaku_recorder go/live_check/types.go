package live_check

import (
	"net/http"
	"time"
)

const (
	UserAgent      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
	AcceptEncoding = "gzip, deflate, br, zstd"
	AcceptLanguage = "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6"
	Accept         = "application/json, text/plain, */*"
	CacheControl   = "no-cache"
	Pragma         = "no-cache"
	ApiUrl         = "https://api.live.bilibili.com/xlive/web-room/v1/index/getRoomBaseInfo?req_biz=web_room_componet"
	FetchBatchSize = 30
	FetchBatchWait = time.Second
)

type LiveCheck struct {
	addRoom  func(roomID int) error
	getRooms func() []int
	client   *http.Client
	RoomIDs  []int
	stopCh   chan struct{}
}
