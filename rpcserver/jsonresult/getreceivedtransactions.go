package jsonresult

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

type ReceivedTransaction struct {
	Hash          string
	ReceivedInfos map[common.Hash]ReceivedInfo
	FromShardID   byte
}

type ReceivedInfo struct {
	OutputCoin privacy.OutputCoin
}

type ListReceivedTransaction struct {
	ReceivedTransactions []ReceivedTransaction
}
