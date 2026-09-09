package live_check

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func (lc *LiveCheck) GetLiveRooms() ([]map[string]any, error) {
	if len(lc.RoomIDs) == 0 {
		return nil, nil
	}

	var byRoomIDs []gjson.Result
	var firstErr error
	for i := 0; i < len(lc.RoomIDs); i += FetchBatchSize {
		if i > 0 && !lc.waitBatch() {
			break
		}

		end := min(i+FetchBatchSize, len(lc.RoomIDs))
		data, err := lc.fetchLiveRoomsBatch(lc.RoomIDs[i:end])
		if err != nil {
			log.Errorf("拉取房间信息失败：%v", err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		batch, err := parseByRoomIDs(data)
		if err != nil {
			log.Errorf("解析房间信息失败：%v", err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		byRoomIDs = append(byRoomIDs, batch)
	}
	if len(byRoomIDs) == 0 {
		return nil, firstErr
	}
	return lc.extractRooms(byRoomIDs), nil
}

func (lc *LiveCheck) fetchLiveRoomsBatch(roomIDs []int) ([]byte, error) {
	var b strings.Builder
	b.WriteString(ApiUrl)
	for _, roomID := range roomIDs {
		fmt.Fprintf(&b, "&room_ids=%d", roomID)
	}

	req, err := http.NewRequest("GET", b.String(), nil)
	if err != nil {
		log.Errorf("创建请求失败: %v", err)
		return nil, err
	}

	req.Header.Set("Accept-Encoding", AcceptEncoding)
	req.Header.Set("Accept-Language", AcceptLanguage)
	req.Header.Set("Accept", Accept)
	req.Header.Set("Cache-Control", CacheControl)
	req.Header.Set("Pragma", Pragma)
	req.Header.Set("User-Agent", UserAgent)

	resp, err := lc.client.Do(req)
	if err != nil {
		log.Errorf("发送请求失败: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	rawData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("读取响应体失败: %v", err)
		return nil, err
	}

	return decompressBody(resp.Header.Get("Content-Encoding"), rawData)
}

func (lc *LiveCheck) waitBatch() bool {
	timer := time.NewTimer(FetchBatchWait)
	defer timer.Stop()
	select {
	case <-lc.stopCh:
		return false
	case <-timer.C:
		return true
	}
}

func decompressBody(enc string, raw []byte) ([]byte, error) {
	switch enc {
	case "gzip":
		reader, err := gzip.NewReader(bytes.NewReader(raw))
		if err != nil {
			log.Errorf("创建gzip读取器失败: %v", err)
			return nil, err
		}
		defer reader.Close()
		data, err := io.ReadAll(reader)
		if err != nil {
			log.Errorf("解码gzip数据失败: %v", err)
			return nil, err
		}
		return data, nil
	case "br":
		data, err := io.ReadAll(brotli.NewReader(bytes.NewReader(raw)))
		if err != nil {
			log.Errorf("解码brotli数据失败: %v", err)
			return nil, err
		}
		return data, nil
	case "":
		return raw, nil
	default:
		log.Warnf("未知的压缩格式: %s，使用原始数据", enc)
		return raw, nil
	}
}

func parseByRoomIDs(data []byte) (gjson.Result, error) {
	jsonStr := string(data)
	if code := gjson.Get(jsonStr, "code").Int(); code != 0 {
		message := gjson.Get(jsonStr, "message").String()
		log.Errorf("API错误: %s (code: %d)", message, code)
		return gjson.Result{}, errors.New(message)
	}
	return gjson.Get(jsonStr, "data.by_room_ids"), nil
}

func (lc *LiveCheck) extractRooms(byRoomIDs []gjson.Result) []map[string]any {
	RoomIDsMap := make(map[int]bool)
	for _, roomID := range lc.getRooms() {
		RoomIDsMap[roomID] = true
	}

	var roomInfos []map[string]any
	for _, byRoomIds := range byRoomIDs {
		byRoomIds.ForEach(func(key, room gjson.Result) bool {
			if room.Get("live_status").Int() != 1 {
				return true
			}

			roomID := int(room.Get("room_id").Int())
			if RoomIDsMap[roomID] {
				return true
			}

			roomInfos = append(roomInfos, map[string]any{
				"room_id":          roomID,
				"title":            room.Get("title").Str,
				"name":             room.Get("uname").Str,
				"uid":              room.Get("uid").Int(),
				"ts":               room.Get("live_time").Str,
				"area_parent_name": room.Get("parent_area_name").Str,
				"area_parent_id":   room.Get("parent_area_id").Int(),
				"area_name":        room.Get("area_name").Str,
				"area_id":          room.Get("area_id").Int(),
			})
			return true
		})
	}
	return roomInfos
}
