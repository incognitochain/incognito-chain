package blockchain

import (
	"github.com/internet-cash/prototype/transaction"
)

const (
	// Default length of list tx in block
	defaultTransactionAlloc = 2048
)

type MsgBlock struct {
	Header       BlockHeader
	Transactions []*transaction.Tx
}

func (msg *MsgBlock) AddTransaction(tx *transaction.Tx) error {
	msg.Transactions = append(msg.Transactions, tx)
	return nil

}

func (msg *MsgBlock) ClearTransactions() {
	msg.Transactions = make([]*transaction.Tx, 0, defaultTransactionAlloc)
}
