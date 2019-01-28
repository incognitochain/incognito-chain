package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
)

const (
	MaxGetBlockPayload = 1000 // 1kb
)

type MessageGetBlockBeacon struct {
	From      uint64
	To        uint64
	SenderID  string
	Timestamp int64
}

func (msg *MessageGetBlockBeacon) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageGetBlockBeacon) MessageType() string {
	return CmdGetBlockBeacon
}

func (msg *MessageGetBlockBeacon) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (msg *MessageGetBlockBeacon) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageGetBlockBeacon) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *MessageGetBlockBeacon) SetSenderID(senderID peer.ID) error {
	msg.SenderID = senderID.Pretty()
	return nil
}

func (msg *MessageGetBlockBeacon) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (msg *MessageGetBlockBeacon) VerifyMsgSanity() error {
	return nil
}
