package jsonresult

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
)

type CreateTransactionResult struct {
	Base58CheckData string
	TxID            string
	ShardID         byte
}

func NewCreateTransactionResult(txID *common.Hash, txIDString string, byteArrays []byte, txShardID byte) CreateTransactionResult {
	result := CreateTransactionResult{
		ShardID: txShardID,
	}
	if txID != nil {
		result.TxID = txID.String()
	}
	if len(txIDString) > 0 {
		result.TxID = txIDString
	}
	if len(byteArrays) > 0 {
		result.Base58CheckData = base58.Base58Check{}.Encode(byteArrays, 0x00)
	}
	return result
}

type CreateTransactionTokenResult struct {
	Base58CheckData string
	ShardID         byte   `json:"ShardID"`
	TxID            string `json:"TxID"`
	TokenID         string `json:"TokenID"`
	TokenName       string `json:"TokenName"`
	TokenAmount     uint64 `json:"TokenAmount"`
}

/*func NewCreateTransactionCustomTokenResult() *CreateTransactionCustomTokenResult {
	result := &CreateTransactionCustomTokenResult{

	}
}*/
