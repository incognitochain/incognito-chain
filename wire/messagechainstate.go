package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"time"
)

const (
	MaxChainStatePayload = 4000000 // 4 Mb
)

type MessageChainState struct {
	Timestamp time.Time
	ChainInfo interface{}
	SenderID  string
}

func (self *MessageChainState) MessageType() string {
	return CmdChainState
}

func (self *MessageChainState) MaxPayloadLength(pver int) int {
	return MaxChainStatePayload
}

func (self *MessageChainState) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self *MessageChainState) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self *MessageChainState) SetSenderID(senderID peer.ID) error {
	self.SenderID = senderID.Pretty()
	return nil
}
