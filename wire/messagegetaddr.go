package wire

import (
	"encoding/json"
)

type MessageGetAddr struct {
}

func (self MessageGetAddr) MessageType() string {
	return CmdGetAddr
}

func (self MessageGetAddr) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self MessageGetAddr) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageGetAddr) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}
