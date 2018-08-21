package wire

import (
	"github.com/internet-cash/prototype/transaction"
	"encoding/json"
	"encoding/hex"
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

func (self MessageTx) JsonSerialize() (string, error) {
	jsonStr, err := json.Marshal(self)
	header := make([]byte, MessageHeaderSize)
	copy(header[:], self.MessageType())
	jsonStr = append(jsonStr, header...)
	return hex.EncodeToString(jsonStr), err
}

func (self MessageTx) JsonDeserialize(jsonStr string) error {
	jsonDecodeString, _ := hex.DecodeString(jsonStr)
	err := json.Unmarshal([]byte(jsonDecodeString), self)
	return err
}
