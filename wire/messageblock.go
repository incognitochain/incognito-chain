package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/blockchain"
)

const (
	MaxBlockPayload = 4000000 // 4 Mb
)

type MessageBlock struct {
	Block blockchain.Block
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
