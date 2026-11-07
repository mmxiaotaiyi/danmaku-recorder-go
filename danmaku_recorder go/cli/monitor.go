package cli

import (
	"fmt"

	"dm_client_go/live_check"
)

func (a *App) cmdMonitor(args []string) {
	if len(args) != 1 {
		printMonitorUsage()
		return
	}
	switch args[0] {
	case "load":
		a.monitorLoad()
	case "start":
		a.monitorStart()
	case "stop":
		a.monitorStop()
	case "list":
		a.monitorList()
	case "clear":
		a.monitorClear()
	default:
		printMonitorUsage()
	}
}

func (a *App) monitorLoad() {
	if a.monitor == nil {
		a.monitor = live_check.NewLiveCheck(a.h.AddRoom, a.h.GetRooms)
	}
	if err := a.monitor.LoadConfig(); err != nil {
		fmt.Printf("加载配置文件失败: %v\n", err)
	}
}

func (a *App) monitorStart() {
	if a.monitor == nil {
		fmt.Println("请先执行 monitor load")
		return
	}
	if a.monitorStarted {
		fmt.Println("直播检查已启动")
		return
	}
	if err := a.monitor.Start(); err != nil {
		fmt.Printf("启动失败: %v\n", err)
		return
	}
	a.monitorStarted = true
	fmt.Println("直播检查已启动")
}

func (a *App) monitorStop() {
	if a.monitor == nil || !a.monitorStarted {
		fmt.Println("直播检查未启动")
		return
	}
	a.monitor.Stop()
	a.monitor = nil
	a.monitorStarted = false
	fmt.Println("直播检查已停止")
}

func (a *App) monitorList() {
	if a.monitor == nil {
		fmt.Println("直播检查未初始化")
		return
	}
	roomIDs := a.monitor.RoomIDs
	fmt.Printf("当前房间(%d 个):\n", len(roomIDs))
	fmt.Printf("房间列表：%v\n", roomIDs)
}

func (a *App) monitorClear() {
	if a.monitor == nil {
		fmt.Println("直播检查未初始化")
		return
	}
	a.monitor.RoomIDs = make([]int, 0)
	fmt.Println("房间列表已清空")
}
