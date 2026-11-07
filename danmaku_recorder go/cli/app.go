package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	bili "dm_client_go"
	"dm_client_go/live_check"
	"dm_client_go/timer"
)

type App struct {
	h              *bili.Handler
	timer          *timer.TaskManager
	timerStarted   bool
	monitor        *live_check.LiveCheck
	monitorStarted bool
}

func Run() {
	a := &App{h: bili.GetNewHandler()}
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("读取输入失败: %v\n", err)
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if line == "exit" {
			a.h.CloseAll()
			return
		}
		a.dispatch(line)
	}
}

func (a *App) dispatch(line string) {
	if cmd, ok := firstWord(line); ok && cmd == "setcookies" {
		a.cmdSetCookies(line)
		return
	}

	fields := strings.Fields(line)
	if len(fields) == 0 {
		return
	}
	cmd, args := fields[0], fields[1:]
	switch cmd {
	case "help":
		printHelp()
	case "start":
		a.cmdStart(args)
	case "remove":
		a.cmdRemove(args)
	case "list":
		a.cmdList()
	case "timer":
		a.cmdTimer(args)
	case "monitor":
		a.cmdMonitor(args)
	default:
		fmt.Println("未知命令。请输入 help 查看用法")
	}
}

func firstWord(line string) (string, bool) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return "", false
	}
	return fields[0], true
}
