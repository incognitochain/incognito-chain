package blockchain

type BlkTmplGenerator struct {
	// blockpool   BlockPool
	txPool            TxPool
	shardToBeaconPool ShardToBeaconPool
	crossShardPool    map[byte]CrossShardPool
	chain             *BlockChain
}

func (blkTmplGenerator BlkTmplGenerator) Init(txPool TxPool, chain *BlockChain, shardToBeaconPool ShardToBeaconPool, crossShardPool map[byte]CrossShardPool) (*BlkTmplGenerator, error) {
	return &BlkTmplGenerator{
		txPool:            txPool,
		shardToBeaconPool: shardToBeaconPool,
		crossShardPool:    crossShardPool,
		chain:             chain,
	}, nil
}
