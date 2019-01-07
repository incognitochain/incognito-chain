package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
)

const (
	MaxGetBlockPayload = 1000 // 1kb
)

type MessageGetBlocks struct {
	LastBlockHash string
	SenderID      string
}

func (self *MessageGetBlocks) Hash() string {
	return ""
}

func (self *MessageGetBlocks) MessageType() string {
	return CmdGetBlocks
}

func (self *MessageGetBlocks) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self *MessageGetBlocks) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self *MessageGetBlocks) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self *MessageGetBlocks) SetSenderID(senderID peer.ID) error {
	self.SenderID = senderID.Pretty()
	return nil
}

func (self *MessageGetBlocks) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (self *MessageGetBlocks) VerifyMsgSanity() error {
	return nil
}
