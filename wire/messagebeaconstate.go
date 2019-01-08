package wire

import (
	"encoding/json"

	"github.com/ninjadotorg/constant/blockchain"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
)

const (
	MaxBeaconStatePayload = 4000000 // 4 Mb
)

type MessageBeaconState struct {
	Timestamp int64
	ChainInfo blockchain.BeaconChainState
	SenderID  string
}

func (self *MessageBeaconState) Hash() string {
	rawBytes, err := self.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (self *MessageBeaconState) MessageType() string {
	return CmdBeaconState
}

func (self *MessageBeaconState) MaxPayloadLength(pver int) int {
	return MaxBeaconStatePayload
}

func (self *MessageBeaconState) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self *MessageBeaconState) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self *MessageBeaconState) SetSenderID(senderID peer.ID) error {
	self.SenderID = senderID.Pretty()
	return nil
}

func (self *MessageBeaconState) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (self *MessageBeaconState) VerifyMsgSanity() error {
	return nil
}
