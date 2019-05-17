package blockchain

import (
	"sync"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/metadata"
)

type BlkTmplGenerator struct {
	// blockpool   BlockPool
	txPool            TxPool
	shardToBeaconPool ShardToBeaconPool
	crossShardPool    map[byte]CrossShardPool
	chain             *BlockChain
	CQuit             chan struct{}
	CPendingTxs       chan []metadata.Transaction
	CRemovedTxs       chan []metadata.Transaction
	PendingTxs        map[common.Hash]metadata.Transaction
	mtx               sync.RWMutex
}

func (blkTmplGenerator BlkTmplGenerator) Init(txPool TxPool, chain *BlockChain, shardToBeaconPool ShardToBeaconPool, crossShardPool map[byte]CrossShardPool, cPendingTxs chan []metadata.Transaction, cRemovedTxs chan []metadata.Transaction) (*BlkTmplGenerator, error) {
	return &BlkTmplGenerator{
		txPool:            txPool,
		shardToBeaconPool: shardToBeaconPool,
		crossShardPool:    crossShardPool,
		chain:             chain,
		PendingTxs:        make(map[common.Hash]metadata.Transaction),
		CPendingTxs:       cPendingTxs,
		CRemovedTxs:       cRemovedTxs,
	}, nil
}

func (blkTmplGenerator *BlkTmplGenerator) Start(cQuit chan struct{}) {
	for {
		select {
		case <-cQuit:
			return
		case addTxs := <-blkTmplGenerator.CPendingTxs:
			{
				go blkTmplGenerator.AddTransaction(addTxs)
			}
		case removeTxs := <-blkTmplGenerator.CRemovedTxs:
			{
				go blkTmplGenerator.RemoveTransaction(removeTxs)
			}
		}
	}
}
func (blkTmplGenerator *BlkTmplGenerator) AddTransaction(txs []metadata.Transaction) {
	blkTmplGenerator.mtx.Lock()
	defer blkTmplGenerator.mtx.Unlock()
	blkTmplGenerator.PendingTxs = make(map[common.Hash]metadata.Transaction)
	for _, tx := range txs {
		blkTmplGenerator.PendingTxs[*tx.Hash()] = tx
	}
}
func (blkTmplGenerator *BlkTmplGenerator) RemoveTransaction(txs []metadata.Transaction) {
	blkTmplGenerator.mtx.Lock()
	defer blkTmplGenerator.mtx.Unlock()
	//Notice: just reset
	blkTmplGenerator.PendingTxs = make(map[common.Hash]metadata.Transaction)
}
func (blkTmplGenerator *BlkTmplGenerator) GetPendingTxs() []metadata.Transaction {
	blkTmplGenerator.mtx.RLock()
	defer blkTmplGenerator.mtx.RUnlock()
	pendingTxs := []metadata.Transaction{}
	for _, tx := range blkTmplGenerator.PendingTxs{
		pendingTxs = append(pendingTxs, tx)
	}
	return pendingTxs
}
