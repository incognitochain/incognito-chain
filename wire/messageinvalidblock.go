package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/common"
)

const (
	MaxInvalidBlockPayload = 1000 // 1 kb
)

type MessageInvalidBlock struct {
	Reason       string //the reason it's invalid could be in
	BlockHash    string
	shardID      byte
	Validator    string
	ValidatorSig string
}

func (self *MessageInvalidBlock) Hash() string {
	rawBytes, err := self.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (self *MessageInvalidBlock) MessageType() string {
	return CmdInvalidBlock
}

func (self *MessageInvalidBlock) MaxPayloadLength(pver int) int {
	return MaxInvalidBlockPayload
}

func (self *MessageInvalidBlock) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self *MessageInvalidBlock) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self *MessageInvalidBlock) SetSenderID(senderID peer.ID) error {
	return nil
}
