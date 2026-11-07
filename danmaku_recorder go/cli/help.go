package cli

import "fmt"

func printHelp() {
	fmt.Print(`setcookies <cookie> | start/remove <roomid> | list | help | exit
timer start | stop | show
monitor load | start | stop | list | clear
`)
}

func printTimerUsage() {
	fmt.Println("用法: timer start | stop | show")
}

func printMonitorUsage() {
	fmt.Println("用法: monitor load | start | stop | list | clear")
}
