package mempool

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/mining"
	"github.com/ninjadotorg/cash-prototype/transaction"
	"github.com/ninjadotorg/cash-prototype/blockchain"
	"github.com/ninjadotorg/cash-prototype/privacy/client"
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
/*type orphanTx struct {
	tx         *transaction.Tx
	id         ID
	expiration time.Time
}*/

// TxDesc is transaction description in mempool
type TxDesc struct {
	// transaction details
	Desc mining.TxDesc

	// StartingPriority is the priority of the transaction when it was added
	// to the pool.
	StartingPriority float64
}

// TxPool is transaction pool
type TxPool struct {
	// The following variables must only be used atomically.
	lastUpdated int64 // last time pool was updated

	mtx    sync.RWMutex
	config Config
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

/**
Init Txpool from config
*/
func (tp *TxPool) Init(cfg *Config) {
	tp.config = *cfg
	tp.pool = make(map[common.Hash]*TxDesc)
	tp.nextExpireScan = time.Now().Add(orphanExpireScanInterval)
}

// check transaction in pool
func (tp *TxPool) isTxInPool(hash *common.Hash) bool {
	if _, exists := tp.pool[*hash]; exists {
		return true
	}

	return false
}

/**
// add transaction into pool
*/
func (tp *TxPool) addTx(tx transaction.Transaction, height int32, fee uint64) *TxDesc {
	txD := &TxDesc{
		Desc: mining.TxDesc{
			Tx:     tx,
			Added:  time.Now(),
			Height: height,
			Fee:    fee,
		},
		StartingPriority: 1, //@todo we will apply calc function for it.
	}
	Logger.log.Infof(tx.Hash().String())
	tp.pool[*tx.Hash()] = txD
	atomic.StoreInt64(&tp.lastUpdated, time.Now().Unix())
	return txD
}

// MaybeAcceptTransaction is the main workhorse for handling insertion of new
// free-standing transactions into a memory pool.  It includes functionality
// such as rejecting duplicate transactions, ensuring transactions follow all
// rules, detecting orphan transactions, and insertion into the memory pool.
//
// If the transaction is an orphan (missing parent transactions), the
// transaction is NOT added to the orphan pool, but each unknown referenced
// parent is returned.  Use ProcessTransaction instead if new orphans should
// be added to the orphan pool.
//
// This function is safe for concurrent access.
func (tp *TxPool) MaybeAcceptTransaction(tx transaction.Transaction) (*common.Hash, *TxDesc, error) {
	tp.mtx.Lock()
	hash, txDesc, err := tp.maybeAcceptTransaction(tx)
	tp.mtx.Unlock()
	return hash, txDesc, err
}

/**
// maybeAcceptTransaction is the internal function which implements the public
// MaybeAcceptTransaction.  See the comment for MaybeAcceptTransaction for
// more details.
//
// This function MUST be called with the mempool lock held (for writes).
*/
func (tp *TxPool) maybeAcceptTransaction(tx transaction.Transaction) (*common.Hash, *TxDesc, error) {
	txHash := tx.Hash()

	//@todo we will apply policy here
	// that make sure transaction is accepted when passed any rules
	bestHeight := tp.config.BlockChain.BestState.BestBlock.Height
	nextBlockHeight := bestHeight + 1

	// Perform several checks on the transaction data using the invariant
	// rules in blockchain for what transactions are allowed into blocks.
	// Also returns the fees associated with the transaction which will be
	// used later.
	txFee, err := tp.config.BlockChain.CheckTransactionData(tx, nextBlockHeight, nil, tp.config.ChainParams)
	if err != nil {
		//if cerr, ok := err.(blockchain.RuleError); ok {
		//	return nil, nil, chainRuleError(cerr)
		//}
		return nil, nil, err
	}

	// Don't accept the transaction if it already exists in the pool.
	if tp.isTxInPool(txHash) {
		str := fmt.Sprintf("already have transaction %v", txHash.String())
		err := TxRuleError{}
		err.Init(RejectDuplicateTx, str)
		return nil, nil, err
	}

	if blockchain.IsCoinBaseTx(tx) {
		err := TxRuleError{}
		err.Init(RejectCoinbaseTx, fmt.Sprintf("%v is coinbase tx", txHash.String()))
		return nil, nil, err
	}

	// sanity data
	// TODO
	if !tp.ValidateSansityData(tx) {
		err := TxRuleError{}.Init(SansityInvalidTx, fmt.Sprintf("transaction's sansity %v is invalid", txHash.String()))
		return nil, nil, err
	}

	// Validate tx by it self
	validate := tx.ValidateTransaction()
	if !validate {
		err := TxRuleError{}
		err.Init(RejectInvalidTx, "Invalid tx")
		return nil, nil, err
	}

	txD := tp.addTx(tx, bestHeight, txFee)
	return tx.Hash(), txD, nil
}

// remove transaction for pool
func (tp *TxPool) removeTx(tx *transaction.Transaction) error {
	Logger.log.Infof((*tx).Hash().String())
	if _, exists := tp.pool[*(*tx).Hash()]; exists {
		delete(tp.pool, *(*tx).Hash())
		atomic.StoreInt64(&tp.lastUpdated, time.Now().Unix())
		return nil
	} else {
		return errors.New("Not exist tx in pool")
	}
}

// RemoveTx safe remove transaction for pool
func (tp *TxPool) RemoveTx(tx *transaction.Transaction) error {
	tp.mtx.Lock()
	err := tp.removeTx(tx)
	tp.mtx.Unlock()
	return err
}

// GetTx get transaction info by hash
func (tp *TxPool) GetTx(txHash *common.Hash) (transaction.Transaction, error) {
	tp.mtx.Lock()
	Logger.log.Info(txHash.String())
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

// Count return len of transaction pool
func (tp *TxPool) Count() int {
	tp.mtx.RLock()
	count := len(tp.pool)
	tp.mtx.RUnlock()

	return count
}

/**
// LastUpdated returns the last time a transaction was added to or
	// removed from the source pool.
 */
func (tp *TxPool) LastUpdated() time.Time {
	return time.Unix(tp.lastUpdated, 0)
}

/**
// HaveTransaction returns whether or not the passed transaction hash
	// exists in the source pool.
 */
func (tp *TxPool) HaveTransaction(hash *common.Hash) bool {
	// Protect concurrent access.
	tp.mtx.RLock()
	haveTx := tp.isTxInPool(hash)
	tp.mtx.RUnlock()

	return haveTx
}

func (tp *TxPool) ValidateSansityData(tx transaction.Transaction) bool {
	if tx.GetType() == common.TxNormalType {
		txN := tx.(*transaction.Tx)
		//check version
		if txN.Version > transaction.TxVersion {
			return false
		}
		//check locktime
		if int64(txN.LockTime) > time.Now().Unix() {
			return false
		}

		if len(txN.JSPubKey) != 32 {
			return false
		}
		if len(txN.JSSig) != 64 {
			return false
		}
		//check Descs
		for _, desc := range txN.Descs {
			if len(desc.Anchor) != 32 {
				return false
			}
			if len(desc.EphemeralPubKey) != client.EphemeralKeyLength {
				return false
			}
			if len(desc.HSigSeed) != 32 {
				return false
			}
			if desc.Type != common.TxOutBondType || desc.Type != common.TxOutCoinType {
				return false
			}
		}

		_ = txN
	} else if tx.GetType() == common.TxActionParamsType {
		txA := tx.(*transaction.ActionParamTx)

		_ = txA
	} else {
		return false
	}

	return true
}
