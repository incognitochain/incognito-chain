package wire

import (
	"encoding/hex"
	"encoding/json"
)

// MsgVerAck defines a bitcoin verack message which is used for a peer to
// acknowledge a version message (MsgVersion) after it has used the information
// to negotiate parameters.  It implements the Message interface.
//
// This message has no payload.
type MessagePing struct {
}

func (self MessagePing) MessageType() string {
	return CmdPing
}

func (self MessagePing) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self MessagePing) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessagePing) JsonDeserialize(jsonStr string) error {
	jsonDecodeString, _ := hex.DecodeString(jsonStr)
	err := json.Unmarshal([]byte(jsonDecodeString), self)
	return err
}
