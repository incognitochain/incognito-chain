package wire

import (
	"github.com/internet-cash/prototype/transaction"
)

type MessageTransaction struct {
	Transaction *transaction.Tx
}

func (msg MessageTransaction) MessageType() string {
	return CmdTx
}

func (msg MessageTransaction) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}
