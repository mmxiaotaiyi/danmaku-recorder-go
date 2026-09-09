package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	bili "dm_client_go"
	"dm_client_go/live_check"
	"dm_client_go/timer"

	log "github.com/sirupsen/logrus"
)

func init() {
	bili.ChangeLogLevel(log.TraceLevel)
}

func main() {
	h := bili.GetNewHandler()
	reader := bufio.NewReader(os.Stdin)

	var taskManager *timer.TaskManager
	var liveCheckCLI *live_check.LiveCheckCLI

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("读取输入失败: %v\n", err)
			continue
		}
		input = strings.TrimSpace(input)
		if input == "exit" {
			h.CloseAll()
			break
		}

		if input == "list" {
			roomIDs := h.GetRooms()
			if len(roomIDs) == 0 {
				fmt.Println("当前没有监控的房间")
			} else {
				fmt.Printf("在监控中的房间：")
				for i, roomID := range roomIDs {
					if i > 0 {
						fmt.Print("、")
					}
					fmt.Printf("%d", roomID)
				}
				fmt.Printf("，共 %d 个\n", len(roomIDs))
			}
			continue
		}

		if strings.HasPrefix(input, "config ") {
			parts := strings.Fields(input)
			if len(parts) < 2 {
				fmt.Println("config命令格式错误。请使用:config cookies {cookies}")
				continue
			}

			subCmd := parts[1]
			switch subCmd {
			case "cookies":
				if len(parts) < 3 {
					fmt.Println("config cookies命令格式错误。请使用: config cookies {cookies}")
					continue
				}
				cookies := strings.Join(parts[2:], " ")
				if cookies == "" {
					fmt.Println("Cookie不能为空")
					continue
				}
				bili.SetHeaderCookie(cookies)
				fmt.Println("Cookie已更新")
				uid := extractDedeUserID(cookies)
				if uid > 0 {
					bili.SetUID(uid)
					fmt.Printf("已提取并设置UID: %d\n", uid)
				} else {
					fmt.Println("未能从Cookie中提取到有效的DedeUserID")
				}
			default:
				fmt.Println("未知的config子命令。请使用: config show 或 config cookies {cookies}")
			}
			continue
		}

		if input == "live check" {
			if liveCheckCLI == nil {
				liveCheckCLI = live_check.NewLiveCheckCLI(reader, h.AddRoom, h.GetRooms)
			}
			liveCheckCLI.Run()
			continue
		}
		if strings.HasPrefix(input, "timer ") {
			parts := strings.Fields(input)
			if len(parts) < 2 {
				fmt.Println("timer命令格式错误。请使用: timer start 或 timer stop 或 timer show")
				continue
			}

			subCmd := parts[1]
			switch subCmd {
			case "start":
				if taskManager == nil {
					taskManager = timer.NewTaskManager(h.AddRoom)
				}
				if err := taskManager.StartAllTasks(); err != nil {
					fmt.Printf("启动定时任务失败: %v\n", err)
				} else {
					fmt.Println("定时任务已启动")
				}
			case "stop":
				if taskManager != nil {
					taskManager.StopAllTasks()
				}
				fmt.Println("定时任务已停止")
			case "show":
				if taskManager == nil {
					fmt.Println("定时任务未启动")
					continue
				}
				tasks := taskManager.ShowPendingTasks()
				if len(tasks) == 0 {
					fmt.Println("当前没有待执行的定时任务")
				} else {
					fmt.Println("待执行的定时任务：")
					for _, task := range tasks {
						fmt.Printf("房间号: %d, 执行时间: %s\n", task.RoomID, task.TriggerAt.Format("2006-01-02 15:04:05"))
					}
				}

			default:
				fmt.Println("未知的timer子命令。请使用: timer start 或 timer stop 或 timer show")
			}
			continue
		}

		parts := strings.Fields(input)
		if len(parts) != 2 {
			fmt.Println("命令格式错误。请使用: start {roomid} 或 remove {roomid} 或 list 或 timer {subcommand} 或 live check 或 config {subcommand}")
			continue
		}
		cmd, roomIdStr := parts[0], parts[1]
		roomId, err := strconv.Atoi(roomIdStr)
		if err != nil {
			fmt.Printf("房间号格式错误: %v\n", err)
			continue
		}
		switch cmd {
		case "start":
			if err := h.AddRoom(roomId); err != nil {
				fmt.Printf("添加房间失败: %v\n", err)
			} else {
				fmt.Printf("已连接到房间 %d\n", roomId)
			}
		case "remove":
			if err := h.DelRoom(roomId); err != nil {
				fmt.Printf("移除房间失败: %v\n", err)
			} else {
				fmt.Printf("已移除房间 %d\n", roomId)
			}
		default:
			fmt.Println("未知命令。请使用: start {roomid} 或 remove {roomid} 或 list 或 timer {subcommand} 或 live check 或 config {subcommand}")
			continue
		}
	}
}

func extractDedeUserID(cookies string) int64 {
	re := regexp.MustCompile(`DedeUserID=(\d+)`)
	matches := re.FindStringSubmatch(cookies)
	if len(matches) >= 2 {
		uid, err := strconv.ParseInt(matches[1], 10, 64)
		if err == nil {
			return uid
		}
	}
	return 0
}
