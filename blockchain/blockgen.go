package blockchain

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy"
)

type BlkTmplGenerator struct {
	// blockpool   BlockPool
	txPool            TxPool
	shardToBeaconPool ShardToBeaconPool
	crossShardPool    CrossShardPool
	chain             *BlockChain
	rewardAgent       RewardAgent
}

type buyBackFromInfo struct {
	paymentAddress privacy.PaymentAddress
	buyBackPrice   uint64
	value          uint64
	requestedTxID  *common.Hash
}

func (self BlkTmplGenerator) Init(txPool TxPool, chain *BlockChain, rewardAgent RewardAgent, shardToBeaconPool ShardToBeaconPool, crossShardPool CrossShardPool) (*BlkTmplGenerator, error) {
	return &BlkTmplGenerator{
		txPool:            txPool,
		shardToBeaconPool: shardToBeaconPool,
		crossShardPool:    crossShardPool,
		chain:             chain,
		rewardAgent:       rewardAgent,
	}, nil
}
