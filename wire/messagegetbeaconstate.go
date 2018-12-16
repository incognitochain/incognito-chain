package wire

import (
	"encoding/json"

	"time"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
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
func (self *MessageGetChainState) SetIntendedReceiver(_ string) error {
	return nil
}

func (self *MessageGetChainState) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (self *MessageGetChainState) VerifyMsgSanity() error {
	return nil
}
