package wire

import (
	"encoding/json"

	"github.com/ninjadotorg/cash-prototype/blockchain"
)

type MessageRequestSign struct {
	Block blockchain.Block
}

func (self MessageRequestSign) MessageType() string {
	return CmdRequestSign
}

func (self MessageRequestSign) MaxPayloadLength(pver int) int {
	return MaxHeaderPayload
}

func (self MessageRequestSign) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageRequestSign) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}
