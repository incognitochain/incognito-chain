package jsonresult

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/rpcserver"
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

func NewGetRawMempoolResult(config rpcserver.RpcServerConfig) *GetRawMempoolResult {
	result := &GetRawMempoolResult{
		TxHashes: config.TxMemPool.ListTxs(),
	}
	return result
}

type GetMempoolEntryResult struct {
	Tx metadata.Transaction
}

func NewGetMempoolEntryResult(config rpcserver.RpcServerConfig, txID *common.Hash) (*GetMempoolEntryResult, error) {
	result := &GetMempoolEntryResult{}
	var err error
	result.Tx, err = config.TxMemPool.GetTx(txID)
	if err != nil {
		return nil, err
	}
	return result, nil
}
