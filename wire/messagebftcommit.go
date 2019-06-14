package wire

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/cashec"
	"github.com/incognitochain/incognito-chain/common"
	peer "github.com/libp2p/go-libp2p-peer"
)

const (
	MaxBFTCommitPayload = 2000 // 1 Kb
)

type MessageBFTCommit struct {
	CommitSig     string
	R             string
	ValidatorsIdx []int
	Pubkey        string
	ContentSig    string
	Timestamp     int64
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

func (msg *MessageBFTCommit) SignMsg(keySet *cashec.KeySet) error {
	dataBytes := []byte{}
	dataBytes = append(dataBytes, []byte(msg.CommitSig)...)
	dataBytes = append(dataBytes, []byte(msg.R)...)
	dataBytes = append(dataBytes, []byte(msg.Pubkey)...)
	dataBytes = append(dataBytes, []byte(fmt.Sprint(msg.ValidatorsIdx))...)
	dataBytes = append(dataBytes, []byte(fmt.Sprint(msg.Timestamp))...)
	var err error
	msg.ContentSig, err = keySet.SignDataB58(dataBytes)
	return err
}

func (msg *MessageBFTCommit) VerifyMsgSanity() error {
	dataBytes := []byte{}
	dataBytes = append(dataBytes, []byte(msg.CommitSig)...)
	dataBytes = append(dataBytes, []byte(msg.R)...)
	dataBytes = append(dataBytes, []byte(msg.Pubkey)...)
	dataBytes = append(dataBytes, []byte(fmt.Sprint(msg.ValidatorsIdx))...)
	dataBytes = append(dataBytes, []byte(fmt.Sprint(msg.Timestamp))...)
	err := cashec.ValidateDataB58(msg.Pubkey, msg.ContentSig, dataBytes)
	return err
}
