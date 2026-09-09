package message

import (
	"encoding/json"
	"fmt"
	"maps"
	"runtime/debug"
	"time"

	stg "dm_client_go/storage"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

var GlobalHandler interface {
	DelRoom(roomId int) error
}

type MessageProcessor func([]byte, *stg.StorageService, int)

func safeProcessor(processor MessageProcessor) MessageProcessor {
	return func(rawData []byte, storage *stg.StorageService, roomId int) {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("message processor panic: room=%d err=%v", roomId, r)
				log.Debugf("message processor stack: room=%d\n%s", roomId, debug.Stack())
			}
		}()

		processor(rawData, storage, roomId)
	}
}

var messageProcessors = map[string]MessageProcessor{}

func init() {
	originalProcessors := map[string]MessageProcessor{
		"DANMU_MSG":                         DanmuMsg,
		"INTERACT_WORD":                     InteractWord,
		"LIKE_INFO_V3_CLICK":                LikeInfoV3Click,
		"LIKE_INFO_V3_UPDATE":               LikeInfoV3Update,
		"ENTRY_EFFECT":                      EntryEffect,
		"WATCHED_CHANGE":                    WatchedChange,
		"ROOM_REAL_TIME_MESSAGE_UPDATE":     RoomRealTimeMessageUpdate,
		"ONLINE_RANK_V2":                    OnlineRankV2,
		"ONLINE_RANK_COUNT":                 OnlineRankCount,
		"DM_INTERACTION":                    DmInteraction,
		"RANK_CHANGED":                      RankChanged,
		"RANK_CHANGED_V2":                   RankChanged,
		"NOTICE_MSG":                        NoticeMsg,
		"REVENUE_RANK_CHANGED":              RevenueRankChanged,
		"SEND_GIFT":                         SendGift,
		"POPULAR_RANK_CHANGED":              PopularRankChanged,
		"COMBO_SEND":                        ComboSend,
		"SUPER_CHAT_MESSAGE":                SuperChatMessage,
		"SUPER_CHAT_MESSAGE_JPN":            SuperChatMessageJpn,
		"GUARD_BUY":                         GuardBuy,
		"USER_TOAST_MSG":                    UserToastMsg,
		"POPULARITY_RED_POCKET_V2_NEW":      PopularityRedPocketV2New,
		"POPULARITY_RED_POCKET_V2_START":    PopularityRedPocketV2Start,
		"POPULARITY_RED_POCKET_WINNER_LIST": PopularityRedPocketWinnerList,
		"ANCHOR_LOT_START":                  AnchorLotStart,
		"DANMU_AGGREGATION":                 DanmuAggregation,
		"ANCHOR_LOT_AWARD":                  AnchorLotAward,
		"ONLINE_RANK_V3":                    OnlineRankV3,
		"INTERACT_WORD_V2":                  InteractWordV2,
		"UNIVERSAL_EVENT_GIFT_V2":           UniversalEventGift,
		"LOG_IN_NOTICE":                     LogInNotice,
		"PREPARING":                         Preparing,
		"HOT_ROOM_NOTIFY":                   HotRoomNotify,
		"RANK_REM":                          RankRem,
		"LIVE":                              Live,
		"INTERACTIVE_USER":                  InteractiveUser,
		"PANEL_INTERACTIVE_NOTIFY_CHANGE":   PanelInteractiveNotifyChange,
		"TIP_CARD":                          TipCard,
		"HOT_BUY_NUM":                       HotBuyNum,
		"DANMU_MSG_MIRROR":                  DanmuMsgMirror,
		"COLLABORATION_LIVE_POPULARITY":     CollaborationLivePopularity,
		"COLLABORATION_LIVE_WATCHED":        CollaborationLiveWatched,
		"COLLABORATION_LIVE_ONLINE":         CollaborationLiveOnline,
		"GIFT_COMBO":                        GiftCombo,
		"SEND_GIFT_V2":                      SendGiftV2,
		"POPULARITY_CHANGE":                 PopularityChange,
		"ROOM_CHANGE":                       RoomChange,
		"RECALL_DANMU_MSG":                  RecallDanmuMsg,
	}

	for cmd, processor := range originalProcessors {
		messageProcessors[cmd] = safeProcessor(processor)
	}
}

func ProcessMessage(cmd string, rawData []byte, storage *stg.StorageService, roomId int) {
	if processor, exists := messageProcessors[cmd]; exists {
		processor(rawData, storage, roomId)
		return
	}
	addRawJSON(storage, "OTHERS", rawData)
}

func DanmuMsg(rawData []byte, storage *stg.StorageService, roomId int) {
	info := gjson.ParseBytes(rawData).Get("info")
	rec := parseDanmu(info)
	storage.LastMsgTs.Store(rec.Ts)
	addJSON(storage, "DANMU_MSG", rec)
}

func InteractWord(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	uinfo := data.Get("uinfo")
	rec := InteractWordRecord{
		Ts:       data.Get("timestamp").Int(),
		Name:     data.Get("uname").Str,
		Relation: int32(data.Get("relation_tail.tail_type").Int()),
		Type:     int32(data.Get("msg_type").Int()),
		WL:       -1,
	}
	uid := data.Get("uid").Int()
	if uid != 0 {
		rec.UID = uid
	}
	if uinfo.Get("wealth").Type != gjson.Null {
		rec.WL = int32(uinfo.Get("wealth.level").Int())
	}
	if uinfo.Get("guard").Type != gjson.Null && uinfo.Get("guard.level").Int() != 0 {
		rec.Guard = json.RawMessage(uinfo.Get("guard").Raw)
	}
	if uinfo.Get("medal").Type != gjson.Null {
		rec.Medal = &Medal{Score: uinfo.Get("medal.score").Int(), Ruid: uinfo.Get("medal.ruid").Int()}
	}
	addJSON(storage, "INTERACT_WORD", rec)
}

func LikeInfoV3Click(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	rec := LikeClickRecord{
		Name: data.Get("uname").Str,
		UID:  data.Get("uid").Int(),
		Ts:   time.Now().Unix(),
	}
	uinfo := data.Get("uinfo")
	if uinfo.Type != gjson.Null {
		if uinfo.Get("medal").Type != gjson.Null {
			rec.Medal = &Medal{Score: uinfo.Get("medal.score").Int(), Ruid: uinfo.Get("medal.ruid").Int()}
		}
		if uinfo.Get("guard").Type != gjson.Null && uinfo.Get("guard.level").Int() != 0 {
			rec.Guard = json.RawMessage(uinfo.Get("guard").Raw)
		}
	}
	addJSON(storage, "LIKE_INFO_V3_CLICK", rec)
}

func LikeInfoV3Update(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	addJSON(storage, "LIKE_INFO_V3_UPDATE", CountTsRecord{Cnt: data.Get("click_count").Int(), Ts: time.Now().Unix()})
}

func EntryEffect(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	uinfo := data.Get("uinfo")
	rec := EntryEffectRecord{
		UID:  data.Get("uid").Int(),
		Name: data.Get("uinfo.base.name").Str,
		Ts:   data.Get("trigger_time").Int() / 1000000000,
		ID:   data.Get("id").Int(),
		WL:   -1,
	}
	if uinfo.Get("medal").Type != gjson.Null {
		rec.Medal = &Medal{Score: uinfo.Get("medal.score").Int(), Ruid: uinfo.Get("medal.ruid").Int()}
	}
	if uinfo.Get("guard").Type != gjson.Null && uinfo.Get("guard.level").Int() != 0 {
		rec.Guard = json.RawMessage(uinfo.Get("guard").Raw)
	}
	if uinfo.Get("wealth").Type != gjson.Null {
		rec.WL = uinfo.Get("wealth.level").Int()
	}
	addJSON(storage, "ENTRY_EFFECT", rec)
}

func WatchedChange(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	addJSON(storage, "WATCHED_CHANGE", CountTsRecord{Cnt: data.Get("num").Int(), Ts: time.Now().Unix()})
}

func RoomRealTimeMessageUpdate(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	addJSON(storage, "ROOM_REAL_TIME_MESSAGE_UPDATE", FansRecord{
		Fans:     data.Get("fans").Int(),
		FansClub: data.Get("fans_club").Int(),
		Ts:       time.Now().Unix(),
	})
}

func OnlineRankV2(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	onlineList := data.Get("online_list")
	if !onlineList.Exists() {
		onlineList = data.Get("list")
	}
	var current []stg.RankItem
	onlineList.ForEach(func(_, user gjson.Result) bool {
		item := stg.RankItem{
			UID:   user.Get("uid").Int(),
			Score: user.Get("score").Int(),
			Name:  user.Get("uname").Str,
			Rank:  int32(user.Get("rank").Int()),
		}
		guard := user.Get("uinfo.guard")
		if guard.Get("level").Int() != 0 {
			item.Guard = &stg.RankGuard{Level: int32(guard.Get("level").Int()), ExpiredStr: guard.Get("expired_str").Str}
		}
		current = append(current, item)
		return true
	})
	applyOnlineRank(storage, current)
}

func OnlineRankCount(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	addJSON(storage, "ONLINE_RANK_COUNT", CountTsRecord{Cnt: data.Get("count").Int(), Ts: time.Now().Unix()})
	if !storage.LiveState {
		now := time.Now().Unix()
		last := storage.LastMsgTs.Load()
		if now-last >= 300 || now-storage.EndTs > 7200 {
			GlobalHandler.DelRoom(roomId)
		}
	}
}

func DmInteraction(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	rec := DmInteractionRecord{
		Type:   data.Get("type").Int(),
		Status: data.Get("status").Int(),
		Ts:     time.Now().Unix(),
	}
	content := gjson.Parse(data.Get("data").Str)
	switch rec.Type {
	case 102:
		content.Get("combo").ForEach(func(_, item gjson.Result) bool {
			rec.Combo = append(rec.Combo, DmCombo{
				Text: item.Get("content").Str,
				Cnt:  item.Get("cnt").Int(),
				Rest: item.Get("left_duration").Int() / 1000,
			})
			return true
		})
	case 103, 105, 106:
		rec.Cnt = content.Get("cnt").Int()
	case 104:
		rec.Cnt = content.Get("cnt").Int()
		rec.GiftID = content.Get("gift_id").Int()
	default:
		rec.Data = data.Get("data").Value()
	}
	addJSON(storage, "DM_INTERACTION", rec)
}

func RankChanged(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	addJSON(storage, "RANK_CHANGED", RankChangedRecord{
		Rank:      data.Get("rank").Int(),
		Ts:        data.Get("timestamp").Int(),
		Type:      fmt.Sprintf("%s-%s-%s", data.Get("on_rank_name_by_type").Str, data.Get("rank_name_by_type").Str, data.Get("rank_type").Str),
		Countdown: data.Get("countdown").Int(),
	})
}

func NoticeMsg(rawData []byte, storage *stg.StorageService, roomId int) {
	jsonResult := gjson.ParseBytes(rawData)
	addJSON(storage, "NOTICE_MSG", NoticeMsgRecord{
		ID:         jsonResult.Get("id").Int(),
		Name:       jsonResult.Get("name").Str,
		RoomID:     jsonResult.Get("roomid").Int(),
		RealRoomID: jsonResult.Get("real_roomid").Int(),
		MsgCommon:  jsonResult.Get("msg_common").Str,
		MsgSelf:    jsonResult.Get("msg_self").Str,
		MsgType:    jsonResult.Get("msg_type").Int(),
		BusinessID: jsonResult.Get("business_id").Str,
	})
}

func RevenueRankChanged(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	addJSON(storage, "REVENUE_RANK_CHANGED", RevenueRankRecord{
		ConfID: data.Get("conf_id").Int(),
		Name:   data.Get("rank_name").Str,
		Rank:   data.Get("rank").Int(),
		Ts:     data.Get("timestamp").Int(),
	})
}

func SendGift(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	sender := data.Get("sender_uinfo")
	rec := SendGiftRecord{
		Name:     sender.Get("base.name").Str,
		UID:      sender.Get("uid").Int(),
		Ts:       data.Get("timestamp").Int(),
		WL:       int32(data.Get("wealth_level").Int()),
		CoinType: coinTypeValue(data.Get("coin_type").Str),
	}
	batchCombo := data.Get("batch_combo_send")
	if batchCombo.Type != gjson.Null {
		rec.ComboSend = ComboSendInfo{
			Num:      int32(batchCombo.Get("batch_combo_num").Int()),
			GiftID:   int32(batchCombo.Get("gift_id").Int()),
			GiftName: batchCombo.Get("gift_name").Str,
			Price:    data.Get("price").Float() / 1000,
			GiftNum:  int32(batchCombo.Get("gift_num").Int()),
		}
		blindGift := batchCombo.Get("blind_gift")
		if blindGift.Type != gjson.Null {
			rec.ComboSend.BlindGift = &BlindGiftInfo{
				ConfigID:          int32(blindGift.Get("blind_gift_config_id").Int()),
				GiftTipPrice:      blindGift.Get("gift_tip_price").Float() / 1000,
				OriginalGiftID:    int32(blindGift.Get("original_gift_id").Int()),
				OriginalGiftName:  blindGift.Get("original_gift_name").Str,
				OriginalGiftPrice: blindGift.Get("original_gift_price").Float() / 1000,
			}
		}
	} else {
		rec.ComboSend = ComboSendInfo{
			GiftID:   int32(data.Get("giftId").Int()),
			GiftName: data.Get("giftName").Str,
			Price:    data.Get("price").Float() / 1000,
			Num:      int32(data.Get("super_gift_num").Int()),
		}
	}

	medal := sender.Get("medal")
	if medal.Type != gjson.Null {
		rec.Medal = &Medal{Score: medal.Get("score").Int(), Ruid: medal.Get("ruid").Int()}
		if g := medal.Get("guard_level").Int(); g != 0 {
			rec.Guard = int32(g)
		}
	}
	addJSON(storage, "SEND_GIFT", rec)
}

func PopularRankChanged(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	addJSON(storage, "RANK_CHANGED", RankChangedRecord{
		Rank:      data.Get("rank_by_type").Int(),
		Ts:        data.Get("timestamp").Int(),
		Type:      fmt.Sprintf("%s-%s", data.Get("on_rank_name_by_type").Str, data.Get("rank_name_by_type").Str),
		Countdown: data.Get("countdown").Int(),
	})
}

func ComboSend(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	sender := data.Get("sender_uinfo")
	rec := ComboSendRecord{
		Cmd:       "COMBO_SEND",
		ComboNum:  data.Get("combo_num").Int(),
		TotalCoin: data.Get("combo_total_coin").Float() / 1000,
		CoinType:  coinTypeValue(data.Get("coin_type").Str),
		GiftID:    data.Get("gift_id").Int(),
		GiftName:  data.Get("gift_name").Str,
		Name:      sender.Get("base.name").Str,
		UID:       sender.Get("uid").Int(),
		Ts:        time.Now().Unix(),
		WL:        data.Get("wealth_level").Int(),
	}
	medal := sender.Get("medal")
	if medal.Type != gjson.Null {
		rec.Medal = &Medal{Score: medal.Get("score").Int(), Ruid: medal.Get("ruid").Int()}
		if g := medal.Get("guard_level").Int(); g != 0 {
			rec.Guard = int32(g)
		}
	}
	addJSON(storage, "SEND_GIFT", rec)
}

func SuperChatMessage(rawData []byte, storage *stg.StorageService, roomId int) {
	jsonResult := gjson.ParseBytes(rawData)
	data := jsonResult.Get("data")
	uinfo := data.Get("uinfo")
	rec := SuperChatRecord{
		Text:        data.Get("message").Str,
		Translation: data.Get("message_trans").Str,
		Price:       data.Get("price").Int(),
		Time:        data.Get("time").Int(),
		UID:         data.Get("uid").Int(),
		Name:        uinfo.Get("base.name").Str,
		UserLevel:   data.Get("user_info.user_level").Int(),
		Ts:          jsonResult.Get("send_time").Int(),
	}
	guard := uinfo.Get("guard")
	if guard.Type != gjson.Null && guard.Get("level").Int() != 0 {
		rec.Guard = json.RawMessage(guard.Raw)
	}
	if uinfo.Get("medal").Type != gjson.Null {
		rec.Medal = &Medal{Ruid: uinfo.Get("medal.ruid").Int(), Score: uinfo.Get("medal.score").Int()}
	}
	addJSON(storage, "SUPER_CHAT_MESSAGE", rec)
}

func SuperChatMessageJpn(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	addJSON(storage, "SUPER_CHAT_MESSAGE", SuperChatJpnRecord{
		Cmd:        "TRANSLATION",
		Message:    data.Get("message").Str,
		MessageJpn: data.Get("message_jpn").Str,
	})
}

func GuardBuy(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	rec := GuardBuyRecord{
		UID:        data.Get("uid").Int(),
		Name:       data.Get("username").Str,
		GuardLevel: data.Get("guard_level").Int(),
		Num:        data.Get("num").Int(),
	}
	startTime := data.Get("start_time").Int()
	endTime := data.Get("end_time").Int()
	if startTime == endTime {
		rec.Ts = startTime
	} else {
		rec.StartTime = startTime
		rec.EndTime = endTime
	}
	addJSON(storage, "GUARD_BUY", rec)
}

func UserToastMsg(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	addJSON(storage, "GUARD_BUY", UserToastRecord{
		TotalGuard: data.Get("target_guard_count").Int(),
		Msg:        data.Get("toast_msg").Str,
		UID:        data.Get("uid").Int(),
		Unit:       data.Get("unit").Str,
		Price:      int(data.Get("price").Float() / 1000),
		Ts:         time.Now().Unix(),
	})
}

func PopularityRedPocketV2New(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	sender := data.Get("sender_info")
	rec := RedPocketNewRecord{
		StartTs: data.Get("start_time").Int(),
		Ts:      data.Get("current_time").Int(),
		Name:    data.Get("uname").Str,
		UID:     data.Get("uid").Int(),
		Price:   data.Get("price").Int() / 10,
		Num:     data.Get("num").Int(),
		WL:      -1,
	}
	if sender.Get("medal").Type != gjson.Null {
		rec.Medal = &Medal{Ruid: sender.Get("medal.ruid").Int(), Score: sender.Get("medal.score").Int()}
	}
	guard := sender.Get("guard")
	if guard.Type != gjson.Null && guard.Get("level").Int() != 0 {
		rec.Guard = json.RawMessage(guard.Raw)
	}
	if sender.Get("wealth").Type != gjson.Null {
		rec.WL = sender.Get("wealth.level").Int()
	}
	addJSON(storage, "POPULARITY_RED_POCKET", rec)
}

func PopularityRedPocketV2Start(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	rec := RedPocketStartRecord{
		UID:        data.Get("sender_uid").Int(),
		Name:       data.Get("sender_name").Str,
		Ts:         data.Get("current_time").Int(),
		StartTs:    data.Get("start_time").Int(),
		Limit:      data.Get("join_requirement").Int(),
		Text:       data.Get("danmu").Str,
		LastTime:   data.Get("last_time").Int(),
		TotalPrice: data.Get("total_price").Int() / 1000,
	}
	data.Get("awards").ForEach(func(_, award gjson.Result) bool {
		rec.Awards = append(rec.Awards, RedPocketAward{
			ID:   award.Get("gift_id").Int(),
			Name: award.Get("gift_name").Str,
			Num:  award.Get("num").Int(),
		})
		return true
	})
	addJSON(storage, "POPULARITY_RED_POCKET", rec)
}

func PopularityRedPocketWinnerList(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	rec := RedPocketWinnerListRecord{
		TotalNum: data.Get("total_num").Int(),
		AwardNum: data.Get("award_num").Int(),
		Ts:       data.Get("timestamp").Int(),
	}
	data.Get("winner_info").ForEach(func(_, winner gjson.Result) bool {
		arr := winner.Array()
		rec.Winners = append(rec.Winners, []any{arr[0].Int(), arr[1].Str, arr[3].Int()})
		return true
	})
	addJSON(storage, "POPULARITY_RED_POCKET", rec)
}

func AnchorLotStart(rawData []byte, storage *stg.StorageService, roomId int) {
	jsonResult := gjson.ParseBytes(rawData)
	data := jsonResult.Get("data")
	rec := AnchorLotStartRecord{
		Text:         data.Get("award_content").Str,
		Name:         data.Get("award_name").Str,
		Num:          data.Get("award_num").Int(),
		Price:        data.Get("award_price_text").Value(),
		Type:         data.Get("award_type").Int(),
		CurrentTime:  data.Get("current_time").Int(),
		Danmu:        data.Get("danmu").Str,
		ID:           data.Get("id").Int(),
		Duration:     data.Get("max_time").Int(),
		RequireText:  data.Get("require_text").Str,
		RequireType:  data.Get("require_type").Int(),
		RequireValue: data.Get("require_value").Int(),
		Ts:           jsonResult.Get("send_time").Int(),
	}
	if giftID := data.Get("gift_id").Int(); giftID != 0 {
		rec.GiftID = giftID
		rec.GiftName = data.Get("gift_name").Str
		rec.GiftNum = data.Get("gift_num").Int()
		rec.GiftPrice = data.Get("gift_price").Float() / 1000
	}
	addJSON(storage, "DANMU_AGGREGATION", rec)
}

func DanmuAggregation(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	addJSON(storage, "DANMU_AGGREGATION", DanmuAggregationRecord{
		ID:  data.Get("activity_identity").Str,
		Msg: data.Get("msg").Str,
		Num: data.Get("aggregation_num").Int(),
		Ts:  data.Get("timestamp").Int(),
	})
}

func AnchorLotAward(rawData []byte, storage *stg.StorageService, roomId int) {
	jsonResult := gjson.ParseBytes(rawData)
	data := jsonResult.Get("data")
	rec := AnchorLotAwardRecord{
		ID:        data.Get("id").Int(),
		AwardName: data.Get("award_name").Str,
		AwardNum:  data.Get("award_num").Int(),
		Ts:        jsonResult.Get("send_time").Int(),
	}
	data.Get("award_users").ForEach(func(_, user gjson.Result) bool {
		rec.AwardUsers = append(rec.AwardUsers, AwardUser{
			Num:  user.Get("num").Int(),
			UID:  user.Get("uid").Int(),
			Name: user.Get("uname").Str,
		})
		return true
	})
	addJSON(storage, "DANMU_AGGREGATION", rec)
}

func OnlineRankV3(rawData []byte, storage *stg.StorageService, roomId int) {
	onlineRank, release := unmarshalRankV3(rawData)
	defer release()

	current := make([]stg.RankItem, 0, len(onlineRank.OnlineList))
	for _, user := range onlineRank.OnlineList {
		item := stg.RankItem{
			UID:   user.Uid,
			Name:  user.Uname,
			Rank:  user.Rank,
			Score: parseScoreInt(user.Score),
		}
		if user.Uinfo.Guard.Level != 0 {
			item.Guard = &stg.RankGuard{
				Level:      user.Uinfo.Guard.Level,
				ExpiredStr: user.Uinfo.Guard.ExpiredStr,
			}
		}
		current = append(current, item)
	}
	applyOnlineRank(storage, current)
}

func InteractWordV2(rawData []byte, storage *stg.StorageService, roomId int) {
	interactWord, release := unmarshalInteract(rawData)
	defer release()

	var relation int32
	if interactWord.RelationTail != nil {
		relation = interactWord.RelationTail.TailType
	}
	rec := InteractWordRecord{
		Ts:       interactWord.Timestamp,
		Name:     interactWord.Uname,
		Relation: relation,
		Type:     interactWord.MsgType,
		WL:       -1,
	}
	uid := interactWord.Uid
	if uid != 0 {
		rec.UID = uid
	}
	if uinfo := interactWord.Uinfo; uinfo != nil {
		if uinfo.Wealth != nil {
			rec.WL = uinfo.Wealth.Level
		}
		if uinfo.Guard != nil && uinfo.Guard.Level != 0 {
			rec.Guard = guardRaw(uinfo.Guard.Level, uinfo.Guard.ExpiredStr)
		}
		if uinfo.Medal != nil {
			rec.Medal = &Medal{Score: int64(uinfo.Medal.Score), Ruid: uinfo.Medal.Ruid}
		}
	}
	addJSON(storage, "INTERACT_WORD", rec)
}

func LogInNotice(rawData []byte, storage *stg.StorageService, roomId int) {
	log.Infof("roomId: %d 弹幕接收正常", roomId)
}

func Preparing(rawData []byte, storage *stg.StorageService, roomId int) {
	log.Infof("roomId: %d 直播结束", roomId)
	storage.LiveState = false
	currentTs := time.Now().Unix()
	storage.EndTs = currentTs
	addJSON(storage, "OTHERS", TsRecord{Ts: currentTs})
}

func UniversalEventGift(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	currentUser := &stg.UserData{
		Owner:          data.Get("room_owner").Int(),
		RoomStartTs:    data.Get("room_start_at").Str,
		SessionStartTs: data.Get("session_start_at").Str,
		Members:        make(map[string]map[string]any),
	}
	rec := UniversalEventGiftLine{
		Ts:      data.Get("system_time_unix").Int(),
		Members: [][2]int64{},
	}
	data.Get("members").ForEach(func(_, member gjson.Result) bool {
		uid := member.Get("uid").Int()
		price := member.Get("biz_extra_data.multi_conn.price").Int() / 100
		rec.Members = append(rec.Members, [2]int64{uid, price})
		currentUser.Members[member.Get("uname").Str] = map[string]any{
			"uid":     uid,
			"join_ts": member.Get("join_time").Int(),
		}
		return true
	})
	if len(storage.UserData.Members) == 0 {
		storage.UserData = currentUser
	} else if storage.UserData.Owner == currentUser.Owner &&
		storage.UserData.RoomStartTs == currentUser.RoomStartTs &&
		storage.UserData.SessionStartTs == currentUser.SessionStartTs {
		maps.Copy(storage.UserData.Members, currentUser.Members)
	} else {
		addJSON(storage, "UNIVERSAL_EVENT_GIFT", storage.UserData)
		storage.UserData = currentUser
	}
	addJSON(storage, "UNIVERSAL_EVENT_GIFT", rec)
}

func HotRoomNotify(rawData []byte, storage *stg.StorageService, roomId int) {
	addJSON(storage, "HOT_ROOM_NOTIFY", TsRecord{Ts: time.Now().Unix()})
}

func RankRem(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	addJSON(storage, "RANK_REM", RankRemRecord{
		Name: data.Get("name").Str,
		Ts:   data.Get("time").Int(),
		UID:  data.Get("uid").Int(),
	})
}

func Live(rawData []byte, storage *stg.StorageService, roomId int) {
	storage.LiveState = true
	addRawJSON(storage, "OTHERS", rawData)
}

func InteractiveUser(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	value := data.Get("value")
	addJSON(storage, "INTERACTIVE_USER", InteractiveUserRecord{
		Cmd:               "INTERACTIVE_USER",
		Ts:                time.Now().Unix(),
		Type:              data.Get("type").Int(),
		CountDownTime:     value.Get("count_down_time").Int(),
		Delay:             value.Get("delay").Int(),
		DmMsg:             value.Get("dm_msg").Str,
		ProphetStatus:     value.Get("prophet_status").Int(),
		ProphetUserStatus: value.Get("prophet_user_status").Int(),
		SendMsg:           value.Get("send_msg").Int(),
	})
}

func PanelInteractiveNotifyChange(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	addJSON(storage, "INTERACTIVE_USER", PanelNotifyRecord{
		Cmd:      "PANEL_INTERACTIVE_NOTIFY_CHANGE",
		Ts:       time.Now().Unix(),
		BizID:    data.Get("biz_id").Int(),
		EndTime:  data.Get("end_time").Int(),
		LastTime: data.Get("last_time").Int(),
		Level:    data.Get("level").Int(),
		Text:     data.Get("text").Str,
	})
}

func TipCard(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	addJSON(storage, "HOT_BUY", TipCardRecord{
		Ts:    time.Now().Unix(),
		BizID: data.Get("biz_id").Int(),
		Extra: data.Get("extra").Str,
	})
}

func HotBuyNum(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	addJSON(storage, "HOT_BUY", HotBuyNumRecord{
		Ts:      time.Now().Unix(),
		GoodsID: data.Get("goods_id").Str,
		Num:     data.Get("num").Int(),
	})
}

func DanmuMsgMirror(rawData []byte, storage *stg.StorageService, roomId int) {
	info := gjson.ParseBytes(rawData).Get("info")
	rec := parseDanmu(info)
	storage.LastMsgTs.Store(rec.Ts)
	addJSON(storage, "DANMU_MSG_MIRROR", rec)
}

func CollaborationLivePopularity(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	addJSON(storage, "COLLABORATION_LIVE_POPULARITY", CountTsRecord{Cnt: data.Get("num").Int(), Ts: time.Now().Unix()})
}

func CollaborationLiveWatched(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	addJSON(storage, "COLLABORATION_LIVE_WATCHED", CountTsRecord{Cnt: data.Get("num").Int(), Ts: time.Now().Unix()})
}

func CollaborationLiveOnline(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	addJSON(storage, "COLLABORATION_LIVE_ONLINE", CountTsRecord{Cnt: data.Get("num").Int(), Ts: time.Now().Unix()})
}

func GiftCombo(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	addJSON(storage, "SEND_GIFT", GiftComboRecord{
		Cmd:        "GIFT_COMBO",
		GiftID:     data.Get("giftId").Int(),
		GiftName:   data.Get("giftName").Str,
		Price:      data.Get("price").Float() / 1000,
		Num:        data.Get("num").Int(),
		BatchNum:   data.Get("super_batch_gift_num").Int(),
		TotalPrice: data.Get("combo_total_coin").Float() / 1000,
		UID:        data.Get("uid").Int(),
		Name:       data.Get("uname").Str,
		GuardLevel: data.Get("guard_level").Int(),
		Ts:         data.Get("timestamp").Int(),
	})
}

func SendGiftV2(rawData []byte, storage *stg.StorageService, roomId int) {
	sendGift, release := unmarshalSendGift(rawData)
	defer release()

	combo := sendGift.GetComboSend()
	rec := SendGiftRecord{
		UID:      sendGift.Uid,
		Name:     sendGift.Uname,
		Ts:       combo.Timestamp,
		WL:       sendGift.GetWealth().GetWealthLevel(),
		CoinType: coinTypeValue(combo.CoinType),
		ComboSend: ComboSendInfo{
			GiftID:   combo.GiftId,
			GiftName: combo.GiftName,
			GiftNum:  combo.GiftNum,
			Price:    float64(combo.DiscountPrice) / 1000,
			Num:      combo.SuperGiftNum,
		},
	}
	if bg := sendGift.GetBlindGift(); bg != nil {
		rec.ComboSend.BlindGift = &BlindGiftInfo{
			ConfigID:          bg.BlindGiftConfigId,
			GiftTipPrice:      float64(bg.GiftTipPrice) / 1000,
			OriginalGiftID:    bg.OriginalGiftId,
			OriginalGiftName:  bg.OriginalGiftName,
			OriginalGiftPrice: float64(combo.Price) / 1000,
		}
	}
	if medal := sendGift.GetSenderUinfo().GetMedal(); medal != nil {
		rec.Medal = &Medal{Score: int64(medal.Score), Ruid: medal.Ruid}
		if medal.GuardLevel != 0 {
			rec.Guard = medal.GuardLevel
		}
	}
	addJSON(storage, "SEND_GIFT", rec)
}

func PopularityChange(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	addJSON(storage, "POPULARITY_CHANGE", PopularityRecord{
		Popularity: data.Get("popularity").Int(),
		Ts:         time.Now().Unix(),
	})
}

func RoomChange(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	addJSON(storage, "OTHERS", RoomChangeRecord{
		Cmd:            "ROOM_CHANGE",
		Ts:             time.Now().Unix(),
		AreaID:         data.Get("area_id").Int(),
		AreaParentID:   data.Get("parent_area_id").Int(),
		AreaName:       data.Get("area_name").Str,
		AreaParentName: data.Get("parent_area_name").Str,
		Title:          data.Get("title").Str,
	})
}

func RecallDanmuMsg(rawData []byte, storage *stg.StorageService, roomId int) {
	data := gjson.ParseBytes(rawData).Get("data")
	addJSON(storage, "OTHERS", RecallDanmuRecord{
		Cmd:  "RECALL_DANMU_MSG",
		Ts:   time.Now().Unix(),
		Type: data.Get("recall_type").Int(),
		UID:  data.Get("target_id").Int(),
	})
}
