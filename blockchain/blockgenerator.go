package blockchain

import (
	"github.com/incognitochain/incognito-chain/basemeta"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/common"
)

type BlockGenerator struct {
	// blockpool   BlockPool
	txPool      TxPool
	syncker     Syncker
	chain       *BlockChain
	CQuit       chan struct{}
	CPendingTxs <-chan basemeta.Transaction
	CRemovedTxs <-chan basemeta.Transaction
	PendingTxs  map[common.Hash]basemeta.Transaction
	mtx         sync.RWMutex
}

func NewBlockGenerator(txPool TxPool, chain *BlockChain, syncker Syncker, cPendingTxs chan basemeta.Transaction, cRemovedTxs chan basemeta.Transaction) (*BlockGenerator, error) {
	return &BlockGenerator{
		txPool:      txPool,
		syncker:     syncker,
		chain:       chain,
		PendingTxs:  make(map[common.Hash]basemeta.Transaction),
		CPendingTxs: cPendingTxs,
		CRemovedTxs: cRemovedTxs,
	}, nil
}

func (blockGenerator *BlockGenerator) Start(cQuit chan struct{}) {
	Logger.log.Critical("Block Gen is starting")
	for w := 0; w < WorkerNumber; w++ {
		go blockGenerator.AddTransactionV2Worker(blockGenerator.CPendingTxs)
	}
	for w := 0; w < WorkerNumber; w++ {
		go blockGenerator.RemoveTransactionV2Worker(blockGenerator.CRemovedTxs)
	}
	for {
		select {
		case <-cQuit:
			return
		}
	}
}
func (blockGenerator *BlockGenerator) AddTransactionV2(tx basemeta.Transaction) {
	blockGenerator.mtx.Lock()
	defer blockGenerator.mtx.Unlock()
	blockGenerator.PendingTxs[*tx.Hash()] = tx
}
func (blockGenerator *BlockGenerator) AddTransactionV2Worker(cPendingTx <-chan basemeta.Transaction) {
	for tx := range cPendingTx {
		blockGenerator.AddTransactionV2(tx)
		time.Sleep(time.Nanosecond)
	}
}
func (blockGenerator *BlockGenerator) RemoveTransactionV2(tx basemeta.Transaction) {
	blockGenerator.mtx.Lock()
	defer blockGenerator.mtx.Unlock()
	delete(blockGenerator.PendingTxs, *tx.Hash())
}
func (blockGenerator *BlockGenerator) RemoveTransactionV2Worker(cRemoveTx <-chan basemeta.Transaction) {
	for tx := range cRemoveTx {
		blockGenerator.RemoveTransactionV2(tx)
		time.Sleep(time.Nanosecond)
	}
}
func (blockGenerator *BlockGenerator) GetPendingTxsV2(shardID byte) []basemeta.Transaction {
	blockGenerator.mtx.Lock()
	defer blockGenerator.mtx.Unlock()
	pendingTxs := []basemeta.Transaction{}
	for _, tx := range blockGenerator.PendingTxs {
		txShardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
		if shardID != 255 && txShardID != shardID {
			continue
		}
		pendingTxs = append(pendingTxs, tx)
	}
	return pendingTxs
}
