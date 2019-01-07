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

type MessageShardToBeacon struct {
	Block blockchain.ShardToBeaconBlock
}

func (self *MessageShardToBeacon) Hash() string {
	rawBytes, err := self.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (self *MessageShardToBeacon) MessageType() string {
	return CmdBlkShardToBeacon
}

func (self *MessageShardToBeacon) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self *MessageShardToBeacon) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self *MessageShardToBeacon) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self *MessageShardToBeacon) SetSenderID(senderID peer.ID) error {
	return nil
}

func (self *MessageShardToBeacon) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (self *MessageShardToBeacon) VerifyMsgSanity() error {
	return nil
}
