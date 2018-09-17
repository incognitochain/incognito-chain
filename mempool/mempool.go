package mempool

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ninjadotorg/cash-prototype/blockchain"
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/mining"
	"github.com/ninjadotorg/cash-prototype/privacy/client"
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

	// that make sure transaction is accepted when passed any rules
	bestHeight := tp.config.BlockChain.BestState.BestBlock.Height

	// Check tx with policy
	// check version
	ok := tp.config.Policy.CheckTxVersion(&tx)
	if !ok {
		err := TxRuleError{}
		err.Init(RejectVersion, fmt.Sprintf("%v's version is invalid", txHash.String()))
		return nil, nil, err
	}

	// check fee of tx
	txFee, err := tp.CheckTransactionFee(tx)
	if err != nil {
		return nil, nil, err
	}
	// end check with policy

	// Don't accept the transaction if it already exists in the pool.
	if tp.isTxInPool(txHash) {
		str := fmt.Sprintf("already have transaction %v", txHash.String())
		err := TxRuleError{}
		err.Init(RejectDuplicateTx, str)
		return nil, nil, err
	}

	// A standalone transaction must not be a coinbase transaction.
	if blockchain.IsCoinBaseTx(tx) {
		err := TxRuleError{}
		err.Init(RejectCoinbaseTx, fmt.Sprintf("%v is coinbase tx", txHash.String()))
		return nil, nil, err
	}

	// sanity data
	if validate, errS := tp.ValidateSanityData(tx); !validate {
		err := TxRuleError{}
		err.Init(RejectSansityTx, fmt.Sprintf("transaction's sansity %v is error %v", txHash.String(), errS.Error()))
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

/**
CheckTransactionFee - check fee of tx
*/
func (tp *TxPool) CheckTransactionFee(tx transaction.Transaction) (uint64, error) {
	// Coinbase transactions have no inputs.
	if blockchain.IsCoinBaseTx(tx) {
		return 0, nil
	}

	txType := tx.GetType()
	switch txType {
	case common.TxNormalType:
		{
			normalTx := tx.(*transaction.Tx)
			err := tp.config.Policy.CheckTransactionFee(normalTx)
			return normalTx.Fee, err
		}
	case common.TxActionParamsType:
		{
			return 0, nil
		}
	default:
		{
			return 0, errors.New("Wrong tx type")
		}
	}
}

/**
ValidateSanityData - validate sansity data of tx
*/
func (tp *TxPool) ValidateSanityData(tx transaction.Transaction) (bool, error) {
	if tx.GetType() == common.TxNormalType {
		txN := tx.(*transaction.Tx)
		//check version
		if txN.Version > transaction.TxVersion {
			return false, errors.New("Wrong version")
		}
		// check LockTime before now
		if int64(txN.LockTime) > time.Now().Unix() {
			return false, errors.New("Wrong locktime")
		}
		// check Type equal "n"
		if txN.Type != common.TxNormalType {
			return false, errors.New("Wrong type")
		}
		// check length of JSPubKey
		if len(txN.JSPubKey) != 32 {
			return false, errors.New("Wrong jspubkey")
		}
		// check length of JSSig
		if len(txN.JSSig) != 64 {
			return false, errors.New("Wrong jssig")
		}
		//check Descs

		// get list nullifiers from db to check spending
		txViewPointTxOutBond, err := tp.config.BlockChain.FetchTxViewPoint(common.TxOutBondType)
		if err != nil {
			return false, errors.New("Wrong nultifier")
		}
		nullifiersInDbTxOutBond := txViewPointTxOutBond.ListNullifiers(common.TxOutBondType)

		txViewPointTxOutCoin, err := tp.config.BlockChain.FetchTxViewPoint(common.TxOutCoinType)
		if err != nil {
			return false, errors.New("Wrong nultifier")
		}
		nullifiersInDbTxOutCoin := txViewPointTxOutCoin.ListNullifiers(common.TxOutCoinType)

		for _, desc := range txN.Descs {
			// check length of Anchor
			if len(desc.Anchor) != 32 {
				return false, errors.New("Wrong anchor")
			}
			// check length of EphemeralPubKey
			if len(desc.EphemeralPubKey) != client.EphemeralKeyLength {
				return false, errors.New("Wrong ephemeralpubkey")
			}
			// check length of HSigSeed
			if len(desc.HSigSeed) != 32 {
				return false, errors.New("Wrong hsigseed")
			}
			// check value of Type
			if desc.Type != common.TxOutBondType || desc.Type != common.TxOutCoinType {
				return false, errors.New("Wrong type")
			}
			// check length of Nullifiers
			if len(desc.Nullifiers) != 2 {
				return false, errors.New("Wrong nullifiers")
			}
			if len(desc.Nullifiers[0]) != 32 {
				return false, errors.New("Wrong nullifiers")
			}
			if len(desc.Nullifiers[1]) != 32 {
				return false, errors.New("Wrong nullifiers")
			}
			// check length of Commitments
			if len(desc.Commitments) != 2 {
				return false, errors.New("Wrong commitments")
			}
			if len(desc.Commitments[0]) != 32 {
				return false, errors.New("Wrong commitments")
			}
			if len(desc.Commitments[1]) != 32 {
				return false, errors.New("Wrong commitments")
			}
			// check length of Vmacs
			if len(desc.Vmacs) != 2 {
				return false, errors.New("Wrong vmacs")
			}
			if len(desc.Vmacs[0]) != 32 {
				return false, errors.New("Wrong vmacs")
			}
			if len(desc.Vmacs[1]) != 32 {
				return false, errors.New("Wrong vmacs")
			}
			//
			if desc.Proof == nil {
				return false, errors.New("Wrong proof")
			}
			// check length of Proof
			if len(desc.Proof.G_A) != 33 ||
				len(desc.Proof.G_APrime) != 33 ||
				len(desc.Proof.G_B) != 33 ||
				len(desc.Proof.G_BPrime) != 65 ||
				len(desc.Proof.G_C) != 33 ||
				len(desc.Proof.G_CPrime) != 33 ||
				len(desc.Proof.G_K) != 33 ||
				len(desc.Proof.G_H) != 33 {
				return false, errors.New("Wrong proof")
			}
			//
			if len(desc.EncryptedData) != 2 {
				return false, errors.New("Wrong encryptedData")
			}
			// check nulltifier is existed in DB
			if desc.Type == common.TxOutBondType {
				checkCandiateNullifier, err := common.SliceExists(nullifiersInDbTxOutBond, desc.Nullifiers[0])
				if err != nil || checkCandiateNullifier == true {
					// candidate nullifier is existed in db
					return false, errors.New("Wrong nullifier")
				}
				checkCandiateNullifier, err = common.SliceExists(nullifiersInDbTxOutBond, desc.Nullifiers[1])
				if err != nil || checkCandiateNullifier == true {
					// candidate nullifier is existed in db
					return false, errors.New("Wrong nullifier")
				}
			}
			if desc.Type == common.TxOutBondType {
				checkCandiateNullifier, err := common.SliceExists(nullifiersInDbTxOutCoin, desc.Nullifiers[0])
				if err != nil || checkCandiateNullifier == true {
					// candidate nullifier is existed in db
					return false, errors.New("Wrong nullifier")
				}
				checkCandiateNullifier, err = common.SliceExists(nullifiersInDbTxOutCoin, desc.Nullifiers[1])
				if err != nil || checkCandiateNullifier == true {
					// candidate nullifier is existed in db
					return false, errors.New("Wrong nullifier")
				}
			}
			if desc.Reward != 0 {
				return false, errors.New("Wrong reward")
			}
		}
	} else if tx.GetType() == common.TxActionParamsType {
		txA := tx.(*transaction.ActionParamTx)
		// check Version
		if txA.Version > transaction.TxVersion {
			return false, errors.New("Wrong version")
		}
		// check LockTime before now
		if int64(txA.LockTime) > time.Now().Unix() {
			return false, errors.New("Wrong lockTime")
		}
		// check Type equal "a"
		if txA.Type != common.TxActionParamsType {
			return false, errors.New("Wrong type")
		}
	} else {
		return false, errors.New("Wrong type")
	}

	return true, nil
}
