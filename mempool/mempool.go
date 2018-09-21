package mempool

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ninjadotorg/cash-prototype/blockchain"
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/transaction"
)

// ID is Peer Ids, so that orphans can be identified by which peer first re-payed them.
type ID uint64

// Config is a descriptor containing the memory pool configuration.
type Config struct {
	// Policy defines the various mempool configuration options related
	// to policy.
	Policy Policy

	// Block chain of node
	BlockChain *blockchain.BlockChain

	ChainParams *blockchain.Params
}

// orphanTx is normal transaction that references an ancestor transaction
// that is not yet available.  It also contains additional information related
// to it such as an expiration time to help prevent caching the orphan forever.
type orphanTx struct {
	tx         *transaction.Tx
	id         ID
	expiration time.Time
}

// TxDesc is transaction description in mempool
type TxDesc struct {
	// transaction details
	Desc transaction.TxDesc

	// StartingPriority is the priority of the transaction when it was added
	// to the pool.
	StartingPriority float64
}

// TxPool is transaction pool
type TxPool struct {
	// The following variables must only be used atomically.
	lastUpdated int64 // last time pool was updated

	mtx    sync.RWMutex
	Config Config
	pool   map[common.Hash]*TxDesc
	//orphans       map[chainhash.Hash]*orphanTx
	//orphansByPrev map[wire.OutPoint]map[common.Hash]*Tx
	//outpoints     map[wire.OutPoint]*Tx
	//pennyTotal    float64 // exponentially decaying total for penny spends.
	//lastPennyUnix int64   // unix time of last ``penny spend''

	// nextExpireScan is the time after which the orphan pool will be
	// scanned in order to evict orphans.  This is NOT a hard deadline as
	// the scan will only run when an orphan is added to the pool as opposed
	// to on an unconditional timer.
	nextExpireScan time.Time
}

// check transaction in pool
func (tp *TxPool) isTxInPool(hash *common.Hash) bool {
	tp.mtx.Lock()
	defer tp.mtx.Unlock()
	if _, exists := tp.pool[*hash]; exists {
		return true
	}
	return false
}

// HaveTx check existed transaction
func (tp *TxPool) HaveTx(hash *common.Hash) bool {
	// Protect concurrent access.
	haveTx := tp.isTxInPool(hash)
	return haveTx
}

// add transaction into pool
func (tp *TxPool) addTx(tx transaction.Transaction, height int32, fee float64) *TxDesc {
	txD := &TxDesc{
		Desc: transaction.TxDesc{
			Tx:     tx,
			Added:  time.Now(),
			Height: height, //@todo we will apply calc function for height.
			Fee:    fee,    //@todo we will apply calc function for fee.
		},
		StartingPriority: 1, //@todo we will apply calc function for it.
	}
	log.Printf(tx.Hash().String())
	tp.mtx.Lock()
	tp.pool[*tx.Hash()] = txD
	tp.mtx.Unlock()
	atomic.StoreInt64(&tp.lastUpdated, time.Now().Unix())
	return txD
}

// CanAcceptTransaction validate transaction is valid and can add to pool
func (tp *TxPool) CanAcceptTransaction(tx transaction.Transaction) (*common.Hash, *TxDesc, error) {
	//@todo we will apply policy here
	// that make sure transaction is accepted when passed any rules
	txInfo := tx.(*transaction.Tx)
	chainID, err := tp.Config.Policy.Consensus.GetTxSenderChain(txInfo.AddressHash)
	if err != nil {
		return nil, nil, err
	}
	bestHeight := tp.Config.BlockChain.BestState[chainID].BestBlock.Height
	nextBlockHeight := bestHeight + 1
	// Perform several checks on the transaction inputs using the invariant
	// rules in blockchain for what transactions are allowed into blocks.
	// Also returns the fees associated with the transaction which will be
	// used later.
	txFee, err := tp.Config.BlockChain.CheckTransactionInputs(&tx, nextBlockHeight, nil, tp.Config.ChainParams)
	if err != nil {
		//if cerr, ok := err.(blockchain.RuleError); ok {
		//	return nil, nil, chainRuleError(cerr)
		//}
		return nil, nil, err
	}

	if tp.HaveTx(tx.Hash()) != true {
		txD := tp.addTx(tx, bestHeight, txFee)
		return tx.Hash(), txD, nil
	}
	return nil, nil, errors.New("Exist this tx in pool")
}

// remove transaction for pool
func (tp *TxPool) removeTx(tx transaction.Tx) {
	log.Printf(tx.Hash().String())
	tp.mtx.Lock()
	if _, exists := tp.pool[*tx.Hash()]; exists {
		delete(tp.pool, *tx.Hash())
		atomic.StoreInt64(&tp.lastUpdated, time.Now().Unix())
	}
	tp.mtx.Unlock()
}

// RemoveTx safe remove transaction for pool
func (tp *TxPool) RemoveTx(tx transaction.Tx) {
	tp.removeTx(tx)
}

// GetTx get transaction info by hash
func (tp *TxPool) GetTx(txHash *common.Hash) (transaction.Transaction, error) {
	tp.mtx.Lock()
	log.Println(txHash.String())
	txDesc, exists := tp.pool[*txHash]
	tp.mtx.Unlock()
	if exists {
		return txDesc.Desc.Tx, nil
	}

	return nil, fmt.Errorf("transaction is not in the pool")
}

// // MiningDescs returns a slice of mining descriptors for all the transactions
// // in the pool.
func (tp *TxPool) MiningDescs() []*transaction.TxDesc {
	descs := []*transaction.TxDesc{}
	tp.mtx.Lock()
	for _, desc := range tp.pool {
		descs = append(descs, &desc.Desc)
	}
	tp.mtx.Unlock()

	return descs
}

// Count return len of transaction pool
func (tp *TxPool) Count() int {
	count := len(tp.pool)
	return count
}

// Clear
func (tp *TxPool) Clear() {
	tp.pool = make(map[common.Hash]*TxDesc)
}

// New returns a new memory pool for validating and storing standalone
// transactions until they are mined into a block.
func New(cfg *Config) *TxPool {
	return &TxPool{
		Config:         *cfg,
		pool:           make(map[common.Hash]*TxDesc),
		nextExpireScan: time.Now().Add(orphanExpireScanInterval),
	}
}
