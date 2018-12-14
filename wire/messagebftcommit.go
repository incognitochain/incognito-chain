package wire

import (
	"encoding/json"

	"github.com/ninjadotorg/constant/privacy-protocol"

	"github.com/libp2p/go-libp2p-peer"
)

const (
	MaxBFTCommitPayload = 1000 // 1 Kb
)

type MessageBFTCommit struct {
	CommitSig     string
	R             privacy.EllipticPoint
	ValidatorsIdx []int
	Pubkey        string
	MsgSig        string
}

func (self *MessageBFTCommit) MessageType() string {
	return CmdBFTCommit
}

func (self *MessageBFTCommit) MaxPayloadLength(pver int) int {
	return MaxBFTCommitPayload
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
