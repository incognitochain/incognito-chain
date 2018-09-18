package wire

import (
	"encoding/json"

	"github.com/ninjadotorg/cash-prototype/common"
)

type MessageGetBlocks struct {
	LastBlockHash common.Hash
	SenderID      string
}

func (self MessageGetBlocks) MessageType() string {
	return CmdGetBlocks
}

func (self MessageGetBlocks) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self MessageGetBlocks) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageGetBlocks) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}
