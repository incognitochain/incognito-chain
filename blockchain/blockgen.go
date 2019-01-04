package blockchain

import (
	"time"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
)

type BlkTmplGenerator struct {
	// blockpool   BlockPool
	txPool            TxPool
	shardToBeaconPool ShardToBeaconPool
	chain             *BlockChain
	rewardAgent       RewardAgent
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

	//CheckTransactionFee
	// CheckTransactionFee(tx metadata.Transaction) (uint64, error)

	// Check tx validate by it self
	// ValidateTxByItSelf(tx metadata.Transaction) bool
}

type buyBackFromInfo struct {
	paymentAddress privacy.PaymentAddress
	buyBackPrice   uint64
	value          uint64
	requestedTxID  *common.Hash
}

type RewardAgent interface {
	GetBasicSalary(shardID byte) uint64
	GetSalaryPerTx(shardID byte) uint64
}

func (self BlkTmplGenerator) Init(txPool TxPool, chain *BlockChain, rewardAgent RewardAgent, shardToBeaconPool ShardToBeaconPool) (*BlkTmplGenerator, error) {
	return &BlkTmplGenerator{
		txPool:            txPool,
		shardToBeaconPool: shardToBeaconPool,
		chain:             chain,
		rewardAgent:       rewardAgent,
	}, nil
}
