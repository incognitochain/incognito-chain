package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
)

const (
	MaxBFTCommitPayload = 1000 // 1 Kb
)

type MessageBFTCommit struct {
	CommitSig     string
	R             string
	ValidatorsIdx []int
	Pubkey        string
	MsgSig        string
}

func (msg *MessageBFTCommit) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageBFTCommit) MessageType() string {
	return CmdBFTCommit
}

func (msg *MessageBFTCommit) MaxPayloadLength(pver int) int {
	return MaxBFTCommitPayload
}

func (msg *MessageBFTCommit) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageBFTCommit) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *MessageBFTCommit) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *MessageBFTCommit) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (msg *MessageBFTCommit) VerifyMsgSanity() error {
	return nil
}
