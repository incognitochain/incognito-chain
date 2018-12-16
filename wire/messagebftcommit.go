package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
)

const (
	MaxBFTCommitPayload = 1000 // 1 Kb
)

type MessageBFTCommit struct {
	CommitSig     string
	R             []byte
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

func (self *MessageBFTCommit) SetIntendedReceiver(_ string) error {
	return nil
}

func (self *MessageBFTCommit) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (self *MessageBFTCommit) VerifyMsgSanity() error {
	return nil
}
