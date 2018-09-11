package wire

import (
	"encoding/json"

	peer "github.com/libp2p/go-libp2p-peer"
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

func (self MessagGetData) SetSenderID(senderID peer.ID) error {
	return nil
}
