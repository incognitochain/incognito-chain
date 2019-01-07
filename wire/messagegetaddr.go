package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
)

const (
	MaxGetAddrPayload = 1000 // 1 1Kb
)

type MessageGetAddr struct {
}

func (self *MessageGetAddr) Hash() string {
	return ""
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
