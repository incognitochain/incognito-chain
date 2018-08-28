package wire

import (
	"encoding/json"
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

func (self MessagGetData) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessagGetData) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}
