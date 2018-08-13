package wire

import (
	"github.com/internet-cash/prototype/transaction"
	"encoding/json"
)

type MessageTransaction struct {
	Transaction transaction.Tx
}

func (self MessageTransaction) MessageType() string {
	return CmdTx
}

func (self MessageTransaction) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self MessageTransaction) JsonSerialize() string {
	jsonStr, _ := json.Marshal(self)
	return string(jsonStr)
}

func (self MessageTransaction) JsonDeserialize(jsonStr string) {
	_ = json.Unmarshal([]byte(jsonStr), self)
}
