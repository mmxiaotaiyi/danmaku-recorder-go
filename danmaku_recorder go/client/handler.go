package client

import (
	"bytes"
	"compress/zlib"
	"io"

	"dm_client_go/message"
	"dm_client_go/storage"

	"github.com/andybalholm/brotli"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

var IGNORED_MESSAGE_TYPES = map[string]bool{
	"STOP_LIVE_ROOM_LIST":                    true,
	"RECOMMEND_CARD":                         true,
	"GOTO_BUY_FLOW":                          true,
	"GIFT_STAR_PROCESS":                      true,
	"WIDGET_GIFT_STAR_PROCESS":               true,
	"WIDGET_BANNER":                          true,
	"COMMON_NOTICE_DANMAKU":                  true,
	"USER_TOAST_MSG_V2":                      true,
	"LIVE_MULTI_VIEW_NEW_INFO":               true,
	"SHOPPING_BUBBLES_STYLE":                 true,
	"SHOPPING_CART_SHOW":                     true,
	"OTHER_SLICE_LOADING_RESULT":             true,
	"SUPER_CHAT_MESSAGE_DELETE":              true,
	"SUPER_CHAT_ENTRANCE":                    true,
	"GUARD_HONOR_THOUSAND":                   true,
	"POPULARITY_RANK_TAB_CHG":                true,
	"CHG_RANK_REFRESH":                       true,
	"ANCHOR_LOT_NOTICE":                      true,
	"POPULARITY_RED_POCKET_NEW":              true,
	"POPULARITY_RED_POCKET_START":            true,
	"POPULARITY_RED_POCKET_V2_WINNER_LIST":   true,
	"COMMON_ANIMATION":                       true,
	"RING_STATUS_CHANGE":                     true,
	"RING_STATUS_CHANGE_V2":                  true,
	"VOICE_JOIN_ROOM_COUNT_INFO":             true,
	"VOICE_JOIN_LIST":                        true,
	"PLAYURL_RELOAD":                         true,
	"SHOPPING_EXPLAIN_CARD":                  true,
	"POPULAR_RANK_GUIDE_CARD":                true,
	"GUARD_LEADER_NOTICE":                    true,
	"TRADING_SCORE":                          true,
	"master_qn_strategy_chg":                 true,
	"SYS_MSG":                                true,
	"ANCHOR_LOT_CHECKSTATUS":                 true,
	"ANCHOR_LOT_END":                         true,
	"UNIVERSAL_EVENT_GIFT":                   true,
	"VOICE_JOIN_SWITCH":                      true,
	"GUARD_ACHIEVEMENT_ROOM":                 true,
	"ROOM_SKIN_MSG":                          true,
	"PLAYURL_RELOAD_MASTER":                  true,
	"REVENUE_DISPLAY_EFFECT":                 true,
	"WIDGET_GIFT_STAR_PROCESS_V2":            true,
	"WIDGET_WISH_INFO":                       true,
	"LIVE_PANEL_CHANGE_CONTENT":              true,
	"OPENPLATFORM_GAME_BUTTON_STATUS_CHANGE": true,
	"FLOW_REWARD_CARD":                       true,
}

type MsgHandler struct {
	RoomId  int
	storage *storage.StorageService
}

var HeartBeatReplyCmd = "HEARTBEAT_REPLY"

func (msgHandler *MsgHandler) CmdHandler(wsHeader *WsHeader, msg []byte) {
	body := msg[wsHeader.HeaderLen:wsHeader.PackageLen]
	result := gjson.GetBytes(body, "cmd")
	if !result.Exists() || result.Str == "" || IGNORED_MESSAGE_TYPES[result.Str] {
		return
	}

	message.ProcessMessage(result.Str, body, msgHandler.storage, msgHandler.RoomId)
}

func (msgHandler *MsgHandler) CmdBrotliProtoDecoder(wsHeader *WsHeader, msg []byte) []byte {
	reader := brotli.NewReader(bytes.NewReader(msg[wsHeader.HeaderLen:wsHeader.PackageLen]))
	resp, err := io.ReadAll(reader)
	if err != nil {
		return []byte{}
	}
	return resp
}

func (msgHandler *MsgHandler) CmdZlibProtoDecoder(wsHeader *WsHeader, msg []byte) []byte {
	var resp bytes.Buffer
	w := zlib.NewWriter(&resp)
	_, err := w.Write(msg[wsHeader.HeaderLen:wsHeader.PackageLen])
	if err != nil {
		return []byte{}
	}
	err = w.Close()
	if err != nil {
		return nil
	}
	return resp.Bytes()
}

func (msgHandler *MsgHandler) MsgHandler(msg []byte) {
	wsHeader := WsHeaderDecoder(msg)
	switch wsHeader.OpCode {
	case OpHeartBeatReply:
		// wsHeartBeatReply := WsHeartBeatReply{}
		// wsHeartBeatReply.SetPackage(wsHeader, msg)
	case OpCmd:
		msgBody := msg
		cmdHeader := wsHeader
		switch wsHeader.ProtoVer {
		case CmdZlibProto:
			msgBody = msgHandler.CmdZlibProtoDecoder(&wsHeader, msg)
			cmdHeader = WsHeaderDecoder(msgBody)
			for {
				msgHandler.CmdHandler(&cmdHeader, msgBody[:int(cmdHeader.PackageLen)])
				msgBody = msgBody[cmdHeader.PackageLen:]
				if len(msgBody) == 0 {
					break
				}
				cmdHeader = WsHeaderDecoder(msgBody)
			}
		case CmdBrotliProto:
			msgBody = msgHandler.CmdBrotliProtoDecoder(&wsHeader, msg)
			cmdHeader = WsHeaderDecoder(msgBody)
			for {
				msgHandler.CmdHandler(&cmdHeader, msgBody[:int(cmdHeader.PackageLen)])
				msgBody = msgBody[cmdHeader.PackageLen:]
				if len(msgBody) == 0 {
					break
				}
				cmdHeader = WsHeaderDecoder(msgBody)
			}
		default:
			msgHandler.CmdHandler(&cmdHeader, msgBody[:int(cmdHeader.PackageLen)])
		}
	case OpAuthReply:
		// wsAuthReplyMessage := WsAuthReplyMessage{}
		// wsAuthReplyMessage.SetPackage(wsHeader, msg)
	case OpError:
		log.Warnf("recv error msg: %s", string(msg[wsHeader.HeaderLen:wsHeader.PackageLen]))
	}
}
