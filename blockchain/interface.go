package blockchain

import (
	"time"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
)

type BFTBlockInterface interface {
	// UnmarshalJSON(data []byte) error
}

type ShardToBeaconPool interface {
	RemovePendingBlock(map[byte]uint64)
	//GetFinalBlock() map[byte][]ShardToBeaconBlock
	AddShardToBeaconBlock(ShardToBeaconBlock) (uint64, uint64, error)
	//ValidateShardToBeaconBlock(ShardToBeaconBlock) error
	GetValidPendingBlockHash() map[byte][]common.Hash
	GetValidPendingBlock(map[byte]uint64) map[byte][]*ShardToBeaconBlock
	GetValidPendingBlockHeight() map[byte][]uint64
	GetLatestValidPendingBlockHeight() map[byte]uint64
	SetShardState(map[byte]uint64)
}

type CrossShardPool interface {
	AddCrossShardBlock(CrossShardBlock) error
	GetValidBlock(map[byte]uint64) map[byte][]*CrossShardBlock
	GetLatestValidBlockHeight() map[byte]uint64
	RemoveBlockByHeight(map[byte]uint64) error
	UpdatePool() error
}

type NodeShardPool interface {
	PushBlock(ShardBlock) error
	GetBlocks(byte, uint64) ([]ShardBlock, error)
	RemoveBlocks(byte, uint64) error
}

type BeaconPool interface {
	RemoveBlock(uint64)
	AddBeaconBlock(block BeaconBlock) error
	GetValidBlockHash() []common.Hash
	GetValidBlock() []*BeaconBlock
	GetValidBlockHeight() []uint64
	GetLatestValidBlockHeight() uint64
	SetBeaconState(uint64)
}
type TxPool interface {
	// LastUpdated returns the last time a transaction was added to or
	// removed from the source pool.
	LastUpdated() time.Time

	// MiningDescs returns a slice of mining descriptors for all the
	// transactions in the source pool.
	MiningDescs() []*metadata.TxDesc

	// HaveTransaction returns whether or not the passed transaction hash
	// exists in the source pool.
	HaveTransaction(hash *common.Hash) bool

	// RemoveTx remove tx from tx resource
	RemoveTx(tx metadata.Transaction) error

	RemoveCandidateList([]string)
	//CheckTransactionFee
	// CheckTransactionFee(tx metadata.Transaction) (uint64, error)

	// Check tx validate by it self
	// ValidateTxByItSelf(tx metadata.Transaction) bool
}

type RewardAgent interface {
	GetBasicSalary(shardID byte) uint64
	GetSalaryPerTx(shardID byte) uint64
}
