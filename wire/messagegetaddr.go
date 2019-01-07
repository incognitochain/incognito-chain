package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
	"time"
)

const (
	MaxGetAddrPayload = 1000 // 1 1Kb
)

type MessageGetAddr struct {
	Timestamp time.Time
}

func (self *MessageGetAddr) Hash() string {
	rawBytes, err := self.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (self *MessageGetAddr) MessageType() string {
	return CmdGetAddr
}

func (self *MessageGetAddr) MaxPayloadLength(pver int) int {
	return MaxGetAddrPayload
}

func (self *MessageGetAddr) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self *MessageGetAddr) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self *MessageGetAddr) SetSenderID(senderID peer.ID) error {
	return nil
}

func (self *MessageGetAddr) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (self *MessageGetAddr) VerifyMsgSanity() error {
	return nil
}
