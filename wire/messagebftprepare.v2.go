package wire

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/wire"
)

type MessageBFTPrepareV2 struct {
	ChainKey   string
	IsOk       bool
	Pubkey     string
	ContentSig string
	BlkHash    string
	RoundKey   string
	Timestamp  int64
}

func (msg *MessageBFTPrepareV2) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageBFTPrepareV2) MessageType() string {
	return wire.CmdBFTPrepare
}

func (msg *MessageBFTPrepareV2) MaxPayloadLength(pver int) int {
	return wire.MaxBFTPreparePayload
}

func (msg *MessageBFTPrepareV2) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageBFTPrepareV2) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *MessageBFTPrepareV2) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *MessageBFTPrepareV2) SignMsg(keySet *incognitokey.KeySet) (err error) {
	msg.ContentSig, err = keySet.SignDataB58(msg.GetBytes())
	return err
}

func (msg *MessageBFTPrepareV2) VerifyMsgSanity() error {
	err := incognitokey.ValidateDataB58(msg.Pubkey, msg.ContentSig, msg.GetBytes())
	return err
}

func (msg *MessageBFTPrepareV2) GetBytes() []byte {
	dataBytes := []byte{}
	var bitSetVar int8
	if msg.IsOk {
		bitSetVar = 1
	}
	dataBytes = append(dataBytes, byte(bitSetVar))
	dataBytes = append(dataBytes, []byte(msg.ChainKey)...)
	dataBytes = append(dataBytes, []byte(msg.BlkHash)...)
	dataBytes = append(dataBytes, []byte(msg.Pubkey)...)
	dataBytes = append(dataBytes, []byte(msg.RoundKey)...)
	dataBytes = append(dataBytes, []byte(fmt.Sprint(msg.Timestamp))...)
	return dataBytes
}
