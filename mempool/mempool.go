package mempool

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy/client"
	"github.com/ninjadotorg/constant/transaction"
)

// config is a descriptor containing the memory pool configuration.
type Config struct {
	// Policy defines the various mempool configuration options related
	// to policy.
	Policy Policy

	// Block chain of node
	BlockChain *blockchain.BlockChain

	ChainParams *blockchain.Params

	// FeeEstimatator provides a feeEstimator. If it is not nil, the mempool
	// records all new transactions it observes into the feeEstimator.
	FeeEstimator map[byte]*FeeEstimator
}

// TxDesc is transaction description in mempool
type TxDesc struct {
	// transaction details
	Desc transaction.TxDesc

	StartingPriority int
}

// TxPool is transaction pool
type TxPool struct {
	// The following variables must only be used atomically.
	lastUpdated int64 // last time pool was updated

	mtx    sync.RWMutex
	config Config
	pool   map[common.Hash]*TxDesc
}

/*
Init Txpool from config
*/
func (tp *TxPool) Init(cfg *Config) {
	tp.config = *cfg
	tp.pool = make(map[common.Hash]*TxDesc)
}

// check transaction in pool
func (tp *TxPool) isTxInPool(hash *common.Hash) bool {
	if _, exists := tp.pool[*hash]; exists {
		return true
	}
	return false
}

/*
// add transaction into pool
*/
func (tp *TxPool) addTx(tx transaction.Transaction, height int32, fee uint64) *TxDesc {
	txD := &TxDesc{
		Desc: transaction.TxDesc{
			Tx:     tx,
			Added:  time.Now(),
			Height: height,
			Fee:    fee,
		},
		StartingPriority: 1, //@todo we will apply calc function for it.
	}
	log.Printf(tx.Hash().String())
	tp.pool[*tx.Hash()] = txD
	atomic.StoreInt64(&tp.lastUpdated, time.Now().Unix())

	// Record this tx for fee estimation if enabled. only apply for normal tx
	if tx.GetType() == common.TxNormalType {
		if tp.config.FeeEstimator != nil {
			chainId, err := common.GetTxSenderChain(tx.(*transaction.Tx).AddressLastByte)
			if err == nil {
				tp.config.FeeEstimator[chainId].ObserveTransaction(txD)
			} else {
				Logger.log.Error(err)
			}
		}
	} else if tx.GetType() == common.TxVotingType {
		if tp.config.FeeEstimator != nil {
			chainId, err := common.GetTxSenderChain(tx.(*transaction.TxVoting).AddressLastByte)
			if err == nil {
				tp.config.FeeEstimator[chainId].ObserveTransaction(txD)
			} else {
				Logger.log.Error(err)
			}
		}
	}

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

/*
// maybeAcceptTransaction is the internal function which implements the public
// MaybeAcceptTransaction.  See the comment for MaybeAcceptTransaction for
// more details.
//
// This function MUST be called with the mempool lock held (for writes).
*/
func (tp *TxPool) maybeAcceptTransaction(tx transaction.Transaction) (*common.Hash, *TxDesc, error) {
	txHash := tx.Hash()

	// that make sure transaction is accepted when passed any rules
	var chainID byte
	var err error

	switch tx.(type) {
	case *transaction.Tx:
		log.Println("Normal tx")
		txInfo := tx.(*transaction.Tx)
		chainID, err = common.GetTxSenderChain(txInfo.AddressLastByte)
	case *transaction.TxVoting:
		log.Println("Tx voting")
		txInfo := tx.(*transaction.TxVoting)
		chainID, err = common.GetTxSenderChain(txInfo.AddressLastByte)
	case *transaction.TxCustomToken:
		log.Println("Tx custom token")
		txInfo := tx.(*transaction.TxCustomToken)
		chainID, err = common.GetTxSenderChain(txInfo.AddressLastByte)
	}

	if err != nil {
		return nil, nil, err
	}
	bestHeight := tp.config.BlockChain.BestState[chainID].BestBlock.Header.Height
	// nextBlockHeight := bestHeight + 1
	// Check tx with policy
	// check version
	ok := tp.config.Policy.CheckTxVersion(&tx)
	if !ok {
		err := MempoolTxError{}
		err.Init(RejectVersion, errors.New(fmt.Sprintf("%+v's version is invalid", txHash.String())))
		return nil, nil, err
	}

	// check fee of tx
	txFee, err := tp.CheckTransactionFee(tx)
	if err != nil {
		return nil, nil, err
	}
	// end check with policy

	// validate double spend for : normal tx, voting tx
	if tx.GetType() == common.TxNormalType || tx.GetType() == common.TxVotingType {
		txViewPoint, err := tp.config.BlockChain.FetchTxViewPoint(chainID)
		if err != nil {
			str := fmt.Sprintf("Can not check double spend for tx")
			err := MempoolTxError{}
			err.Init(CanNotCheckDoubleSpend, errors.New(str))
			return nil, nil, err
		}
		nullifierDb := txViewPoint.ListNullifiers()
		var descs []*transaction.JoinSplitDesc
		if tx.GetType() == common.TxNormalType {
			descs = tx.(*transaction.Tx).Descs
		} else if tx.GetType() == common.TxVotingType {
			descs = tx.(*transaction.TxVoting).Descs
		}
		for _, desc := range descs {
			for _, nullifer := range desc.Nullifiers {
				existed, err := common.SliceBytesExists(nullifierDb, nullifer)
				if err != nil {
					str := fmt.Sprintf("Can not check double spend for tx")
					err := MempoolTxError{}
					err.Init(CanNotCheckDoubleSpend, errors.New(str))
					return nil, nil, err
				}
				if existed {
					str := fmt.Sprintf("Nullifiers of transaction %+v already existed", txHash.String())
					err := MempoolTxError{}
					err.Init(RejectDuplicateTx, errors.New(str))
					return nil, nil, err
				}
			}
		}
	}

	// Don't accept the transaction if it already exists in the pool.
	if tp.isTxInPool(txHash) {
		str := fmt.Sprintf("already have transaction %+v", txHash.String())
		err := MempoolTxError{}
		err.Init(RejectDuplicateTx, errors.New(str))
		return nil, nil, err
	}

	// A standalone transaction must not be a salary transaction.
	if blockchain.IsSalaryTx(tx) {
		err := MempoolTxError{}
		err.Init(RejectSalaryTx, errors.New(fmt.Sprintf("%+v is salary tx", txHash.String())))
		return nil, nil, err
	}

	// sanity data
	if validate, errS := tp.ValidateSanityData(tx); !validate {
		err := MempoolTxError{}
		err.Init(RejectSansityTx, errors.New(fmt.Sprintf("transaction's sansity %v is error %v", txHash.String(), errS.Error())))
		return nil, nil, err
	}

	// Validate tx by it self
	validate := tx.ValidateTransaction()
	if !validate {
		err := MempoolTxError{}
		err.Init(RejectInvalidTx, errors.New("Invalid tx"))
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
	return nil
}

// RemoveTx safe remove transaction for pool
func (tp *TxPool) RemoveTx(tx transaction.Transaction) error {
	tp.mtx.Lock()
	err := tp.removeTx(&tx)
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

/*
Sum of all transactions sizes
*/
func (tp *TxPool) Size() uint64 {
	tp.mtx.RLock()
	size := uint64(0)
	for _, tx := range tp.pool {
		// TODO: need to implement size func in each type of transactions
		// https://stackoverflow.com/questions/31496804/how-to-get-the-size-of-struct-and-its-contents-in-bytes-in-golang?rq=1
		size += tx.Desc.Tx.GetTxVirtualSize()
	}
	tp.mtx.RUnlock()

	return size
}

func (tp *TxPool) MaxFee() uint64 {
	tp.mtx.RLock()
	fee := uint64(0)
	for _, tx := range tp.pool {
		if tx.Desc.Fee > fee {
			fee = tx.Desc.Fee
		}
	}
	tp.mtx.RUnlock()

	return fee
}

/*
// LastUpdated returns the last time a transaction was added to or
	// removed from the source pool.
*/
func (tp *TxPool) LastUpdated() time.Time {
	return time.Unix(tp.lastUpdated, 0)
}

/*
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

/*
CheckTransactionFee - check fee of tx
*/
func (tp *TxPool) CheckTransactionFee(tx transaction.Transaction) (uint64, error) {
	// Salary transactions have no inputs.
	if blockchain.IsSalaryTx(tx) {
		return 0, nil
	}

	txType := tx.GetType()
	switch txType {
	case common.TxCustomTokenType:
		{
			{
				tx := tx.(*transaction.TxCustomToken)
				err := tp.config.Policy.CheckCustomTokenTransactionFee(tx)
				return tx.Fee, err
			}
		}
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
	case common.TxVotingType:
		{
			votingTx := tx.(*transaction.TxVoting)
			err := tp.config.Policy.CheckVotingTransactionFee(votingTx)
			return votingTx.Fee, err
		}
	default:
		{
			return 0, errors.New("Wrong tx type")
		}
	}
}

func (tp *TxPool) validateSanityNormalTxData(tx *transaction.Tx) (bool, error) {
	txN := tx
	//check version
	if txN.Version > transaction.TxVersion {
		return false, errors.New("Wrong tx version")
	}
	// check LockTime before now
	if int64(txN.LockTime) > time.Now().Unix() {
		return false, errors.New("Wrong tx locktime")
	}
	// check Type is normal or salary tx
	if len(txN.Type) != 1 || (txN.Type != common.TxNormalType && txN.Type != common.TxSalaryType) { // only 1 byte
		return false, errors.New("Wrong tx type")
	}
	// check length of JSPubKey
	if len(txN.JSPubKey) != 64 {
		return false, errors.New("Wrong tx jspubkey")
	}
	// check length of JSSig
	if len(txN.JSSig) != 64 {
		return false, errors.New("Wrong tx jssig")
	}
	//check Descs

	for _, desc := range txN.Descs {
		// check length of Anchor
		if len(desc.Anchor) != 2 {
			return false, errors.New("Wrong tx desc's anchor")
		}
		// check length of EphemeralPubKey
		if len(desc.EphemeralPubKey) != client.EphemeralKeyLength {
			return false, errors.New("Wrong tx desc's ephemeralpubkey")
		}
		// check length of HSigSeed
		if len(desc.HSigSeed) != 32 {
			return false, errors.New("Wrong tx desc's hsigseed")
		}
		// check length of Nullifiers
		if len(desc.Nullifiers) != 2 {
			return false, errors.New("Wrong tx desc's nullifiers")
		}
		if len(desc.Nullifiers[0]) != 32 {
			return false, errors.New("Wrong tx desc's nullifiers")
		}
		if len(desc.Nullifiers[1]) != 32 {
			return false, errors.New("Wrong tx desc's nullifiers")
		}
		// check length of Commitments
		if len(desc.Commitments) != 2 {
			return false, errors.New("Wrong tx desc's commitments")
		}
		if len(desc.Commitments[0]) != 32 {
			return false, errors.New("Wrong tx desc's commitments")
		}
		if len(desc.Commitments[1]) != 32 {
			return false, errors.New("Wrong tx desc's commitments")
		}
		// check length of Vmacs
		if len(desc.Vmacs) != 2 {
			return false, errors.New("Wrong tx desc's vmacs")
		}
		if len(desc.Vmacs[0]) != 32 {
			return false, errors.New("Wrong tx desc's vmacs")
		}
		if len(desc.Vmacs[1]) != 32 {
			return false, errors.New("Wrong tx desc's vmacs")
		}
		//
		if desc.Proof == nil {
			return false, errors.New("Wrong tx desc's proof")
		}
		// check length of Proof
		if len(desc.Proof.G_A) != 33 ||
			len(desc.Proof.G_APrime) != 33 ||
			len(desc.Proof.G_B) != 65 ||
			len(desc.Proof.G_BPrime) != 33 ||
			len(desc.Proof.G_C) != 33 ||
			len(desc.Proof.G_CPrime) != 33 ||
			len(desc.Proof.G_K) != 33 ||
			len(desc.Proof.G_H) != 33 {
			return false, errors.New("Wrong tx desc's proof")
		}
		//
		if len(desc.EncryptedData) != 2 {
			return false, errors.New("Wrong tx desc's encryptedData")
		}
		// check nulltifier is existed in DB
		if desc.Reward != 0 {
			return false, errors.New("Wrong tx desc's reward")
		}
	}
	return true, nil
}

func (tp *TxPool) validateSanityVotingTxData(txVoting *transaction.TxVoting) (bool, error) {
	if !common.ValidateNodeAddress(txVoting.PublicKey) {
		return false, errors.New("Wrong voting node data")
	}
	tx := txVoting.Tx
	txN := tx
	//check version
	if txN.Version > transaction.TxVersion {
		return false, errors.New("Wrong tx version")
	}
	// check LockTime before now
	if int64(txN.LockTime) > time.Now().Unix() {
		return false, errors.New("Wrong tx locktime")
	}
	// check Type equal "n"
	if txN.Type != common.TxVotingType {
		return false, errors.New("Wrong tx type")
	}
	// check length of JSPubKey
	if len(txN.JSPubKey) != 64 {
		return false, errors.New("Wrong tx jspubkey")
	}
	// check length of JSSig
	if len(txN.JSSig) != 64 {
		return false, errors.New("Wrong tx jssig")
	}
	//check Descs

	for _, desc := range txN.Descs {
		// check length of Anchor
		if len(desc.Anchor) != 2 {
			return false, errors.New("Wrong tx desc's anchor")
		}
		// check length of EphemeralPubKey
		if len(desc.EphemeralPubKey) != client.EphemeralKeyLength {
			return false, errors.New("Wrong tx desc's ephemeralpubkey")
		}
		// check length of HSigSeed
		if len(desc.HSigSeed) != 32 {
			return false, errors.New("Wrong tx desc's hsigseed")
		}
		// check length of Nullifiers
		if len(desc.Nullifiers) != 2 {
			return false, errors.New("Wrong tx desc's nullifiers")
		}
		if len(desc.Nullifiers[0]) != 32 {
			return false, errors.New("Wrong tx desc's nullifiers")
		}
		if len(desc.Nullifiers[1]) != 32 {
			return false, errors.New("Wrong tx desc's nullifiers")
		}
		// check length of Commitments
		if len(desc.Commitments) != 2 {
			return false, errors.New("Wrong tx desc's commitments")
		}
		if len(desc.Commitments[0]) != 32 {
			return false, errors.New("Wrong tx desc's commitments")
		}
		if len(desc.Commitments[1]) != 32 {
			return false, errors.New("Wrong tx desc's commitments")
		}
		// check length of Vmacs
		if len(desc.Vmacs) != 2 {
			return false, errors.New("Wrong tx desc's vmacs")
		}
		if len(desc.Vmacs[0]) != 32 {
			return false, errors.New("Wrong tx desc's vmacs")
		}
		if len(desc.Vmacs[1]) != 32 {
			return false, errors.New("Wrong tx desc's vmacs")
		}
		//
		if desc.Proof != nil { // no privacy
			return false, errors.New("Wrong tx desc's proof")
		}
		if len(desc.EncryptedData) != 0 {
			return false, errors.New("Wrong tx desc's encryptedData")
		}
		// check nulltifier is existed in DB
		if desc.Reward != 0 {
			return false, errors.New("Wrong tx desc's reward")
		}
	}
	return true, nil
}

/*
ValidateSanityData - validate sansity data of tx
*/
func (tp *TxPool) ValidateSanityData(tx transaction.Transaction) (bool, error) {
	if tx.GetType() == common.TxNormalType || tx.GetType() == common.TxSalaryType {
		txA := tx.(*transaction.Tx)
		ok, err := tp.validateSanityNormalTxData(txA)
		if !ok {
			return false, err
		}
	} else if tx.GetType() == common.TxActionParamsType {
		txA := tx.(*transaction.ActionParamTx)
		// check Version
		if txA.Version > transaction.TxVersion {
			return false, errors.New("Wrong tx version")
		}
		// check LockTime before now
		if int64(txA.LockTime) > time.Now().Unix() {
			return false, errors.New("Wrong tx lockTime")
		}
		// check Type equal "a"
		if txA.Type != common.TxActionParamsType {
			return false, errors.New("Wrong tx type")
		}
	} else if tx.GetType() == common.TxVotingType {
		txA := tx.(*transaction.TxVoting)
		ok, err := tp.validateSanityVotingTxData(txA)
		if !ok {
			return false, err
		}
	} else if tx.GetType() == common.TxCustomTokenType {
		// TODO check sanity
		return true, nil
	} else {
		return false, errors.New("Wrong tx type")
	}

	return true, nil
}

/*
List all tx ids in mempool
*/
func (tp *TxPool) ListTxs() []string {
	result := make([]string, 0)
	for _, tx := range tp.pool {
		result = append(result, tx.Desc.Tx.Hash().String())
	}
	return result
}
