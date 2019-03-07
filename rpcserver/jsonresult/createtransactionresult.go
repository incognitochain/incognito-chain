package jsonresult

type CreateTransactionResult struct {
	Base58CheckData string
	//HexData         string
	TxID    string
	ShardID byte
}

type CreateTransactionCustomTokenResult struct {
	Base58CheckData string
	ShardID         byte   `json:"ShardID"`
	TxID            string `json:"TxID"`
	TokenID         string `json:"TokenID"`
	TokenName       string `json:"TokenName"`
	TokenAmount     uint64 `json:"TokenAmount"`
}
