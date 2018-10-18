package jsonresult

type GetBlockResult struct {
	Data              string             `json:"Data"`
	Hash              string             `json:"Hash"`
	Confirmations     int64              `json:"confirmations"`
	Height            int32              `json:"Height"`
	Version           int                `json:"Version"`
	MerkleRoot        string             `json:"MerkleRoot"`
	Time              int64              `json:"Time"`
	ChainID           byte               `json:"ChainID"`
	PreviousBlockHash string             `json:"PreviousBlockHash"`
	NextBlockHash     string             `json:"NextBlockHash"`
	TxHashes          []string           `json:"TxHashes"`
	Txs               []GetBlockTxResult `json:"Txs"`
}

type GetBlockTxResult struct {
	Hash     string `json:"Hash"`
	Locktime int64  `json:"Locktime"`
	HexData  string `json:"HexData"`
}
