package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/cashec"
)

const (
	MaxBlockPayload = 1000000 // 1 Mb
)

type MessageBlock struct {
	Block blockchain.BlockV2
}

func (self MessageBlock) MessageType() string {
	return CmdBlock
}

func (self MessageBlock) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self MessageBlock) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageBlock) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self MessageBlock) SetSenderID(senderID peer.ID) error {
	return nil
}

func (self *MessageBlock) SetIntendedReceiver(_ string) error {
	return nil
}

func (self *MessageBlock) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (self *MessageBlock) VerifyMsgSanity() error {
	return nil
}
