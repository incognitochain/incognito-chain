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
	MaxTxPayload = 4000000 // 4 Mb
)

type MessageTx struct {
	Transaction metadata.Transaction
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
	jsonDecodeString, _ := hex.DecodeString(jsonStr)
	err := json.Unmarshal([]byte(jsonDecodeString), msg)
	return err
}

func (msg *MessageTx) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *MessageTx) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (msg *MessageTx) VerifyMsgSanity() error {
	return nil
}
