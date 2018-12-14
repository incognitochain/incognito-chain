package jsonresult

import "github.com/ninjadotorg/constant/metadata"

type GetRawMempoolResult struct {
	TxHashes []string
}

type GetMempoolEntryResult struct {
	Tx metadata.Transaction
}
