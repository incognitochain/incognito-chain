package wire

import (
	"encoding/json"

	peer "github.com/libp2p/go-libp2p-peer"
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

func (self MessageGetAddr) SetSenderID(senderID peer.ID) error {
	return nil
}
