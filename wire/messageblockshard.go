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

type MessageBlockShard struct {
	Block blockchain.BlockV2
}

func (self MessageBlockShard) MessageType() string {
	return CmdBlock
}

func (self MessageBlockShard) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self MessageBlockShard) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageBlockShard) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self MessageBlockShard) SetSenderID(senderID peer.ID) error {
	return nil
}

func (self *MessageBlockShard) SetIntendedReceiver(_ string) error {
	return nil
}

func (self *MessageBlockShard) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (self *MessageBlockShard) VerifyMsgSanity() error {
	return nil
}
