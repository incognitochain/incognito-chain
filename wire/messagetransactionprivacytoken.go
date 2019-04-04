package wire

import (
	"encoding/hex"
	"encoding/json"

	"github.com/constant-money/constant-chain/cashec"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/libp2p/go-libp2p-peer"
)

const (
	MaxTxPrivacyTokenPayload = 4000000 // 4 Mb
)

type MessageTxPrivacyToken struct {
	Transaction metadata.Transaction
}

func (msg *MessageTxPrivacyToken) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageTxPrivacyToken) MessageType() string {
	return CmdPrivacyCustomToken
}

func (msg *MessageTxPrivacyToken) MaxPayloadLength(pver int) int {
	return MaxTxPrivacyTokenPayload
}

func (msg *MessageTxPrivacyToken) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageTxPrivacyToken) JsonDeserialize(jsonStr string) error {
	jsonDecodeString, _ := hex.DecodeString(jsonStr)
	err := json.Unmarshal([]byte(jsonDecodeString), msg)
	return err
}

func (msg *MessageTxPrivacyToken) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *MessageTxPrivacyToken) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (msg *MessageTxPrivacyToken) VerifyMsgSanity() error {
	return nil
}
