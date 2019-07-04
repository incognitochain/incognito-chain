package wire

import (
	"encoding/hex"
	"encoding/json"

	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/libp2p/go-libp2p-peer"
)

const (
	MaxTxTokenPayload = 4000000 // 4 Mb
)

type MessageTxToken struct {
	Transaction metadata.Transaction
}

func (msg *MessageTxToken) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageTxToken) MessageType() string {
	return CmdCustomToken
}

func (msg *MessageTxToken) MaxPayloadLength(pver int) int {
	return MaxTxTokenPayload
}

func (msg *MessageTxToken) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageTxToken) JsonDeserialize(jsonStr string) error {
	jsonDecodeString, _ := hex.DecodeString(jsonStr)
	err := json.Unmarshal([]byte(jsonDecodeString), msg)
	return err
}

func (msg *MessageTxToken) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *MessageTxToken) SignMsg(_ *incognitokey.KeySet) error {
	return nil
}

func (msg *MessageTxToken) VerifyMsgSanity() error {
	return nil
}
