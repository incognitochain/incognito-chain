package wire

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/cashec"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/libp2p/go-libp2p-peer"
)

type MessageBFTProposeV2 struct {
	ChainKey   string
	Block      string
	ContentSig string
	Pubkey     string
	Timestamp  int64
	RoundKey   string
}

func (msg *MessageBFTProposeV2) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageBFTProposeV2) MessageType() string {
	return CmdBFTPropose
}

func (msg *MessageBFTProposeV2) MaxPayloadLength(pver int) int {
	return MaxBFTProposePayload
}

func (msg *MessageBFTProposeV2) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageBFTProposeV2) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *MessageBFTProposeV2) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *MessageBFTProposeV2) SignMsg(keySet *cashec.KeySet) (err error) {
	msg.ContentSig, err = keySet.SignDataB58(msg.GetBytes())
	return err
}

func (msg *MessageBFTProposeV2) VerifyMsgSanity() error {
	err := cashec.ValidateDataB58(msg.Pubkey, msg.ContentSig, msg.GetBytes())
	return err
}

func (msg *MessageBFTProposeV2) GetBytes() []byte {
	dataBytes := []byte{}
	dataBytes = append(dataBytes, []byte(msg.ChainKey)...)
	dataBytes = append(dataBytes, msg.Block...)
	dataBytes = append(dataBytes, []byte(msg.Pubkey)...)
	dataBytes = append(dataBytes, []byte(msg.RoundKey)...)
	dataBytes = append(dataBytes, []byte(fmt.Sprint(msg.Timestamp))...)
	return dataBytes
}
