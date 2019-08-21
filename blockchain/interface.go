package blockchain

import (
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
)

type ShardToBeaconPool interface {
	RemoveBlock(map[byte]uint64)
	//GetFinalBlock() map[byte][]ShardToBeaconBlock
	AddShardToBeaconBlock(*ShardToBeaconBlock) (uint64, uint64, error)
	//ValidateShardToBeaconBlock(ShardToBeaconBlock) error
	GetValidBlockHash() map[byte][]common.Hash
	GetValidBlock(map[byte]uint64) map[byte][]*ShardToBeaconBlock
	GetValidBlockHeight() map[byte][]uint64
	GetLatestValidPendingBlockHeight() map[byte]uint64
	GetBlockByHeight(shardID byte, height uint64) *ShardToBeaconBlock
	SetShardState(map[byte]uint64)
	GetAllBlockHeight() map[byte][]uint64
}

type CrossShardPool interface {
	AddCrossShardBlock(*CrossShardBlock) (map[byte]uint64, byte, error)
	GetValidBlock(map[byte]uint64) map[byte][]*CrossShardBlock
	GetLatestValidBlockHeight() map[byte]uint64
	GetValidBlockHeight() map[byte][]uint64
	GetBlockByHeight(_shardID byte, height uint64) *CrossShardBlock
	RemoveBlockByHeight(map[byte]uint64)
	UpdatePool() map[byte]uint64
	GetAllBlockHeight() map[byte][]uint64
}

type ShardPool interface {
	RemoveBlock(uint64)
	AddShardBlock(block *ShardBlock) error
	GetValidBlockHash() []common.Hash
	GetValidBlock() []*ShardBlock
	GetValidBlockHeight() []uint64
	GetLatestValidBlockHeight() uint64
	SetShardState(uint64)
	GetAllBlockHeight() []uint64
	Start(chan struct{})
}

type BeaconPool interface {
	RemoveBlock(uint64)
	AddBeaconBlock(block *BeaconBlock) error
	GetValidBlock() []*BeaconBlock
	GetValidBlockHeight() []uint64
	SetBeaconState(uint64)
	GetAllBlockHeight() []uint64
	Start(chan struct{})
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
	RemoveTx(txs []metadata.Transaction, isInBlock bool)

	RemoveCandidateList([]string)

	RemoveTokenIDList([]string)

	EmptyPool() bool

	MaybeAcceptTransactionForBlockProducing(metadata.Transaction) (*metadata.TxDesc, error)
	ValidateTxList(txs []metadata.Transaction) error
	//CheckTransactionFee
	// CheckTransactionFee(tx metadata.Transaction) (uint64, error)

	// Check tx validate by it self
	// ValidateTxByItSelf(tx metadata.Transaction) bool
}

type FeeEstimator interface {
	RegisterBlock(block *ShardBlock) error
}

// type ConsensusEngineInterface interface {
// 	IsOngoing(chainkey string) bool

// 	GetMiningPublicKey() (publickey string, keyType string)
// 	SignDataWithMiningKey(data []byte) (string, error)

// 	ValidateProducerPosition(block BlockInterface, chain ChainInterface) error
// 	ValidateProducerSig(block BlockInterface, chain ChainInterface) error
// 	ValidateCommitteeSig(block BlockInterface, chain ChainInterface) error

// 	VerifyData(data []byte, sig string, publicKey string, consensusType string) error

// 	SwitchConsensus(chainkey string, consensus string) error
// }

// type ConsensusInterface interface {
// 	NewInstance() ConsensusInterface
// 	GetConsensusName() string

// 	ValidateBlock(block BlockInterface) error
// 	IsOngoing() bool
// 	ValidateProducerPosition(block BlockInterface) error
// 	ValidateProducerSig(block BlockInterface) error
// 	ValidateCommitteeSig(block BlockInterface) error
// }

// type BlockInterface interface {
// 	GetHeight() uint64
// 	Hash() *common.Hash
// 	AddValidationField(validateData string) error
// 	GetValidationField() string
// 	GetRound() int
// 	GetRoundKey() string
// }

type ChainInterface interface {
	GetChainName() string
	GetConsensusType() string
	GetLastBlockTimeStamp() int64
	GetMinBlkInterval() time.Duration
	GetMaxBlkCreateTime() time.Duration
	IsReady() bool
	GetActiveShardNumber() int

	GetPubkeyRole(pubkey string, round int) (string, byte)
	CurrentHeight() uint64
	GetCommitteeSize() int
	GetCommittee() []string
	GetPubKeyCommitteeIndex(string) int
	GetLastProposerIndex() int

	CreateNewBlock(round int) common.BlockInterface
	InsertBlk(common.BlockInterface, bool)
	ValidateBlock(common.BlockInterface) error
	ValidateBlockSanity(common.BlockInterface) error
	ValidateBlockWithBlockChain(common.BlockInterface) error
	GetShardID() int
}

type BestStateInterface interface {
	GetLastBlockTimeStamp() uint64
	GetBlkMinInterval() time.Duration
	GetBlkMaxCreateTime() time.Duration
	CurrentHeight() uint64
	GetCommittee() []string
	GetLastProposerIdx() int
}
