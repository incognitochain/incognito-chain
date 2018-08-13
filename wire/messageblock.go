package wire

import (
	"github.com/internet-cash/prototype/blockchain"
	"encoding/json"
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

func (self MessageBlock) JsonSerialize() string {
	jsonStr, _ := json.Marshal(self)
	return string(jsonStr)
}

func (self MessageBlock) JsonDeserialize(jsonStr string) {
	_ = json.Unmarshal([]byte(jsonStr), self)
}
