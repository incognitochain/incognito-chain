package wire

import (
	"encoding/json"

	"github.com/ninjadotorg/cash-prototype/blockchain"
)

const (
	MaxHeaderPayload = 4000000
)

type MessageBlockHeader struct {
	Header blockchain.BlockHeader
}

func (self MessageBlockHeader) MessageType() string {
	return CmdBlock
}

func (self MessageBlockHeader) MaxPayloadLength(pver int) int {
	return MaxHeaderPayload
}

func (self MessageBlockHeader) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageBlockHeader) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}
