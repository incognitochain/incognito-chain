package jsonresult

import "github.com/incognitochain/incognito-chain/common"

type ReceivedTransaction struct {
	Hash           string
	ReceivedAmount map[common.Hash]uint64
	FromShardID    byte
}

type ListReceivedTransaction struct {
	ReceivedTransactions []ReceivedTransaction
}
