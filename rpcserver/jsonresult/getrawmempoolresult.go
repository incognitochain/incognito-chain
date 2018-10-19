package jsonresult

import "github.com/ninjadotorg/cash/transaction"

type GetRawMempoolResult struct {
	TxHashes []string
}

type GetMempoolEntryResult struct {
	Tx transaction.Transaction
}
