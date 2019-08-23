package jsonresult

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/mempool"
	"github.com/incognitochain/incognito-chain/metadata"
)

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

func NewGetRawMempoolResult(txMemPool mempool.TxPool) *GetRawMempoolResult {
	result := &GetRawMempoolResult{
		TxHashes: txMemPool.ListTxs(),
	}
	return result
}

type GetMempoolEntryResult struct {
	Tx metadata.Transaction
}

func NewGetMempoolEntryResult(txMemPool mempool.TxPool, txID *common.Hash) (*GetMempoolEntryResult, error) {
	result := &GetMempoolEntryResult{}
	var err error
	result.Tx, err = txMemPool.GetTx(txID)
	if err != nil {
		return nil, err
	}
	return result, nil
}
