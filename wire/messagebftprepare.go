package wire

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/common"
	peer "github.com/libp2p/go-libp2p-peer"
)

const (
	MaxBFTAgreePayload = 2000 // 1 Kb
)

type MessageBFTAgree struct {
	BlkHash    common.Hash
	Ri         []byte
	Pubkey     string
	ContentSig string
	Timestamp  int64
}

func (msg *MessageBFTAgree) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageBFTAgree) MessageType() string {
	return CmdBFTAgree
}

func (msg *MessageBFTAgree) MaxPayloadLength(pver int) int {
	return MaxBFTAgreePayload
}

func (msg *MessageBFTAgree) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageBFTAgree) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *MessageBFTAgree) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *MessageBFTAgree) SignMsg(keySet *incognitokey.KeySet) error {
	dataBytes := []byte{}
	dataBytes = append(dataBytes, msg.BlkHash.GetBytes()...)
	dataBytes = append(dataBytes, []byte(msg.Pubkey)...)
	dataBytes = append(dataBytes, msg.Ri...)
	dataBytes = append(dataBytes, []byte(fmt.Sprint(msg.Timestamp))...)
	var err error
	msg.ContentSig, err = keySet.SignDataB58(dataBytes)
	return err
}

func (msg *MessageBFTAgree) VerifyMsgSanity() error {
	dataBytes := []byte{}
	dataBytes = append(dataBytes, msg.BlkHash.GetBytes()...)
	dataBytes = append(dataBytes, []byte(msg.Pubkey)...)
	dataBytes = append(dataBytes, msg.Ri...)
	dataBytes = append(dataBytes, []byte(fmt.Sprint(msg.Timestamp))...)
	err := incognitokey.ValidateDataB58(msg.Pubkey, msg.ContentSig, dataBytes)
	return err
}
