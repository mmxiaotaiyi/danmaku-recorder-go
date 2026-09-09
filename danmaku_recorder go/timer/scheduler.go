package timer

import (
	"sync"
	"time"
)

type scheduler struct {
	exec *executor

	mu      sync.Mutex
	pending []ScheduledTask
	timer   *time.Timer
	stopCh  chan struct{}
	doneCh  chan struct{}
}

func newScheduler(exec *executor) *scheduler {
	t := time.NewTimer(time.Hour)
	if !t.Stop() {
		<-t.C
	}
	return &scheduler{
		exec:   exec,
		timer:  t,
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}
}

func (s *scheduler) start(tasks []ScheduledTask) {
	s.mu.Lock()
	s.pending = append([]ScheduledTask(nil), tasks...)
	s.stopCh = make(chan struct{})
	s.doneCh = make(chan struct{})
	s.mu.Unlock()
	s.armTimer()
	go s.loop()
}

func (s *scheduler) stop() {
	select {
	case <-s.stopCh:
		return
	default:
		close(s.stopCh)
	}
	<-s.doneCh
}

func (s *scheduler) clear() {
	s.mu.Lock()
	s.pending = nil
	s.mu.Unlock()
}

func (s *scheduler) snapshot() []ScheduledTask {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]ScheduledTask, len(s.pending))
	copy(out, s.pending)
	return out
}

func (s *scheduler) loop() {
	defer close(s.doneCh)
	for {
		select {
		case <-s.stopCh:
			s.timer.Stop()
			return
		case <-s.timer.C:
			due := s.popDue(time.Now())
			for i := range due {
				s.exec.enqueue(due[i])
			}
			s.armTimer()
		}
	}
}

func (s *scheduler) popDue(now time.Time) []ScheduledTask {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for n < len(s.pending) && !s.pending[n].TriggerAt.After(now) {
		n++
	}
	due := append([]ScheduledTask(nil), s.pending[:n]...)
	s.pending = s.pending[n:]
	return due
}

func (s *scheduler) armTimer() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.timer.Stop() {
		select {
		case <-s.timer.C:
		default:
		}
	}
	if len(s.pending) == 0 {
		return
	}
	d := time.Until(s.pending[0].TriggerAt)
	if d < 0 {
		d = 0
	}
	s.timer.Reset(d)
}
