package jsonresult

import "github.com/big0t/constant-chain/metadata"

type GetRawMempoolResult struct {
	TxHashes []string
}

type GetMempoolEntryResult struct {
	Tx metadata.Transaction
}
