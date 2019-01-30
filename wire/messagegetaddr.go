package wire

import (
	"encoding/json"

	"time"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
)

const (
	MaxGetAddrPayload = 1000 // 1 1Kb
)

type MessageGetAddr struct {
	Timestamp time.Time
}

func (msg *MessageGetAddr) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageGetAddr) MessageType() string {
	return CmdGetAddr
}

func (msg *MessageGetAddr) MaxPayloadLength(pver int) int {
	return MaxGetAddrPayload
}

func (msg *MessageGetAddr) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageGetAddr) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *MessageGetAddr) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *MessageGetAddr) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (msg *MessageGetAddr) VerifyMsgSanity() error {
	return nil
}
