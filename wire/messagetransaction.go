package wire

import (
	"encoding/hex"
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/common"
)

const (
	MaxTxPayload = 4000000 // 4 Mb
)

type MessageTx struct {
	Transaction metadata.Transaction
}

func (self *MessageTx) Hash() string {
	rawBytes, err := self.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (self *MessageTx) MessageType() string {
	return CmdTx
}

func (self *MessageTx) MaxPayloadLength(pver int) int {
	return MaxTxPayload
}

func (self *MessageTx) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self *MessageTx) JsonDeserialize(jsonStr string) error {
	jsonDecodeString, _ := hex.DecodeString(jsonStr)
	err := json.Unmarshal([]byte(jsonDecodeString), self)
	return err
}

func (self *MessageTx) SetSenderID(senderID peer.ID) error {
	return nil
}

func (self *MessageTx) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (self *MessageTx) VerifyMsgSanity() error {
	return nil
}
