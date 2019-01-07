package wire

import (
	"encoding/json"

	"time"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
)

const (
	MaxShardStatePayload = 4000000 // 4 Mb
)

type MessageShardState struct {
	Timestamp time.Time
	ChainInfo interface{}
	SenderID  string
}

func (self *MessageShardState) Hash() string {
	rawBytes, err := self.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (self *MessageShardState) MessageType() string {
	return CmdShardState
}

func (self *MessageShardState) MaxPayloadLength(pver int) int {
	return MaxShardStatePayload
}

func (self *MessageShardState) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self *MessageShardState) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self *MessageShardState) SetSenderID(senderID peer.ID) error {
	self.SenderID = senderID.Pretty()
	return nil
}

func (self *MessageShardState) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (self *MessageShardState) VerifyMsgSanity() error {
	return nil
}
