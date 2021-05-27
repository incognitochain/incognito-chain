package wire

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/transaction"

	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/libp2p/go-libp2p-peer"
)

const (
	MaxTxPrivacyTokenPayload = 4000000 // 4 Mb
)

type MessageTxPrivacyToken struct {
	Transaction metadata.Transaction
}

func (msg *MessageTxPrivacyToken) UnmarshalJSON(data []byte) error {
	temp := &struct {
		Transaction *json.RawMessage
	}{}
	err := json.Unmarshal(data, temp)
	if err != nil {
		return errors.New("Cannot unmarshal message tx temp struct")
	}
	if temp.Transaction == nil {
		return errors.New("Cannot unmarshal message tx, transaction is empty")
	} else {
		txToken, err := transaction.NewTransactionTokenFromJsonBytes(*temp.Transaction)
		if err != nil {
			return errors.New("Cannot unmarshal message tx, new transaction from json byte error")
		}
		msg.Transaction = txToken
	}
	return nil
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
	jsonDecode, err := hex.DecodeString(jsonStr)
	if err != nil {
		return err
	}
	txToken, err := transaction.NewTransactionTokenFromJsonBytes(jsonDecode)
	if err != nil {
		return err
	}
	msg.Transaction = txToken
	return nil
}

func (msg *MessageTxPrivacyToken) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *MessageTxPrivacyToken) SignMsg(_ *incognitokey.KeySet) error {
	return nil
}

func (msg *MessageTxPrivacyToken) VerifyMsgSanity() error {
	return nil
}
