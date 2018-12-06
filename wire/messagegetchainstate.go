package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"time"
)

const (
	MaxGetChainStatePayload = 1000 // 1kb
)

type MessageGetChainState struct {
	Timestamp time.Time
	SenderID  string
}

func (self *MessageGetChainState) MessageType() string {
	return CmdGetChainState
}

func (self *MessageGetChainState) MaxPayloadLength(pver int) int {
	return MaxGetChainStatePayload
}

func (self *MessageGetChainState) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self *MessageGetChainState) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self *MessageGetChainState) SetSenderID(senderID peer.ID) error {
	self.SenderID = senderID.Pretty()
	return nil
}
