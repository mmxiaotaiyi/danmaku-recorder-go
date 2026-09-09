package timer

import "time"

const (
	configFileName  = "schedule_config.json"
	addRoomInterval = time.Second
	persistDebounce = time.Second
)

type RoomConfig struct {
	RoomID int     `json:"room_id"`
	Ts     []int64 `json:"ts"`
}

type ScheduledTask struct {
	RoomID    int
	TriggerAt time.Time
}
