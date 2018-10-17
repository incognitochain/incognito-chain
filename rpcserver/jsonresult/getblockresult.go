package jsonresult

type GetBlockResult struct {
	Data              string                 `json:"Data"`
	Hash              string                 `json:"Hash"`
	Confirmations     int64                  `json:"confirmations"`
	Size              int                    `json:"Size"`
	StrippedSize      int                    `json:"StrippedSize"`
	Weight            int                    `json:"Weight"`
	Height            int32                  `json:"Height"`
	Version           int32                  `json:"Version"`
	VersionHex        string                 `json:"VersionHex"`
	MerkleRoot        string                 `json:"MerkleRoot"`
	Time              int64                  `json:"Time"`
	MedianTime        int64                  `json:"MedianTime"`
	Bits              string                 `json:"Bits"`
	ChainID           string                 `json:"ChainID"`
	PreviousBlockHash string                 `json:"PreviousBlockHash"`
	NextBlockHash     string                 `json:"NextBlockHash"`
	TxHashes          []string               `json:"TxHashes"`
	Txs               map[string]interface{} `json:"Txs"`
}

type GetBlockTxResult struct {
	Hash     string `json:"Hash"`
	Locktime int64  `json:"Locktime"`
	HexData  string `json:"HexData"`
}
