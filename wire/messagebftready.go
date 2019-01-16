package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
)

const (
	MaxBFTReadyPayload = 5000 // 5 Kb
)

type MessageBFTReady struct {
	Timestamp int64
}

func (self *MessageBFTReady) Hash() string {
	rawBytes, err := self.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (self *MessageBFTReady) MessageType() string {
	return CmdBFTReady
}

func (self *MessageBFTReady) MaxPayloadLength(pver int) int {
	return MaxBFTReadyPayload
}

func (self *MessageBFTReady) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self *MessageBFTReady) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self *MessageBFTReady) SetSenderID(senderID peer.ID) error {
	return nil
}

func (self *MessageBFTReady) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (self *MessageBFTReady) VerifyMsgSanity() error {
	return nil
}
