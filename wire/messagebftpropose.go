package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/cashec"
)

const (
	MaxBFTProposePayload = 1000 // 1 Kb
)

type MessageBFTPropose struct {
	AggregatedSig string
	ValidatorsIdx []int
	Block         blockchain.BFTBlockInterface
	MsgSig        string
}

func (self *MessageBFTPropose) MessageType() string {
	return CmdBFTPropose
}

func (self *MessageBFTPropose) MaxPayloadLength(pver int) int {
	return MaxBFTProposePayload
}

func (self *MessageBFTPropose) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self *MessageBFTPropose) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self *MessageBFTPropose) SetSenderID(senderID peer.ID) error {
	return nil
}

func (self *MessageBFTPropose) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (self *MessageBFTPropose) VerifyMsgSanity() error {
	return nil
}
