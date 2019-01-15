package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
)

type MessageGetBlockShard struct {
	From      uint64
	To        uint64
	ShardID   byte
	SenderID  string
	Timestamp int64
}

func (self *MessageGetBlockShard) Hash() string {
	rawBytes, err := self.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (self *MessageGetBlockShard) MessageType() string {
	return CmdGetBlockShard
}

func (self *MessageGetBlockShard) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self *MessageGetBlockShard) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self *MessageGetBlockShard) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self *MessageGetBlockShard) SetSenderID(senderID peer.ID) error {
	self.SenderID = senderID.Pretty()
	return nil
}

func (self *MessageGetBlockShard) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (self *MessageGetBlockShard) VerifyMsgSanity() error {
	return nil
}
