package wire

import (
	"github.com/internet-cash/prototype/transaction"
	"encoding/json"
	"encoding/hex"
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
	return hex.EncodeToString(jsonStr)
}

func (self MessageTx) JsonDeserialize(jsonStr string) {
	jsonDecodeString, _ := hex.DecodeString(jsonStr)
	_ = json.Unmarshal([]byte(jsonDecodeString), self)
}
