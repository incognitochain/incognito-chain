package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
)

type MessageGetShardToBeacon struct {
	BlockHash common.Hash
	ShardID   byte
	SenderID  string
	Timestamp int64
}

func (self *MessageGetShardToBeacon) Hash() string {
	rawBytes, err := self.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (self *MessageGetShardToBeacon) MessageType() string {
	return CmdGetShardToBeacon
}

func (self *MessageGetShardToBeacon) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self *MessageGetShardToBeacon) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self *MessageGetShardToBeacon) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self *MessageGetShardToBeacon) SetSenderID(senderID peer.ID) error {
	self.SenderID = senderID.Pretty()
	return nil
}

func (self *MessageGetShardToBeacon) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (self *MessageGetShardToBeacon) VerifyMsgSanity() error {
	return nil
}
