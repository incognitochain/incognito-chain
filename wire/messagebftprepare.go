package wire

import (
	"encoding/json"
	"fmt"

	"github.com/big0t/constant-chain/cashec"
	"github.com/big0t/constant-chain/common"
	peer "github.com/libp2p/go-libp2p-peer"
)

const (
	MaxBFTPreparePayload = 1000 // 1 Kb
)

type MessageBFTPrepare struct {
	BlkHash    common.Hash
	Ri         []byte
	Pubkey     string
	ContentSig string
	Timestamp  int64
}

func (msg *MessageBFTPrepare) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageBFTPrepare) MessageType() string {
	return CmdBFTPrepare
}

func (msg *MessageBFTPrepare) MaxPayloadLength(pver int) int {
	return MaxBFTPreparePayload
}

func (msg *MessageBFTPrepare) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageBFTPrepare) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *MessageBFTPrepare) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *MessageBFTPrepare) SignMsg(keySet *cashec.KeySet) error {
	dataBytes := []byte{}
	dataBytes = append(dataBytes, msg.BlkHash.GetBytes()...)
	dataBytes = append(dataBytes, []byte(msg.Pubkey)...)
	dataBytes = append(dataBytes, msg.Ri...)
	dataBytes = append(dataBytes, []byte(fmt.Sprint(msg.Timestamp))...)
	var err error
	msg.ContentSig, err = keySet.SignDataB58(dataBytes)
	return err
}

func (msg *MessageBFTPrepare) VerifyMsgSanity() error {
	dataBytes := []byte{}
	dataBytes = append(dataBytes, msg.BlkHash.GetBytes()...)
	dataBytes = append(dataBytes, []byte(msg.Pubkey)...)
	dataBytes = append(dataBytes, msg.Ri...)
	dataBytes = append(dataBytes, []byte(fmt.Sprint(msg.Timestamp))...)
	err := cashec.ValidateDataB58(msg.Pubkey, msg.ContentSig, dataBytes)
	return err
}
