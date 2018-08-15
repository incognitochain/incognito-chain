package wire

import (
	"github.com/internet-cash/prototype/transaction"
	"encoding/json"
)

type MessageTx struct {
	Transaction transaction.Tx
}

func (self MessageTx) MessageType() string {
	return CmdTx
}

func (self MessageTx) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self MessageTx) JsonSerialize() string {
	jsonStr, _ := json.Marshal(self)
	return string(jsonStr)
}

func (self MessageTx) JsonDeserialize(jsonStr string) {
	_ = json.Unmarshal([]byte(jsonStr), self)
}
