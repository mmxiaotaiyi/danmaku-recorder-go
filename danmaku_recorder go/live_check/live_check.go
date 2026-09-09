package live_check

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func NewLiveCheck(addRoom func(roomID int) error, getRooms func() []int) *LiveCheck {
	return &LiveCheck{
		addRoom:  addRoom,
		getRooms: getRooms,
		client:   &http.Client{},
		RoomIDs:  make([]int, 0),
		stopCh:   make(chan struct{}),
	}
}

func (lc *LiveCheck) LoadConfig() error {
	configFile := "./up_info.jsonl"
	file, err := os.Open(configFile)
	if err != nil {
		log.Errorf("打开配置文件失败: %v\n", err)
		return err
	}
	defer file.Close()

	existingRooms := make(map[int]bool)
	for _, roomID := range lc.RoomIDs {
		existingRooms[roomID] = true
	}

	addedCount := 0
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		uid := gjson.Get(line, "0").Int()
		roomID := gjson.Get(line, "1").Int()
		name := gjson.Get(line, "2").Str
		room_ID := int(roomID)
		if roomID < 1 {
			fmt.Printf("房间ID不合法: %s (房间号: %d)\n", line, roomID)
			continue
		}
		if existingRooms[room_ID] {
			continue
		}

		fmt.Printf("UID: %d, 房间号: %d, 名称: %s\n", uid, roomID, name)
		lc.RoomIDs = append(lc.RoomIDs, room_ID)
		existingRooms[room_ID] = true
		addedCount++
	}

	if err := scanner.Err(); err != nil {
		log.Errorf("读取文件时发生错误: %v\n", err)
		return err
	}
	fmt.Printf("已存在房间：%v\n", lc.RoomIDs)
	fmt.Printf("新增房间数量: %d, 总房间数量: %d\n", addedCount, len(lc.RoomIDs))
	return nil
}

func (lc *LiveCheck) Start() error {
	go func() {
		if err := lc.checkAndAddRooms(); err != nil {
			log.Errorf("初始检查失败: %v", err)
		}
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-lc.stopCh:
				log.Info("LiveCheck goroutine 已停止")
				return
			case <-ticker.C:
				if err := lc.checkAndAddRooms(); err != nil {
					log.Errorf("检查房间失败: %v", err)
				}
			}
		}
	}()
	return nil
}

func (lc *LiveCheck) Stop() {
	close(lc.stopCh)
}

func (lc *LiveCheck) checkAndAddRooms() error {
	rooms, err := lc.GetLiveRooms()
	if err != nil {
		log.Errorf("获取直播房间列表失败: %v", err)
		return err
	}

	if len(rooms) == 0 {
		return nil
	}
	lc.Save(rooms)
	for i, room := range rooms {
		go func(room map[string]any, index int) {
			time.Sleep(time.Duration(index*3) * time.Second)
			roomID := room["room_id"]
			title := room["title"]
			name := room["name"]
			if err := lc.addRoom(roomID.(int)); err != nil {
				log.Errorf("添加房间 %d 失败: %v", roomID, err)
			} else {
				fmt.Printf("房间启动: %d: %s (%s)\n", roomID, title, name)
			}
		}(room, i)
	}

	return nil
}

func (lc *LiveCheck) Save(rooms []map[string]any) error {
	recordDir := "record"
	if err := os.MkdirAll(recordDir, 0755); err != nil {
		log.Errorf("创建record目录失败: %v", err)
		return err
	}

	today := time.Now().Format("2006-01-02")
	filename := filepath.Join(recordDir, today+".jsonl")

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Errorf("打开文件失败: %v", err)
		return err
	}
	defer file.Close()

	for _, room := range rooms {
		jsonData, _ := json.Marshal(room)
		if _, err := file.WriteString(string(jsonData) + "\n"); err != nil {
			log.Errorf("写入房间信息失败: %v", err)
			return err
		}
	}
	return nil
}
