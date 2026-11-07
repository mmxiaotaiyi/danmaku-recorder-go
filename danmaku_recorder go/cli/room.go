package cli

import (
	"fmt"
	"strconv"
)

func (a *App) cmdStart(args []string) {
	if len(args) != 1 {
		fmt.Println("用法: start <roomid>")
		return
	}
	roomID, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("房间号格式错误: %v\n", err)
		return
	}
	if err := a.h.AddRoom(roomID); err != nil {
		fmt.Printf("添加房间失败: %v\n", err)
		return
	}
	fmt.Printf("已连接到房间 %d\n", roomID)
}

func (a *App) cmdRemove(args []string) {
	if len(args) != 1 {
		fmt.Println("用法: remove <roomid>")
		return
	}
	roomID, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("房间号格式错误: %v\n", err)
		return
	}
	if err := a.h.DelRoom(roomID); err != nil {
		fmt.Printf("移除房间失败: %v\n", err)
		return
	}
	fmt.Printf("已移除房间 %d\n", roomID)
}

func (a *App) cmdList() {
	roomIDs := a.h.GetRooms()
	if len(roomIDs) == 0 {
		fmt.Println("当前没有监控的房间")
		return
	}
	fmt.Printf("在监控中的房间：")
	for i, roomID := range roomIDs {
		if i > 0 {
			fmt.Print("、")
		}
		fmt.Printf("%d", roomID)
	}
	fmt.Printf("，共 %d 个\n", len(roomIDs))
}
