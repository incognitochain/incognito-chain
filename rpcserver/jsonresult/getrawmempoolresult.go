package jsonresult

import "github.com/ninjadotorg/constant/transaction"

type GetRawMempoolResult struct {
	TxHashes []string
}

type GetMempoolEntryResult struct {
	Tx transaction.Transaction
}
