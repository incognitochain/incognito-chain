package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/blockchain"
)

const (
	MaxBFTCommitPayload = 1000 // 1 Kb
)

type MessageBFTCommit struct {
	Phase string
	Block blockchain.BlockV2
}

func (self *MessageBFTCommit) MessageType() string {
	return CmdBlockSig
}

func (self *MessageBFTCommit) MaxPayloadLength(pver int) int {
	return MaxBlockSigPayload
}

func (self *MessageBFTCommit) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self *MessageBFTCommit) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self *MessageBFTCommit) SetSenderID(senderID peer.ID) error {
	return nil
}
