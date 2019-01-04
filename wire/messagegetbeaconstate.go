package wire

import (
	"encoding/json"

	"time"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
)

const (
	MaxGetBeaconStatePayload = 1000 // 1kb
)

type MessageGetBeaconState struct {
	Timestamp time.Time
}

func (self *MessageGetBeaconState) MessageType() string {
	return CmdGetBeaconState
}

func (self *MessageGetBeaconState) MaxPayloadLength(pver int) int {
	return MaxGetBeaconStatePayload
}

func (self *MessageGetBeaconState) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self *MessageGetBeaconState) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self *MessageGetBeaconState) SetSenderID(senderID peer.ID) error {
	// self.SenderID = senderID.Pretty()
	return nil
}

func (self *MessageGetBeaconState) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (self *MessageGetBeaconState) VerifyMsgSanity() error {
	return nil
}
