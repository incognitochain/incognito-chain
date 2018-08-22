package wire

import (
	"encoding/json"
	"encoding/hex"
)

// MsgVerAck defines a bitcoin verack message which is used for a peer to
// acknowledge a version message (MsgVersion) after it has used the information
// to negotiate parameters.  It implements the Message interface.
//
// This message has no payload.
type MessageVerAck struct{}

func (self MessageVerAck) MessageType() string {
	return Cmdveack
}

func (self MessageVerAck) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self MessageVerAck) JsonSerialize() (string, error) {
	jsonStr, err := json.Marshal(self)
	header := make([]byte, MessageHeaderSize)
	copy(header[:], self.MessageType())
	jsonStr = append(jsonStr, header...)
	return hex.EncodeToString(jsonStr), err
}

func (self MessageVerAck) JsonDeserialize(jsonStr string) error {
	jsonDecodeString, _ := hex.DecodeString(jsonStr)
	err := json.Unmarshal([]byte(jsonDecodeString), self)
	return err
}
