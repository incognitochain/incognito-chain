package wire

import (
	"encoding/hex"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/pkg/errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/libp2p/go-libp2p-peer"
)

const (
	MaxTxPayload = 4000000 // 4 Mb
)

type MessageTx struct {
	Transaction metadata.Transaction
}

func (msg *MessageTx) UnmarshalJSON(data []byte) error {

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
		tx, err := transaction.NewTransactionFromJsonBytes(*temp.Transaction)
		if err != nil {
			return errors.New("Cannot unmarshal message tx, new transaction from json byte error")
		}
		msg.Transaction = tx
	}
	return nil
}

func (msg *MessageTx) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageTx) MessageType() string {
	return CmdTx
}

func (msg *MessageTx) MaxPayloadLength(pver int) int {
	return MaxTxPayload
}

func (msg *MessageTx) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageTx) JsonDeserialize(jsonStr string) error {
	jsonDecode, _ := hex.DecodeString(jsonStr)
	tx, err := transaction.NewTransactionFromJsonBytes(jsonDecode)
	if err != nil {
		return err
	}
	msg.Transaction = tx
	return nil
}

func (msg *MessageTx) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *MessageTx) SignMsg(_ *incognitokey.KeySet) error {
	return nil
}

func (msg *MessageTx) VerifyMsgSanity() error {
	return nil
}
