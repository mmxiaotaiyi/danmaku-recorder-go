package message

import (
	"encoding/json"
	"strconv"
	"time"

	stg "dm_client_go/storage"

	"github.com/bytedance/sonic"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func addJSON(ss *stg.StorageService, cmd string, v any) {
	data, err := sonic.Marshal(v)
	if err != nil {
		log.Errorf("storage marshal error: cmd=%s err=%v", cmd, err)
		return
	}
	ss.AddLine(cmd, data)
}

func addRawJSON(ss *stg.StorageService, cmd string, data []byte) {
	if len(data) == 0 {
		return
	}
	line := make([]byte, len(data))
	copy(line, data)
	ss.AddLine(cmd, line)
}

func coinTypeValue(s string) int64 {
	switch s {
	case "gold":
		return 1
	default:
		return 0
	}
}

func rankEqual(a, b []stg.RankItem) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].UID != b[i].UID || a[i].Score != b[i].Score || a[i].Name != b[i].Name || a[i].Rank != b[i].Rank {
			return false
		}
		ag, bg := a[i].Guard, b[i].Guard
		if ag == nil && bg == nil {
			continue
		}
		if ag == nil || bg == nil || ag.Level != bg.Level || ag.ExpiredStr != bg.ExpiredStr {
			return false
		}
	}
	return true
}

func applyOnlineRank(storage *stg.StorageService, current []stg.RankItem) {
	currentTs := time.Now().Unix()
	rankData := storage.RankData
	if rankData.Rank == nil {
		storage.RankData = &stg.RankData{Rank: current, Ts: []int64{currentTs}}
		return
	}
	if rankEqual(rankData.Rank, current) {
		rankData.Ts = append(rankData.Ts, currentTs)
		return
	}
	if len(rankData.Ts) > 2 {
		rankData.Ts = []int64{rankData.Ts[0], rankData.Ts[len(rankData.Ts)-1]}
	}
	addJSON(storage, "ONLINE_RANK", rankData)
	storage.RankData = &stg.RankData{Rank: current, Ts: []int64{currentTs}}
}

func parseDanmu(info gjson.Result) DanmuMsgRecord {
	rec := DanmuMsgRecord{
		Text:    info.Get("1").Str,
		Name:    info.Get("2.1").Str,
		Ts:      info.Get("0.4").Int() / 1000,
		MidHash: info.Get("0.7").Str,
		WL:      -1,
	}
	uid := info.Get("2.0").Int()
	if uid != 0 {
		rec.UID = uid
		wl := info.Get("16")
		if wl.Type != gjson.Null {
			rec.WL = wl.Get("0").Int()
		}
	} else {
		wealth := info.Get("0.15.user.wealth")
		if wealth.Type != gjson.Null {
			rec.WL = wealth.Get("level").Int()
		}
	}
	medal := info.Get("0.15.user.medal")
	if medal.Type != gjson.Null {
		rec.Medal = &Medal{Score: medal.Get("score").Int(), Ruid: medal.Get("ruid").Int()}
		if g := medal.Get("guard_level").Int(); g != 0 {
			rec.GuardLevel = g
		}
	}
	return rec
}

func parseScoreInt(s string) int64 {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

func guardRaw(level int32, expired string) json.RawMessage {
	b, err := sonic.Marshal(GuardInfo{Level: level, ExpiredStr: expired})
	if err != nil {
		return nil
	}
	return b
}
