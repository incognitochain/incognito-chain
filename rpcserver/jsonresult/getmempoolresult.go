package jsonresult

import "github.com/constant-money/constant-chain/metadata"

type GetMempoolInfo struct {
	Size          int                `json:"Size"`
	Bytes         uint64             `json:"Bytes"`
	Usage         uint64             `json:"Usage"`
	MaxMempool    uint64             `json:"MaxMempool"`
	MempoolMinFee uint64             `json:"MempoolMinFee"`
	MempoolMaxFee uint64             `json:"MempoolMaxFee"`
	ListTxs       []GetMempoolInfoTx `json:"ListTxs"`
}

type GetMempoolInfoTx struct {
	TxID     string `json:"TxID"`
	LockTime int64  `json:"LockTime"`
}

type GetRawMempoolResult struct {
	TxHashes []string
}

type GetMempoolEntryResult struct {
	Tx metadata.Transaction
}
