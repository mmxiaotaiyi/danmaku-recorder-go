package cli

import (
	"fmt"

	"dm_client_go/timer"
)

func (a *App) cmdTimer(args []string) {
	if len(args) != 1 {
		printTimerUsage()
		return
	}
	switch args[0] {
	case "start":
		if a.timer == nil {
			a.timer = timer.NewTaskManager(a.h.AddRoom)
		}
		if err := a.timer.StartAllTasks(); err != nil {
			fmt.Printf("启动定时任务失败: %v\n", err)
			return
		}
		a.timerStarted = true
		fmt.Println("定时任务已启动")
	case "stop":
		if a.timer == nil || !a.timerStarted {
			fmt.Println("定时任务未启动")
			return
		}
		a.timer.StopAllTasks()
		a.timerStarted = false
		fmt.Println("定时任务已停止")
	case "show":
		if a.timer == nil || !a.timerStarted {
			fmt.Println("定时任务未启动")
			return
		}
		tasks := a.timer.ShowPendingTasks()
		if len(tasks) == 0 {
			fmt.Println("当前没有待执行的定时任务")
			return
		}
		fmt.Println("待执行的定时任务：")
		for _, task := range tasks {
			fmt.Printf("房间号: %d, 执行时间: %s\n", task.RoomID, task.TriggerAt.Format("2006-01-02 15:04:05"))
		}
	default:
		printTimerUsage()
	}
}
