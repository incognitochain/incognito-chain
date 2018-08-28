package wire

import (
	"encoding/json"
)

type MessageGetBlockHeader struct {
}

func (self MessageGetBlockHeader) MessageType() string {
	return CmdGetBlockHeader
}

func (self MessageGetBlockHeader) MaxPayloadLength(pver int) int {
	return MaxHeaderPayload
}

func (self MessageGetBlockHeader) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageGetBlockHeader) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}
