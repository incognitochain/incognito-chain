package wire

import (
	"encoding/json"
	"time"
	"net"
)

type MessageVersion struct {
	ProtocolVersion int
	Timestamp       time.Time
	RemoteAddress   net.Addr
	LocalAddress    net.Addr
	LastBlock       int
}

func (self MessageVersion) MessageType() string {
	return CmdVersion
}

func (self MessageVersion) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self MessageVersion) JsonSerialize() (string, error) {
	jsonStr, err := json.Marshal(self)
	header := make([]byte, MessageHeaderSize)
	copy(header[:], self.MessageType())
	jsonStr = append(jsonStr, header...)
	return string(jsonStr), err
}

func (self MessageVersion) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}
