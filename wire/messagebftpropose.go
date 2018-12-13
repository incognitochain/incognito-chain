package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/blockchain"
)

const (
	MaxBFTProposePayload = 1000 // 1 Kb
)

type MessageBFTPropose struct {
	Phase string
	Block blockchain.BlockV2
}

func (self *MessageBFTPropose) MessageType() string {
	return CmdBlockSig
}

func (self *MessageBFTPropose) MaxPayloadLength(pver int) int {
	return MaxBlockSigPayload
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
