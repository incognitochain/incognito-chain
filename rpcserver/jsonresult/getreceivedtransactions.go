package jsonresult

import (
	"github.com/incognitochain/incognito-chain/common"
)

type ReceivedTransaction struct {
	TxID          string                       `json:"TxID"`
	ReceivedInfos map[common.Hash]ReceivedInfo `json:"ReceivedInfos"`
	FromShardID   byte                         `json:"FromShardID"`
	LockTime      int64                        `json:"LockTime"`
}

type ReceivedInfo struct {
	CoinDetails          ReceivedCoin `json:"CoinDetails"`
	CoinDetailsEncrypted string       `json:"CoinDetailsEncrypted"`
}

type ReceivedCoin struct {
	PublicKey string `json:"PublicKey"`
	Info      string `json:"Info"`
	Value     uint64 `json:"Value"`
}

type ListReceivedTransaction struct {
	ReceivedTransactions []ReceivedTransaction `json:"ReceivedTransactions"`
}
