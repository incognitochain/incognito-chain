package jsonresult

import "github.com/constant-money/constant-chain/metadata"

type GetRawMempoolResult struct {
	TxHashes []string
}

type GetMempoolEntryResult struct {
	Tx metadata.Transaction
}
