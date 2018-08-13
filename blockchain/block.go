package blockchain

const defaultTransactionAlloc = 2048

type MsgBlock struct {
	Header       BlockHeader
	Transactions []*MsgTx
}

func (msg *MsgBlock) AddTransaction(tx *MsgTx) error {
	msg.Transactions = append(msg.Transactions, tx)
	return nil

}

func (msg *MsgBlock) ClearTransactions() {
	msg.Transactions = make([]*MsgTx, 0, defaultTransactionAlloc)
}
