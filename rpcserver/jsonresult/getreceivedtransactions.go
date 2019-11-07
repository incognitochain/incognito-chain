package jsonresult

import (
	"github.com/incognitochain/incognito-chain/common"
)

type ReceivedTransaction struct {
	Hash          string
	ReceivedInfos map[common.Hash]ReceivedInfo
	FromShardID   byte
}

type ReceivedInfo struct {
	CoinDetails          ReceivedCoin
	CoinDetailsEncrypted string
}

type ReceivedCoin struct {
	PublicKey string
	Info      string
	Value     uint64
}

type ListReceivedTransaction struct {
	ReceivedTransactions []ReceivedTransaction
}
