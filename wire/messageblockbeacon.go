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

type MessageBlockBeacon struct {
	Block blockchain.BeaconBlock
}

func (self *MessageBlockBeacon) Hash() string {
	rawBytes, err := self.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (self *MessageBlockBeacon) MessageType() string {
	return CmdBlockBeacon
}

func (self *MessageBlockBeacon) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self *MessageBlockBeacon) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self *MessageBlockBeacon) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self *MessageBlockBeacon) SetSenderID(senderID peer.ID) error {
	return nil
}

func (self *MessageBlockBeacon) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (self *MessageBlockBeacon) VerifyMsgSanity() error {
	return nil
}
