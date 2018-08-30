package wire

import (
	"encoding/json"
)

type MessageSignedBlock struct {
	BlockHash    string
	ChainID      byte
	Validator    string
	ValidatorSig string
}

func (self MessageSignedBlock) MessageType() string {
	return CmdSignedBlock
}

func (self MessageSignedBlock) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self MessageSignedBlock) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageSignedBlock) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}
