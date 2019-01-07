package wire

import (
	"encoding/json"

	"time"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
)

const (
	MaxGetShardStatePayload = 1000 // 1kb
)

type MessageGetShardState struct {
	Timestamp time.Time
}

func (self *MessageGetShardState) Hash() string {
	return ""
}

func (self *MessageGetShardState) MessageType() string {
	return CmdGetShardState
}

func (self *MessageGetShardState) MaxPayloadLength(pver int) int {
	return MaxGetShardStatePayload
}

func (self *MessageGetShardState) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self *MessageGetShardState) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self *MessageGetShardState) SetSenderID(senderID peer.ID) error {
	// self.SenderID = senderID.Pretty()
	return nil
}

func (self *MessageGetShardState) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (self *MessageGetShardState) VerifyMsgSanity() error {
	return nil
}
