package blockchain

import (
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/ethrelaying/les"
	"github.com/incognitochain/incognito-chain/metadata"
)

type BlkTmplGenerator struct {
	// blockpool   BlockPool
	txPool            TxPool
	shardToBeaconPool ShardToBeaconPool
	crossShardPool    map[byte]CrossShardPool
	chain             *BlockChain
	CQuit             chan struct{}
	CPendingTxs       <-chan metadata.Transaction
	CRemovedTxs       <-chan metadata.Transaction
	PendingTxs        map[common.Hash]metadata.Transaction
	mtx               sync.RWMutex
}

func (blkTmplGenerator BlkTmplGenerator) Init(txPool TxPool, chain *BlockChain, shardToBeaconPool ShardToBeaconPool, crossShardPool map[byte]CrossShardPool, cPendingTxs chan metadata.Transaction, cRemovedTxs chan metadata.Transaction) (*BlkTmplGenerator, error) {
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
	Logger.log.Critical("Block Gen is starting")
	for w := 0; w < workerNum; w++ {
		go blkTmplGenerator.AddTransactionV2Worker(blkTmplGenerator.CPendingTxs)
	}
	for w := 0; w < workerNum; w++ {
		go blkTmplGenerator.RemoveTransactionV2Worker(blkTmplGenerator.CRemovedTxs)
	}
	for {
		select {
		case <-cQuit:
			return
		//case addTxs := <-blkTmplGenerator.CPendingTxs:
		//	{
		//		go blkTmplGenerator.AddTransactionV2(addTxs)
		//	}
		//case removeTxs := <-blkTmplGenerator.CRemovedTxs:
		//	{
		//		go blkTmplGenerator.RemoveTransactionV2(removeTxs)
		//	}
		}
	}
}

//func (blkTmplGenerator *BlkTmplGenerator) AddTransaction(txs []metadata.Transaction) {
//	blkTmplGenerator.mtx.Lock()
//	defer blkTmplGenerator.mtx.Unlock()
//	Logger.log.Critical("Number of transaction get from pool: ", len(txs))
//	blkTmplGenerator.PendingTxs = make(map[common.Hash]metadata.Transaction)
//	for _, tx := range txs {
//		blkTmplGenerator.PendingTxs[*tx.Hash()] = tx
//	}
//}
//func (blkTmplGenerator *BlkTmplGenerator) RemoveTransaction(txs []metadata.Transaction) {
//	blkTmplGenerator.mtx.Lock()
//	defer blkTmplGenerator.mtx.Unlock()
//	//Notice: just reset
//	blkTmplGenerator.PendingTxs = make(map[common.Hash]metadata.Transaction)
//}
//func (blkTmplGenerator *BlkTmplGenerator) GetPendingTxs() []metadata.Transaction {
//	blkTmplGenerator.mtx.Lock()
//	defer blkTmplGenerator.mtx.Unlock()
//	pendingTxs := []metadata.Transaction{}
//	for _, tx := range blkTmplGenerator.PendingTxs{
//		pendingTxs = append(pendingTxs, tx)
//	}
//	blkTmplGenerator.PendingTxs = make(map[common.Hash]metadata.Transaction)
//	return pendingTxs
//}

func (blkTmplGenerator *BlkTmplGenerator) AddTransactionV2(tx metadata.Transaction) {
	blkTmplGenerator.mtx.Lock()
	defer blkTmplGenerator.mtx.Unlock()
	blkTmplGenerator.PendingTxs[*tx.Hash()] = tx
}
func (blkTmplGenerator *BlkTmplGenerator) AddTransactionV2Worker(cPendingTx <-chan metadata.Transaction) {
	for tx := range cPendingTx {
		blkTmplGenerator.AddTransactionV2(tx)
		time.Sleep(time.Nanosecond)
	}
}
func (blkTmplGenerator *BlkTmplGenerator) RemoveTransactionV2(tx metadata.Transaction) {
	blkTmplGenerator.mtx.Lock()
	defer blkTmplGenerator.mtx.Unlock()
	delete(blkTmplGenerator.PendingTxs, *tx.Hash())
}
func (blkTmplGenerator *BlkTmplGenerator) RemoveTransactionV2Worker(cRemoveTx <-chan metadata.Transaction) {
	for tx := range cRemoveTx {
		blkTmplGenerator.RemoveTransactionV2(tx)
		time.Sleep(time.Nanosecond)
	}
}
func (blkTmplGenerator *BlkTmplGenerator) GetPendingTxsV2() []metadata.Transaction {
	blkTmplGenerator.mtx.Lock()
	defer blkTmplGenerator.mtx.Unlock()
	pendingTxs := []metadata.Transaction{}
	for _, tx := range blkTmplGenerator.PendingTxs {
		pendingTxs = append(pendingTxs, tx)
	}
	return pendingTxs
}

func (blkTmplGenerator *BlkTmplGenerator) SetLightETHToChain(leth *les.LightEthereum) {
	blkTmplGenerator.chain.LightEthereum = leth
}
