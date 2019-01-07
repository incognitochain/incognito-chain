package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
)

// const (
// 	MaxBlockPayload = 1000000 // 1 Mb
// )

type MessageCrossShard struct {
	Block blockchain.CrossShardBlock
}

func (self *MessageCrossShard) Hash() string {
	rawBytes, err := self.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (self *MessageCrossShard) MessageType() string {
	return CmdCrossShard
}

func (self *MessageCrossShard) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self *MessageCrossShard) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self *MessageCrossShard) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self *MessageCrossShard) SetSenderID(senderID peer.ID) error {
	return nil
}

func (self *MessageCrossShard) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (self *MessageCrossShard) VerifyMsgSanity() error {
	return nil
}
