package model

import (
	"github.com/incognitochain/incognito-chain/appservices/data"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type ShardState struct {
	ShardID           byte               `json:"ShardID"`
	BlockHash         string             `json:"Hash"`
	PreviousBlockHash string             `json:"PreviousBlockHash"`
	NextBlockHash     string             `json:"NextBlockHash"`
	Height            uint64             `json:"Height"`
	Version           int                `json:"Version"`
	TxRoot            string             `json:"TxRoot"`
	ShardTxRoot            string            `json:"ShardTxRoot"`           // output root created for other shard
	CrossTransactionRoot   string            `json:"CrossTransactionRoot"`  // transaction root created from transaction of micro shard to shard block (from other shard)
	InstructionsRoot      string            `json:"InstructionsRoot"`      // actions root created from Instructions and Metadata of transaction
	CommitteeRoot         string            `json:"CommitteeRoot"`         // hash from public key list of all committees designated to create this block
	PendingValidatorRoot  string            `json:"PendingValidatorRoot"`  // hash from public key list of all pending validators designated to this ShardID
	StakingTxRoot         string            `json:"StakingTxRoot"`         // hash from staking transaction map in shard best state
	InstructionMerkleRoot string            `json:"InstructionMerkleRoot"` // Merkle root of all instructions (using Keccak256 hash func) to relay to Ethreum
	TotalTxsFee           map[string]uint64 `json:"TotalTxsFee"`           // fee of all txs in block
	Time              int64              `json:"Time"`
	TxHashes          []string           `json:"TxHashes"`
	Txs               []data.TxInfo			 `json:"Txs"`
	BlockProducer     string             `json:"BlockProducer"`
    BlockProducerPubKeyStr          string                            `json:"BlockProducerPubKeyStr"`
	Proposer    			string							 `json:"Proposer"`
	ProposeTime 			int64							 `json:"ProposeTime"`
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

	MaxShardCommitteeSize  int                               `json:"MaxShardCommitteeSize"`
	MinShardCommitteeSize  int                               `json:"MinShardCommitteeSize"`
	ShardProposerIdx       int                               `json:"ShardProposerIdx"`
	MetricBlockHeight      uint64   `json:"MetricBlockHeight"`
	BestCrossShard         map[byte]uint64                   `json:"BestCrossShard"` // Best cross shard block by heigh

	ShardCommittee         []incognitokey.CommitteeKeyString `json:"ShardCommittee"`
	ShardPendingValidator  []incognitokey.CommitteeKeyString `json:"ShardPendingValidator"`
	StakingTx              map[string]string                  `json:"StakingTx"`


}