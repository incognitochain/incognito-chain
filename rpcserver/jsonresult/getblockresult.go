package jsonresult

import "github.com/ninjadotorg/constant/blockchain"

type GetBlockResult struct {
	Hash              string             `json:"Hash"`
	ShardID           byte               `json:"ShardID"`
	Height            uint64             `json:"Height"`
	Confirmations     int64              `json:"Confirmations"`
	Version           int                `json:"Version"`
	MerkleRoot        string             `json:"TransactionRoot"`
	Time              int64              `json:"Time"`
	PreviousBlockHash string             `json:"PreviousBlockHash"`
	NextBlockHash     string             `json:"NextBlockHash"`
	TxHashes          []string           `json:"TxHashes"`
	Txs               []GetBlockTxResult `json:"Txs"`
	BlockProducerSign string             `json:"BlockProducerSign"`
	BlockProducer     string             `json:"BlockProducer"`
	Data              string             `json:"Data"`
	BeaconHeight      uint64             `json:"BeaconHeight"`
	BeaconBlockHash   string             `json:"BeaconBlockHash"`
	AggregatedSig     string             `json:"AggregatedSig"`
}

type GetBlockTxResult struct {
	Hash     string `json:"Hash"`
	Locktime int64  `json:"Locktime"`
	HexData  string `json:"HexData"`
}

func (getBlockResult *GetBlockResult) Init(block *blockchain.ShardBlock) {
	getBlockResult.BlockProducerSign = block.ProducerSig
	getBlockResult.BlockProducer = block.Header.Producer
	getBlockResult.Hash = block.Hash().String()
	getBlockResult.PreviousBlockHash = block.Header.PrevBlockHash.String()
	getBlockResult.Version = block.Header.Version
	getBlockResult.Height = block.Header.Height
	getBlockResult.Time = block.Header.Timestamp
	getBlockResult.ShardID = block.Header.ShardID
	getBlockResult.MerkleRoot = block.Header.ShardTxRoot.String()
	getBlockResult.TxHashes = make([]string, 0)
	for _, tx := range block.Body.Transactions {
		getBlockResult.TxHashes = append(getBlockResult.TxHashes, tx.Hash().String())
	}
	getBlockResult.BeaconHeight = block.Header.BeaconHeight
	getBlockResult.BeaconBlockHash = block.Header.BeaconHash.String()
	getBlockResult.AggregatedSig = block.AggregatedSig
}
