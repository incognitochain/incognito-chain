package wire

import (
	"encoding/json"
	"encoding/hex"
)

type MessagGetData struct {
	InvList []InvVect
}

func (self MessagGetData) MessageType() string {
	return CmdGetData
}

func (self MessagGetData) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self MessagGetData) JsonSerialize() (string, error) {
	jsonStr, err := json.Marshal(self)
	header := make([]byte, MessageHeaderSize)
	copy(header[:], self.MessageType())
	jsonStr = append(jsonStr, header...)
	return hex.EncodeToString(jsonStr), err
}

func (self MessagGetData) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}
