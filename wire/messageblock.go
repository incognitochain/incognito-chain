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

func (self MessageBlock) JsonSerialize() (string, error) {
	jsonStr, err := json.Marshal(self)
	header := make([]byte, MessageHeaderSize)
	copy(header[:], self.MessageType())
	jsonStr = append(jsonStr, header...)
	return string(jsonStr), err
}

func (self MessageBlock) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}
