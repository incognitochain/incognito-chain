package wire

import (
	"encoding/json"
	"encoding/hex"
)

type MessageInv struct {
	InvList []InvVect
}

func (self MessageInv) MessageType() string {
	return CmdInv
}

func (self MessageInv) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self MessageInv) JsonSerialize() (string, error) {
	jsonStr, err := json.Marshal(self)
	header := make([]byte, MessageHeaderSize)
	copy(header[:], self.MessageType())
	jsonStr = append(jsonStr, header...)
	return hex.EncodeToString(jsonStr), err
}

func (self MessageInv) JsonDeserialize(jsonStr string) error {
	jsonDecodeString, _ := hex.DecodeString(jsonStr)
	err := json.Unmarshal([]byte(jsonDecodeString), self)
	return err
}
