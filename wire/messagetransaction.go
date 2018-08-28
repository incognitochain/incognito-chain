package wire

import (
	"encoding/hex"
	"encoding/json"

	"github.com/ninjadotorg/cash-prototype/transaction"
)

type MessageTx struct {
	Transaction transaction.Transaction
}

func (self MessageTx) MessageType() string {
	return CmdTx
}

func (self MessageTx) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self MessageTx) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageTx) JsonDeserialize(jsonStr string) error {
	jsonDecodeString, _ := hex.DecodeString(jsonStr)
	err := json.Unmarshal([]byte(jsonDecodeString), self)
	return err
}
