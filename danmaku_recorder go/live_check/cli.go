package live_check

import (
	"bufio"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

type LiveCheckCLI struct {
	liveCheck *LiveCheck
	reader    *bufio.Reader
	addRoom   func(roomID int) error
	getRooms  func() []int
}

func NewLiveCheckCLI(reader *bufio.Reader, addRoom func(roomID int) error, getRooms func() []int) *LiveCheckCLI {
	return &LiveCheckCLI{
		liveCheck: nil,
		reader:    reader,
		addRoom:   addRoom,
		getRooms:  getRooms,
	}
}

func (cli *LiveCheckCLI) Run() {
	for {
		fmt.Print("live_check> ")
		input, err := cli.reader.ReadString('\n')
		if err != nil {
			log.Errorf("读取输入失败: %v", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		command := parts[0]

		switch command {
		case "load":
			cli.handleLoad()
		case "list":
			cli.handleList()
		case "start":
			cli.handleStart()
		case "stop":
			cli.handleStop()
		case "clear":
			cli.handleClear()
		case "back":
			fmt.Println("退出 Live Check CLI")
			return
		default:
			fmt.Printf("未知命令.请使用: load, list, start, stop, clear, back\n")
		}
	}
}

func (cli *LiveCheckCLI) handleLoad() {
	if cli.liveCheck == nil {
		fmt.Println("直播检查未启动")
		return
	}
	if err := cli.liveCheck.LoadConfig(); err != nil {
		fmt.Printf("加载配置文件失败: %v\n", err)
	}
}

func (cli *LiveCheckCLI) handleList() {
	if cli.liveCheck == nil {
		fmt.Println("直播检查未启动")
		return
	}
	roomIDs := cli.liveCheck.RoomIDs

	fmt.Printf("当前房间(%d 个):\n", len(roomIDs))
	fmt.Printf("房间列表：%v\n", roomIDs)
}

func (cli *LiveCheckCLI) handleStart() {
	if cli.liveCheck != nil {
		fmt.Println("直播检查已启动")
		return
	}
	cli.liveCheck = NewLiveCheck(cli.addRoom, cli.getRooms)
	cli.liveCheck.LoadConfig()
	err := cli.liveCheck.Start()
	if err != nil {
		fmt.Printf("启动失败: %v\n", err)
	} else {
		fmt.Println("直播检查已启动")
	}
}

func (cli *LiveCheckCLI) handleStop() {
	if cli.liveCheck != nil {
		cli.liveCheck.Stop()
		cli.liveCheck = nil
		fmt.Println("直播检查已停止")
	} else {
		fmt.Println("直播检查未启动")
	}
}

func (cli *LiveCheckCLI) handleClear() {
	if cli.liveCheck == nil {
		fmt.Println("直播检查未启动")
		return
	}
	cli.liveCheck.RoomIDs = make([]int, 0)
	fmt.Println("房间列表已清空")
}
