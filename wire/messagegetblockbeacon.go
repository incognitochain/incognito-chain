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

func (self *MessageGetBlockBeacon) Hash() string {
	rawBytes, err := self.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (self *MessageGetBlockBeacon) MessageType() string {
	return CmdGetBlockBeacon
}

func (self *MessageGetBlockBeacon) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self *MessageGetBlockBeacon) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self *MessageGetBlockBeacon) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self *MessageGetBlockBeacon) SetSenderID(senderID peer.ID) error {
	self.SenderID = senderID.Pretty()
	return nil
}

func (self *MessageGetBlockBeacon) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (self *MessageGetBlockBeacon) VerifyMsgSanity() error {
	return nil
}
