package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
)

const (
	MaxSwapUpdatePayload = 1000 // 1 Kb
)

type MessageSwapUpdate struct {
	LockTime   int64
	Requester  string
	shardID    byte
	Candidate  string
	Signatures map[string]string
}

func (self *MessageSwapUpdate) Hash() string {
	return ""
}

func (self *MessageSwapUpdate) MessageType() string {
	return CmdSwapUpdate
}

func (self *MessageSwapUpdate) MaxPayloadLength(pver int) int {
	return MaxSwapUpdatePayload
}

func (self *MessageSwapUpdate) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self *MessageSwapUpdate) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self *MessageSwapUpdate) SetSenderID(senderID peer.ID) error {
	return nil
}
