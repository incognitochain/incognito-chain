package jsonresult

import "github.com/constant-money/constant-chain/common"

type CreateTransactionResult struct {
	Base58CheckData string
	TxID            string
	ShardID         byte
	inputCoinsHash  []common.Hash
}

func (result *CreateTransactionResult) SetInputCoinsHash(list []common.Hash) {
	result.inputCoinsHash = list
}

func (result CreateTransactionResult) GetInputCoinsHash() []common.Hash {
	return result.inputCoinsHash
}

type CreateTransactionCustomTokenResult struct {
	Base58CheckData string
	ShardID         byte   `json:"ShardID"`
	TxID            string `json:"TxID"`
	TokenID         string `json:"TokenID"`
	TokenName       string `json:"TokenName"`
	TokenAmount     uint64 `json:"TokenAmount"`
	inputCoinsHash  []common.Hash
}

func (result *CreateTransactionCustomTokenResult) SetInputCoinsHash(list []common.Hash) {
	result.inputCoinsHash = list
}

func (result CreateTransactionCustomTokenResult) GetInputCoinsHash() []common.Hash {
	return result.inputCoinsHash
}
