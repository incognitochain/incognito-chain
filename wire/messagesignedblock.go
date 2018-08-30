package wire

import (
	"encoding/json"
)

type MessageSignedBlock struct {
}

func (self MessageSignedBlock) MessageType() string {
	return CmdRequestSign
}

func (self MessageSignedBlock) MaxPayloadLength(pver int) int {
	return MaxHeaderPayload
}

func (self MessageSignedBlock) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageSignedBlock) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}
