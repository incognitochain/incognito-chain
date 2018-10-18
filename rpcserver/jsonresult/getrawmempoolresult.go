package jsonresult

import "github.com/ninjadotorg/cash-prototype/transaction"

type GetRawMempoolResult struct {
	TxHashes []string
}

type GetMempoolEntryResult struct {
	Tx transaction.Transaction
}
