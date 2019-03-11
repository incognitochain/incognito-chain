package jsonresult

import (
	"github.com/constant-money/constant-chain/blockchain"
)

type GetBlocksBeaconResult struct {
	Hash              string     `json:"Hash"`
	Height            uint64     `json:"Height"`
	AggregatedSig     string     `json:"AggregatedSig"`
	R                 string     `json:"R"`
	BlockProducerSign string     `json:"BlockProducerSign"`
	BlockProducer     string     `json:"BlockProducer"`
	Version           int        `json:"Version"`
	Epoch             uint64     `json:"Epoch"`
	Round             int        `json:"Round"`
	Time              int64      `json:"Time"`
	PreviousBlockHash string     `json:"PreviousBlockHash"`
	NextBlockHash     string     `json:"NextBlockHash"`
	Instructions      [][]string `json:"Instructions"`
	Size              uint64     `json:"Size"`
}

type GetBlockResult struct {
	Hash              string             `json:"Hash"`
	ShardID           byte               `json:"ShardID"`
	Height            uint64             `json:"Height"`
	Confirmations     int64              `json:"Confirmations"`
	Version           int                `json:"Version"`
	TxRoot            string             `json:"TxRoot"`
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
	R                 string             `json:"R"`
	Round             int                `json:"Round"`
	CrossShards       []int              `json:"CrossShards"`
	Epoch             uint64             `json:"Epoch"`
	Reward            uint64             `json:"Reward"` // salary tx output
	Fee               uint64             `json:"Fee"`    // total fee
	Size              uint64             `json:"Size"`
}

type GetBlockTxResult struct {
	Hash     string `json:"Hash"`
	Locktime int64  `json:"Locktime"`
	HexData  string `json:"HexData"`
}

func (getBlockResult *GetBlocksBeaconResult) Init(block *blockchain.BeaconBlock, size uint64) {
	getBlockResult.Version = block.Header.Version
	getBlockResult.Hash = block.Hash().String()
	getBlockResult.Height = block.Header.Height
	getBlockResult.AggregatedSig = block.AggregatedSig
	getBlockResult.R = block.R
	getBlockResult.BlockProducer = block.Header.Producer
	getBlockResult.BlockProducerSign = block.ProducerSig
	getBlockResult.Epoch = block.Header.Epoch
	getBlockResult.Round = block.Header.Round
	getBlockResult.Time = block.Header.Timestamp
	getBlockResult.PreviousBlockHash = block.Header.PrevBlockHash.String()
	getBlockResult.Instructions = block.Body.Instructions
	getBlockResult.Size = size
}

func (getBlockResult *GetBlockResult) Init(block *blockchain.ShardBlock, size uint64) {
	getBlockResult.BlockProducerSign = block.ProducerSig
	getBlockResult.BlockProducer = block.Header.Producer
	getBlockResult.Hash = block.Hash().String()
	getBlockResult.PreviousBlockHash = block.Header.PrevBlockHash.String()
	getBlockResult.Version = block.Header.Version
	getBlockResult.Height = block.Header.Height
	getBlockResult.Time = block.Header.Timestamp
	getBlockResult.ShardID = block.Header.ShardID
	getBlockResult.TxRoot = block.Header.TxRoot.String()
	getBlockResult.TxHashes = make([]string, 0)
	getBlockResult.Fee = uint64(0)
	getBlockResult.Size = size
	for _, tx := range block.Body.Transactions {
		getBlockResult.TxHashes = append(getBlockResult.TxHashes, tx.Hash().String())
		getBlockResult.Fee += tx.GetTxFee()
	}
	getBlockResult.BeaconHeight = block.Header.BeaconHeight
	getBlockResult.BeaconBlockHash = block.Header.BeaconHash.String()
	getBlockResult.AggregatedSig = block.AggregatedSig
	getBlockResult.R = block.R
	getBlockResult.Round = block.Header.Round
	getBlockResult.CrossShards = []int{}
	if len(block.Header.CrossShards) > 0 {
		for _, shardID := range block.Header.CrossShards {
			getBlockResult.CrossShards = append(getBlockResult.CrossShards, int(shardID))
		}
	}
	getBlockResult.Epoch = block.Header.Epoch
	if len(block.Body.Transactions) > 0 && block.Body.Transactions[0].GetProof() != nil {
		getBlockResult.Reward = block.Body.Transactions[0].GetProof().OutputCoins[0].CoinDetails.Value
	}
}
