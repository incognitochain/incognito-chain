package model

import "github.com/incognitochain/incognito-chain/appservices/data"

type ShardState struct {
	ShardID           byte               `json:"ShardID"`
	BlockHash         string             `json:"Hash"`
	PreviousBlockHash string             `json:"PreviousBlockHash"`
	Height            uint64             `json:"Height"`
	Version           int                `json:"Version"`
	TxRoot            string             `json:"TxRoot"`
	Time              int64              `json:"Time"`
	TxHashes          []string           `json:"TxHashes"`
	Txs               []data.TxInfo			 `json:"Txs"`
	BlockProducer     string             `json:"BlockProducer"`
	ValidationData    string             `json:"ValidationData"`
	ConsensusType     string             `json:"ConsensusType"`
	Data              string             `json:"Data"`
	BeaconHeight      uint64             `json:"BeaconHeight"`
	BeaconBlockHash   string             `json:"BeaconBlockHash"`
	Round             int                `json:"Round"`
	Epoch             uint64             `json:"Epoch"`
	Reward            uint64             `json:"Reward"`
	RewardBeacon      uint64             `json:"RewardBeacon"`
	Fee               uint64             `json:"Fee"`
	Size              uint64             `json:"Size"`
	Instruction       		[][]string         `json:"Instruction"`
	CrossShardBitMap  		[]int              `json:"CrossShardBitMap"`
	NumTxns                uint64                            `json:"NumTxns"`                // The number of txns in the block.
	TotalTxns              uint64                            `json:"TotalTxns"`              // The total number of txns in the chain.
	NumTxnsExcludeSalary   uint64							  `json:"NumTxnsExcludeSalary"`
	TotalTxnsExcludeSalary uint64                            `json:"TotalTxnsExcludeSalary"` // for testing and benchmark
	ActiveShards           int                               `json:"ActiveShards"`
	ConsensusAlgorithm     string                            `json:"ConsensusAlgorithm"`
	NumOfBlocksByProducers map[string]uint64 			`json:"NumOfBlocksByProducers"`
}