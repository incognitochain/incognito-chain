package wire

import (
	"encoding/json"
	"fmt"

	"github.com/constant-money/constant-chain/cashec"
	"github.com/constant-money/constant-chain/common"
	peer "github.com/libp2p/go-libp2p-peer"
)

const (
	MaxBFTProposePayload = 2000000 // 2000 Kb ~= 2 MB
)

type MessageBFTPropose struct {
	Layer      string
	ShardID    byte
	Block      json.RawMessage
	ContentSig string
	Pubkey     string
	Timestamp  int64
}

func (msg *MessageBFTPropose) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageBFTPropose) MessageType() string {
	return CmdBFTPropose
}

func (msg *MessageBFTPropose) MaxPayloadLength(pver int) int {
	return MaxBFTProposePayload
}

func (msg *MessageBFTPropose) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageBFTPropose) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *MessageBFTPropose) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *MessageBFTPropose) SignMsg(keySet *cashec.KeySet) error {
	dataBytes := []byte{}
	dataBytes = append(dataBytes, []byte(msg.Layer)...)
	dataBytes = append(dataBytes, msg.ShardID)
	dataBytes = append(dataBytes, msg.Block...)
	dataBytes = append(dataBytes, []byte(msg.Pubkey)...)
	dataBytes = append(dataBytes, []byte(fmt.Sprint(msg.Timestamp))...)
	var err error
	msg.ContentSig, err = keySet.SignDataB58(dataBytes)
	return err
}

func (msg *MessageBFTPropose) VerifyMsgSanity() error {
	dataBytes := []byte{}
	dataBytes = append(dataBytes, []byte(msg.Layer)...)
	dataBytes = append(dataBytes, msg.ShardID)
	dataBytes = append(dataBytes, msg.Block...)
	dataBytes = append(dataBytes, []byte(msg.Pubkey)...)
	dataBytes = append(dataBytes, []byte(fmt.Sprint(msg.Timestamp))...)
	err := cashec.ValidateDataB58(msg.Pubkey, msg.ContentSig, dataBytes)
	return err
}
