package timer

import (
	"fmt"
	"sync"
)

type TaskManager struct {
	scheduler *scheduler
	executor  *executor
	mutex     sync.Mutex
	isRunning bool
}

func NewTaskManager(addRoomFunc func(int) error) *TaskManager {
	exec := newExecutor(addRoomFunc)
	sched := newScheduler(exec)
	exec.remaining = func() []ScheduledTask {
		return append(sched.snapshot(), exec.queuedSnapshot()...)
	}
	return &TaskManager{
		scheduler: sched,
		executor:  exec,
	}
}

func (tm *TaskManager) StartAllTasks() error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if tm.isRunning {
		fmt.Println("定时任务已在运行中，正在重启...")
		tm.stopLocked()
	}

	tasks, err := loadTasks()
	if err != nil {
		return err
	}
	tm.executor.start()
	tm.scheduler.start(tasks)
	tm.isRunning = true
	return nil
}

func (tm *TaskManager) StopAllTasks() {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	if !tm.isRunning {
		fmt.Println("定时任务未在运行")
		return
	}
	tm.stopLocked()
}

func (tm *TaskManager) stopLocked() {
	tm.scheduler.stop()
	tm.executor.stop()
	tm.scheduler.clear()
	tm.isRunning = false
}

func (tm *TaskManager) ShowPendingTasks() []ScheduledTask {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	if !tm.isRunning {
		return nil
	}
	return tm.executor.remaining()
}
