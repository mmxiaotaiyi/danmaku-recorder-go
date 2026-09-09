package message

import (
	"encoding/base64"
	"sync"

	interactpb "dm_client_go/bilibili_interact.pb"
	rankpb "dm_client_go/bilibili_rank.pb"
	sendgiftpb "dm_client_go/bilibili_send_gift.pb"

	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/proto"
)

var (
	pbBufPool    = sync.Pool{New: func() any { b := make([]byte, 0, 4096); return &b }}
	interactPool = sync.Pool{New: func() any { return new(interactpb.InteractMessage) }}
	rankPool     = sync.Pool{New: func() any { return new(rankpb.OnlineRankV3) }}
	giftPool     = sync.Pool{New: func() any { return new(sendgiftpb.SendGift) }}
)

func decodePBBytes(rawData []byte) ([]byte, func()) {
	pb := gjson.GetBytes(rawData, "data.pb")
	src := pb.Raw
	if src[0] == '"' {
		src = src[1 : len(src)-1]
	}
	bufp := pbBufPool.Get().(*[]byte)
	need := base64.StdEncoding.DecodedLen(len(src))
	if cap(*bufp) < need {
		*bufp = make([]byte, need)
	} else {
		*bufp = (*bufp)[:need]
	}
	n, err := base64.StdEncoding.Decode(*bufp, []byte(src))
	if err != nil {
		pbBufPool.Put(bufp)
		return nil, func() {}
	}
	*bufp = (*bufp)[:n]
	return *bufp, func() { pbBufPool.Put(bufp) }
}

func unmarshalInteract(rawData []byte) (*interactpb.InteractMessage, func()) {
	decoded, release := decodePBBytes(rawData)
	msg := interactPool.Get().(*interactpb.InteractMessage)
	msg.Reset()
	if err := proto.Unmarshal(decoded, msg); err != nil {
		interactPool.Put(msg)
		release()
		return nil, func() {}
	}
	return msg, func() {
		msg.Reset()
		interactPool.Put(msg)
		release()
	}
}

func unmarshalRankV3(rawData []byte) (*rankpb.OnlineRankV3, func()) {
	decoded, release := decodePBBytes(rawData)
	msg := rankPool.Get().(*rankpb.OnlineRankV3)
	msg.Reset()
	if err := proto.Unmarshal(decoded, msg); err != nil {
		rankPool.Put(msg)
		release()
		return nil, func() {}
	}
	return msg, func() {
		msg.Reset()
		rankPool.Put(msg)
		release()
	}
}

func unmarshalSendGift(rawData []byte) (*sendgiftpb.SendGift, func()) {
	decoded, release := decodePBBytes(rawData)
	msg := giftPool.Get().(*sendgiftpb.SendGift)
	msg.Reset()
	if err := proto.Unmarshal(decoded, msg); err != nil {
		giftPool.Put(msg)
		release()
		return nil, func() {}
	}
	return msg, func() {
		msg.Reset()
		giftPool.Put(msg)
		release()
	}
}
