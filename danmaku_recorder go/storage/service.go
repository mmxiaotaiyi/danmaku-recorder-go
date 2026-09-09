package storage

import (
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bytedance/sonic"
	log "github.com/sirupsen/logrus"
)

const (
	queueSize        = 512
	backlogWarnDepth = 128
	dropWarnEvery    = 100
	flushInterval    = 2 * time.Second
)

type RankGuard struct {
	Level      int32  `json:"level"`
	ExpiredStr string `json:"expired_str,omitempty"`
}

type RankItem struct {
	UID   int64      `json:"uid"`
	Score int64      `json:"score"`
	Name  string     `json:"name"`
	Rank  int32      `json:"rank"`
	Guard *RankGuard `json:"guard,omitempty"`
}

type RankData struct {
	Rank []RankItem `json:"rank"`
	Ts   []int64    `json:"ts"`
}

type UserData struct {
	Owner          int64                     `json:"owner"`
	RoomStartTs    string                    `json:"room_start_ts"`
	SessionStartTs string                    `json:"session_start_ts"`
	Members        map[string]map[string]any `json:"members"`
}

type encodedLine struct {
	cmd  string
	data []byte
}

type StorageService struct {
	roomID      int
	currentDir  string
	handles     map[string]*fileHandle
	filesClosed bool

	queue    chan encodedLine
	stopCh   chan struct{}
	doneCh   chan struct{}
	stopOnce sync.Once
	started  atomic.Bool
	stopped  atomic.Bool
	dropped  atomic.Uint64
	backlog  atomic.Bool

	RankData  *RankData
	UserData  *UserData
	LiveState bool
	LastMsgTs atomic.Int64
	EndTs     int64
}

func NewStorageService(roomID int) *StorageService {
	ss := &StorageService{
		roomID:   roomID,
		handles:  make(map[string]*fileHandle),
		queue:    make(chan encodedLine, queueSize),
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
		RankData: &RankData{},
		UserData: &UserData{
			Members: make(map[string]map[string]any),
		},
		LiveState: true,
	}
	if err := ss.initDirectory(); err != nil {
		log.Errorf("storage init directory error: room=%d err=%v", roomID, err)
	}
	return ss
}

func (ss *StorageService) Start() {
	if ss.started.CompareAndSwap(false, true) {
		go ss.writerLoop()
	}
}

func (ss *StorageService) AddLine(cmd string, data []byte) {
	if ss.stopped.Load() || len(data) == 0 {
		return
	}

	select {
	case ss.queue <- encodedLine{cmd: cmd, data: data}:
		if len(ss.queue) > backlogWarnDepth && ss.backlog.CompareAndSwap(false, true) {
			log.Warnf("storage queue backlog: room=%d depth=%d", ss.roomID, len(ss.queue))
		}
	default:
		n := ss.dropped.Add(1)
		if n == 1 || n%dropWarnEvery == 0 {
			log.Warnf("storage queue full: room=%d dropped=%d", ss.roomID, n)
		}
	}
}

func (ss *StorageService) Stop() {
	ss.stopOnce.Do(func() {
		if len(ss.UserData.Members) > 0 {
			if data, err := sonic.Marshal(ss.UserData); err != nil {
				log.Errorf("storage marshal error: room=%d cmd=UNIVERSAL_EVENT_GIFT err=%v", ss.roomID, err)
			} else {
				ss.AddLine("UNIVERSAL_EVENT_GIFT", data)
			}
		}

		ss.stopped.Store(true)
		if !ss.started.Load() {
			ss.closeFiles()
			return
		}

		close(ss.stopCh)
		<-ss.doneCh
	})
}

func (ss *StorageService) writerLoop() {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("storage writer panic: room=%d err=%v", ss.roomID, r)
			log.Debugf("storage writer stack: room=%d\n%s", ss.roomID, debug.Stack())
		}
		ss.closeFiles()
		close(ss.doneCh)
	}()

	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ss.stopCh:
			ss.drainQueue()
			return
		case line := <-ss.queue:
			ss.appendLine(line.cmd, line.data)
			ss.drainQueue()
			if len(ss.queue) <= backlogWarnDepth {
				ss.backlog.Store(false)
			}
		case <-ticker.C:
			ss.flushFiles()
		}
	}
}

func (ss *StorageService) drainQueue() {
	for {
		select {
		case line := <-ss.queue:
			ss.appendLine(line.cmd, line.data)
		default:
			return
		}
	}
}
