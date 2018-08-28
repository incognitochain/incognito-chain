package wire

import (
	"encoding/json"

	"github.com/ninjadotorg/cash-prototype/blockchain"
)

const (
	MaxBlockPayload = 4000000
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
