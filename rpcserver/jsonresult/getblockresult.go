package jsonresult

import "github.com/ninjadotorg/constant/blockchain"

type GetBlockResult struct {
	Data              string             `json:"Data"`
	Hash              string             `json:"Hash"`
	Confirmations     int64              `json:"confirmations"`
	Height            uint64             `json:"Height"`
	Version           int                `json:"Version"`
	MerkleRoot        string             `json:"MerkleRoot"`
	Time              int64              `json:"Time"`
	ShardID           byte               `json:"shardID"`
	PreviousBlockHash string             `json:"PreviousBlockHash"`
	NextBlockHash     string             `json:"NextBlockHash"`
	TxHashes          []string           `json:"TxHashes"`
	Txs               []GetBlockTxResult `json:"Txs"`
	BlockProducerSign string             `json:"BlockProducerSign"`
	BlockProducer     string             `json:"BlockProducer"`
}

type GetBlockTxResult struct {
	Hash     string `json:"Hash"`
	Locktime int64  `json:"Locktime"`
	HexData  string `json:"HexData"`
}

func (getBlockResult *GetBlockResult) Init(block *blockchain.ShardBlock) {
// 	getBlockResult.BlockProducerSign = block.BlockProducerSig
// 	getBlockResult.BlockProducer = block.BlockProducer
// 	getBlockResult.Hash = block.Hash().String()
// 	getBlockResult.PreviousBlockHash = block.Header.PrevBlockHash.String()
// 	getBlockResult.Version = block.Header.Version
// 	getBlockResult.Height = block.Header.Height
// 	getBlockResult.Time = block.Header.Timestamp
// 	getBlockResult.ShardID = block.Header.ShardID
// 	getBlockResult.MerkleRoot = block.Header.MerkleRoot.String()
// 	getBlockResult.TxHashes = make([]string, 0)
// 	for _, tx := range block.Transactions {
// 		getBlockResult.TxHashes = append(getBlockResult.TxHashes, tx.Hash().String())
// 	}
}
