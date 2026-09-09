package timer

import (
	"sync"
	"time"
)

type executor struct {
	addRoom   func(int) error
	remaining func() []ScheduledTask

	mu      sync.Mutex
	queue   []ScheduledTask
	dirty   bool
	first   bool
	stopCh  chan struct{}
	workCh  chan struct{}
	doneCh  chan struct{}
	persist *time.Timer
}

func newExecutor(addRoom func(int) error) *executor {
	t := time.NewTimer(time.Hour)
	if !t.Stop() {
		<-t.C
	}
	return &executor{
		addRoom: addRoom,
		first:   true,
		stopCh:  make(chan struct{}),
		workCh:  make(chan struct{}, 1),
		doneCh:  make(chan struct{}),
		persist: t,
	}
}

func (e *executor) start() {
	e.mu.Lock()
	e.queue = nil
	e.dirty = false
	e.first = true
	e.stopCh = make(chan struct{})
	e.doneCh = make(chan struct{})
	e.mu.Unlock()
	go e.loop()
}

func (e *executor) stop() {
	select {
	case <-e.stopCh:
	default:
		close(e.stopCh)
	}
	<-e.doneCh
	e.mu.Lock()
	e.queue = nil
	e.dirty = false
	e.mu.Unlock()
}

func (e *executor) enqueue(task ScheduledTask) {
	e.mu.Lock()
	e.queue = append(e.queue, task)
	e.mu.Unlock()
	select {
	case e.workCh <- struct{}{}:
	default:
	}
}

func (e *executor) queuedSnapshot() []ScheduledTask {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]ScheduledTask, len(e.queue))
	copy(out, e.queue)
	return out
}

func (e *executor) loop() {
	defer close(e.doneCh)
	for {
		select {
		case <-e.stopCh:
			e.flush()
			e.persist.Stop()
			return
		case <-e.persist.C:
			e.flush()
		case <-e.workCh:
			e.drain()
		}
	}
}

func (e *executor) drain() {
	for {
		select {
		case <-e.stopCh:
			return
		default:
		}

		e.mu.Lock()
		if len(e.queue) == 0 {
			e.first = true
			e.mu.Unlock()
			return
		}
		needGap := !e.first
		e.mu.Unlock()

		if needGap {
			gap := time.NewTimer(addRoomInterval)
			select {
			case <-e.stopCh:
				gap.Stop()
				return
			case <-gap.C:
			}
		}

		e.mu.Lock()
		task := e.queue[0]
		e.queue = e.queue[1:]
		e.first = false
		e.mu.Unlock()

		_ = e.addRoom(task.RoomID)

		e.mu.Lock()
		e.dirty = true
		e.mu.Unlock()
		if !e.persist.Stop() {
			select {
			case <-e.persist.C:
			default:
			}
		}
		e.persist.Reset(persistDebounce)
	}
}

func (e *executor) flush() {
	e.mu.Lock()
	if !e.dirty {
		e.mu.Unlock()
		return
	}
	e.dirty = false
	e.mu.Unlock()
	_ = saveTasks(e.remaining())
}
