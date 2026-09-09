package timer

import (
	"os"
	"sort"
	"time"

	"github.com/bytedance/sonic"
)

func loadTasks() ([]ScheduledTask, error) {
	data, err := os.ReadFile(configFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}

	var rooms []RoomConfig
	if err := sonic.Unmarshal(data, &rooms); err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	seen := make(map[int]map[int64]struct{}, len(rooms))
	var tasks []ScheduledTask
	for _, room := range rooms {
		got := seen[room.RoomID]
		if got == nil {
			got = make(map[int64]struct{})
			seen[room.RoomID] = got
		}
		for _, ts := range room.Ts {
			if ts <= now {
				continue
			}
			if _, ok := got[ts]; ok {
				continue
			}
			got[ts] = struct{}{}
			tasks = append(tasks, ScheduledTask{
				RoomID:    room.RoomID,
				TriggerAt: time.Unix(ts, 0),
			})
		}
	}
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].TriggerAt.Equal(tasks[j].TriggerAt) {
			return tasks[i].RoomID < tasks[j].RoomID
		}
		return tasks[i].TriggerAt.Before(tasks[j].TriggerAt)
	})
	return tasks, nil
}

func saveTasks(tasks []ScheduledTask) error {
	grouped := make(map[int][]int64, len(tasks))
	roomIDs := make([]int, 0, len(tasks))
	for _, t := range tasks {
		ts := t.TriggerAt.Unix()
		if _, ok := grouped[t.RoomID]; !ok {
			roomIDs = append(roomIDs, t.RoomID)
		}
		grouped[t.RoomID] = append(grouped[t.RoomID], ts)
	}
	sort.Ints(roomIDs)

	rooms := make([]RoomConfig, 0, len(roomIDs))
	for _, id := range roomIDs {
		ts := grouped[id]
		sort.Slice(ts, func(i, j int) bool { return ts[i] < ts[j] })
		rooms = append(rooms, RoomConfig{RoomID: id, Ts: ts})
	}

	data, err := sonic.MarshalIndent(rooms, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(configFileName, data, 0644)
}
