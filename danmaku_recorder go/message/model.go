package message

import "encoding/json"

type Medal struct {
	Score int64 `json:"score"`
	Ruid  int64 `json:"ruid"`
}

type GuardInfo struct {
	Level      int32  `json:"level,omitempty"`
	ExpiredStr string `json:"expired_str,omitempty"`
}

type BlindGiftInfo struct {
	ConfigID          int32   `json:"config_id"`
	GiftTipPrice      float64 `json:"gift_tip_price"`
	OriginalGiftID    int32   `json:"original_gift_id"`
	OriginalGiftName  string  `json:"original_gift_name"`
	OriginalGiftPrice float64 `json:"original_gift_price"`
}

type ComboSendInfo struct {
	Num       int32          `json:"num,omitempty"`
	GiftID    int32          `json:"gift_id"`
	GiftName  string         `json:"gift_name"`
	Price     float64        `json:"price"`
	GiftNum   int32          `json:"gift_num,omitempty"`
	BlindGift *BlindGiftInfo `json:"blind_gift,omitempty"`
}

type DanmuMsgRecord struct {
	Text       string `json:"text"`
	Name       string `json:"name"`
	Ts         int64  `json:"ts"`
	MidHash    string `json:"midhash"`
	UID        int64  `json:"uid,omitempty"`
	WL         int64  `json:"wl"`
	GuardLevel int64  `json:"guard_level,omitempty"`
	Medal      *Medal `json:"medal,omitempty"`
}

type InteractWordRecord struct {
	Ts       int64           `json:"ts"`
	Name     string          `json:"name"`
	Relation int32           `json:"relation"`
	Type     int32           `json:"type"`
	UID      int64           `json:"uid,omitempty"`
	WL       int32           `json:"wl"`
	Guard    json.RawMessage `json:"guard,omitempty"`
	Medal    *Medal          `json:"medal,omitempty"`
}

type LikeClickRecord struct {
	Name  string          `json:"name"`
	UID   int64           `json:"uid"`
	Ts    int64           `json:"ts"`
	Medal *Medal          `json:"medal,omitempty"`
	Guard json.RawMessage `json:"guard,omitempty"`
}

type CountTsRecord struct {
	Cnt int64 `json:"cnt"`
	Ts  int64 `json:"ts"`
}

type FansRecord struct {
	Fans     int64 `json:"fans"`
	FansClub int64 `json:"fans_club"`
	Ts       int64 `json:"ts"`
}

type EntryEffectRecord struct {
	UID   int64           `json:"uid"`
	Name  string          `json:"name"`
	Ts    int64           `json:"ts"`
	ID    int64           `json:"id"`
	Medal *Medal          `json:"medal,omitempty"`
	Guard json.RawMessage `json:"guard,omitempty"`
	WL    int64           `json:"wl"`
}

type DmCombo struct {
	Text string `json:"text"`
	Cnt  int64  `json:"cnt"`
	Rest int64  `json:"rest"`
}

type DmInteractionRecord struct {
	Type   int64     `json:"type"`
	Status int64     `json:"status"`
	Ts     int64     `json:"ts"`
	Combo  []DmCombo `json:"combo,omitempty"`
	Cnt    int64     `json:"cnt,omitempty"`
	GiftID int64     `json:"gift_id,omitempty"`
	Data   any       `          json:"data,omitempty"`
}

type RankChangedRecord struct {
	Rank      int64  `json:"rank"`
	Ts        int64  `json:"ts"`
	Type      string `json:"type"`
	Countdown int64  `json:"countdown"`
}

type NoticeMsgRecord struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	RoomID     int64  `json:"roomid"`
	RealRoomID int64  `json:"real_roomid"`
	MsgCommon  string `json:"msg_common"`
	MsgSelf    string `json:"msg_self"`
	MsgType    int64  `json:"msg_type"`
	BusinessID string `json:"business_id"`
}

type RevenueRankRecord struct {
	ConfID int64  `json:"conf_id"`
	Name   string `json:"name"`
	Rank   int64  `json:"rank"`
	Ts     int64  `json:"ts"`
}

type SendGiftRecord struct {
	Name      string        `json:"name"`
	UID       int64         `json:"uid"`
	Ts        int64         `json:"ts"`
	WL        int32         `json:"wl"`
	ComboSend ComboSendInfo `json:"combo_send"`
	CoinType  int64         `json:"coin_type"`
	Medal     *Medal        `json:"medal,omitempty"`
	Guard     int32         `json:"guard,omitempty"`
}

type ComboSendRecord struct {
	Cmd       string  `json:"cmd"`
	ComboNum  int64   `json:"combo_num"`
	TotalCoin float64 `json:"total_coin"`
	CoinType  int64   `json:"coin_type"`
	GiftID    int64   `json:"gift_id"`
	GiftName  string  `json:"gift_name"`
	Name      string  `json:"name"`
	UID       int64   `json:"uid"`
	Ts        int64   `json:"ts"`
	WL        int64   `json:"wl"`
	Medal     *Medal  `json:"medal,omitempty"`
	Guard     int32   `json:"guard,omitempty"`
}

type SuperChatRecord struct {
	Text        string          `json:"text"`
	Translation string          `json:"translation"`
	Price       int64           `json:"price"`
	Time        int64           `json:"time"`
	UID         int64           `json:"uid"`
	Name        string          `json:"name"`
	UserLevel   int64           `json:"user_level"`
	Ts          int64           `json:"ts"`
	Guard       json.RawMessage `json:"guard,omitempty"`
	Medal       *Medal          `json:"medal,omitempty"`
}

type SuperChatJpnRecord struct {
	Cmd        string `json:"cmd"`
	Message    string `json:"message"`
	MessageJpn string `json:"message_jpn"`
}

type GuardBuyRecord struct {
	UID        int64  `json:"uid"`
	Name       string `json:"name"`
	GuardLevel int64  `json:"guard_level"`
	Num        int64  `json:"num"`
	Ts         int64  `json:"ts,omitempty"`
	StartTime  int64  `json:"start_time,omitempty"`
	EndTime    int64  `json:"end_time,omitempty"`
}

type UserToastRecord struct {
	TotalGuard int64  `json:"total_guard"`
	Msg        string `json:"msg"`
	UID        int64  `json:"uid"`
	Unit       string `json:"unit"`
	Price      int    `json:"price"`
	Ts         int64  `json:"ts"`
}

type RedPocketNewRecord struct {
	StartTs int64           `json:"start_ts"`
	Ts      int64           `json:"ts"`
	Name    string          `json:"name"`
	UID     int64           `json:"uid"`
	Price   int64           `json:"price"`
	Num     int64           `json:"num"`
	Medal   *Medal          `json:"medal,omitempty"`
	Guard   json.RawMessage `json:"guard,omitempty"`
	WL      int64           `json:"wl"`
}

type RedPocketAward struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Num  int64  `json:"num"`
}

type RedPocketStartRecord struct {
	UID        int64            `json:"uid"`
	Name       string           `json:"name"`
	Ts         int64            `json:"ts"`
	StartTs    int64            `json:"start_ts"`
	Limit      int64            `json:"limit"`
	Text       string           `json:"text"`
	LastTime   int64            `json:"last_time"`
	TotalPrice int64            `json:"total_price"`
	Awards     []RedPocketAward `json:"awards,omitempty"`
}

type RedPocketWinnerListRecord struct {
	TotalNum int64   `json:"total_num"`
	AwardNum int64   `json:"award_num"`
	Ts       int64   `json:"ts"`
	Winners  [][]any `json:"winners,omitempty"`
}

type AnchorLotStartRecord struct {
	Text         string  `json:"text"`
	Name         string  `json:"name"`
	Num          int64   `json:"num"`
	Price        any     `         json:"price"`
	Type         int64   `json:"type"`
	CurrentTime  int64   `json:"current_time"`
	Danmu        string  `json:"danmu"`
	ID           int64   `json:"id"`
	Duration     int64   `json:"duration"`
	RequireText  string  `json:"require_text"`
	RequireType  int64   `json:"require_type"`
	RequireValue int64   `json:"require_value"`
	Ts           int64   `json:"ts"`
	GiftID       int64   `json:"gift_id,omitempty"`
	GiftName     string  `json:"gift_name,omitempty"`
	GiftNum      int64   `json:"gift_num,omitempty"`
	GiftPrice    float64 `json:"gift_price,omitempty"`
}

type DanmuAggregationRecord struct {
	ID  string `json:"id"`
	Msg string `json:"msg"`
	Num int64  `json:"num"`
	Ts  int64  `json:"ts"`
}

type AwardUser struct {
	Num  int64  `json:"num"`
	UID  int64  `json:"uid"`
	Name string `json:"name"`
}

type AnchorLotAwardRecord struct {
	ID         int64       `json:"id"`
	AwardName  string      `json:"award_name"`
	AwardNum   int64       `json:"award_num"`
	Ts         int64       `json:"ts"`
	AwardUsers []AwardUser `json:"award_users,omitempty"`
}

type UniversalEventGiftLine struct {
	Ts      int64      `json:"ts"`
	Members [][2]int64 `json:"members"`
}

type TsRecord struct {
	Ts int64 `json:"ts"`
}

type RankRemRecord struct {
	Name string `json:"name"`
	Ts   int64  `json:"ts"`
	UID  int64  `json:"uid"`
}

type InteractiveUserRecord struct {
	Cmd               string `json:"cmd"`
	Ts                int64  `json:"ts"`
	Type              int64  `json:"type"`
	CountDownTime     int64  `json:"count_down_time"`
	Delay             int64  `json:"delay"`
	DmMsg             string `json:"dm_msg"`
	ProphetStatus     int64  `json:"prophet_status"`
	ProphetUserStatus int64  `json:"prophet_user_status"`
	SendMsg           int64  `json:"send_msg"`
}

type PanelNotifyRecord struct {
	Cmd      string `json:"cmd"`
	Ts       int64  `json:"ts"`
	BizID    int64  `json:"biz_id"`
	EndTime  int64  `json:"end_time"`
	LastTime int64  `json:"last_time"`
	Level    int64  `json:"level"`
	Text     string `json:"text"`
}

type TipCardRecord struct {
	Ts    int64  `json:"ts"`
	BizID int64  `json:"biz_id"`
	Extra string `json:"extra"`
}

type HotBuyNumRecord struct {
	Ts      int64  `json:"ts"`
	GoodsID string `json:"goods_id"`
	Num     int64  `json:"num"`
}

type GiftComboRecord struct {
	Cmd        string  `json:"cmd"`
	GiftID     int64   `json:"gift_id"`
	GiftName   string  `json:"gift_name"`
	Price      float64 `json:"price"`
	Num        int64   `json:"num"`
	BatchNum   int64   `json:"batch_num"`
	TotalPrice float64 `json:"total_price"`
	UID        int64   `json:"uid"`
	Name       string  `json:"name"`
	GuardLevel int64   `json:"guard_level"`
	Ts         int64   `json:"ts"`
}

type PopularityRecord struct {
	Popularity int64 `json:"popularity"`
	Ts         int64 `json:"ts"`
}

type RoomChangeRecord struct {
	Cmd            string `json:"cmd"`
	Ts             int64  `json:"ts"`
	AreaID         int64  `json:"area_id"`
	AreaParentID   int64  `json:"area_parent_id"`
	AreaName       string `json:"area_name"`
	AreaParentName string `json:"area_parent_name"`
	Title          string `json:"title"`
}

type RecallDanmuRecord struct {
	Cmd  string `json:"cmd"`
	Ts   int64  `json:"ts"`
	Type int64  `json:"type"`
	UID  int64  `json:"uid"`
}
