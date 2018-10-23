package wire

import (
	"encoding/hex"
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/cash/transaction"
)

const (
	MaxTxRegisterationPayload = 4000000 // 4 Mb
)

type MessageRegisteration struct {
	Transaction transaction.Transaction
}

func (self MessageRegisteration) MessageType() string {
	return CmdRegisteration
}

func (self MessageRegisteration) MaxPayloadLength(pver int) int {
	return MaxTxRegisterationPayload
}

func (self MessageRegisteration) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageRegisteration) JsonDeserialize(jsonStr string) error {
	jsonDecodeString, _ := hex.DecodeString(jsonStr)
	err := json.Unmarshal([]byte(jsonDecodeString), self)
	return err
}

func (self MessageRegisteration) SetSenderID(senderID peer.ID) error {
	return nil
}
