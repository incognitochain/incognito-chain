package mempool

import (
	"sync/atomic"
	"time"

	"fmt"
	"github.com/internet-cash/prototype/common"
	"github.com/internet-cash/prototype/transaction"
	"sync"
	"github.com/internet-cash/prototype/mining"
	"log"
)

//Peer Ids, so that orphans can be identified by which peer first repayed them.
type Id uint64

// Config is a descriptor containing the memory pool configuration.
type Config struct {
	// Policy defines the various mempool configuration options related
	// to policy.
	Policy Policy
}

// orphanTx is normal transaction that references an ancestor transaction
// that is not yet available.  It also contains additional information related
// to it such as an expiration time to help prevent caching the orphan forever.
type orphanTx struct {
	tx         *transaction.Tx
	id         Id
	expiration time.Time
}

type TxDesc struct {
	//tracsaction details
	Desc mining.TxDesc

	// StartingPriority is the priority of the transaction when it was added
	// to the pool.
	StartingPriority float64
}

type TxPool struct {
	// The following variables must only be used atomically.
	lastUpdated int64 // last time pool was updated

	mtx  sync.RWMutex
	cfg  Config
	pool map[common.Hash]*TxDesc
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
//check transaction in pool
func (mp *TxPool) isTxInPool(hash *common.Hash) bool {
	if _, exists := mp.pool[*hash]; exists {
		return true
	}

	return false
}

//check existed transaction
func (tp *TxPool) HaveTx(hash *common.Hash) bool {
	// Protect concurrent access.
	tp.mtx.RLock()
	haveTx := tp.isTxInPool(hash)
	tp.mtx.RUnlock()

	return haveTx
}
//add transaction into pool
func (tp *TxPool) addTx(tx transaction.Transaction) *TxDesc {
	txD := &TxDesc{
		Desc: mining.TxDesc{
			Tx:     tx,
			Added:  time.Now(),
			Height: 1,   //@todo we will apply calc function for height.
			Fee:    200, //@todo we will apply calc function for fee.
		},
		StartingPriority: 1, //@todo we will apply calc function for it.
	}
	tp.pool[*tx.Hash()] = txD
	atomic.StoreInt64(&tp.lastUpdated, time.Now().Unix())
	return txD
}

func (tp *TxPool) CanAcceptTransaction(tx transaction.Transaction) (*common.Hash, *TxDesc, error) {
	//@todo we will apply policy here
	// that make sure transaction is accepted when passed any rules
	txD := tp.addTx(tx)
	return tx.Hash(), txD, nil
}

//remove transaction for pool
func (tp *TxPool) removeTx(tx transaction.Tx) {
	if _, exists := tp.pool[*tx.Hash()]; exists {
		delete(tp.pool, *tx.Hash())
		atomic.StoreInt64(&tp.lastUpdated, time.Now().Unix())
	}
}

//safe remove transaction for pool
func (tp *TxPool) RemoveTx(tx transaction.Tx) {
	tp.mtx.Lock()
	tp.removeTx(tx)
	tp.mtx.Unlock()
}

//this function is safe for access concurrent access.
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


// MiningDescs returns a slice of mining descriptors for all the transactions
// in the pool.
func (tp *TxPool) MiningDescs() []*mining.TxDesc {
	tp.mtx.RLock()
	descs := make([]*mining.TxDesc, len(tp.pool))
	i := 0
	for _, desc := range tp.pool {
		descs[i] = &desc.Desc
		i++
	}
	tp.mtx.RUnlock()

	return descs
}


// return len of transaction pool
func (tp *TxPool) Count() int {
	tp.mtx.RLock()
	count := len(tp.pool)
	tp.mtx.RUnlock()

	return count
}

// New returns a new memory pool for validating and storing standalone
// transactions until they are mined into a block.
func New(cfg *Config) *TxPool {
	return &TxPool{
		cfg:            *cfg,
		pool:           make(map[common.Hash]*TxDesc),
		nextExpireScan: time.Now().Add(orphanExpireScanInterval),
	}
}
