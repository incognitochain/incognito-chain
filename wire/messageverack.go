package wire

import (
	"encoding/hex"
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"time"
)

type MessageVerAck struct {
	Valid     bool
	Timestamp time.Time
}

func (self MessageVerAck) MessageType() string {
	return CmdVerack
}

func (self MessageVerAck) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self MessageVerAck) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageVerAck) JsonDeserialize(jsonStr string) error {
	jsonDecodeString, _ := hex.DecodeString(jsonStr)
	err := json.Unmarshal([]byte(jsonDecodeString), self)
	return err
}

func (self MessageVerAck) SetSenderID(senderID peer.ID) error {
	return nil
}
