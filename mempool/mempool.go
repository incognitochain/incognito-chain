package mempool

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/constant-money/constant-chain/blockchain"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/common/base58"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/transaction"
)

// config is a descriptor containing the memory pool configuration.
type Config struct {
	// Block chain of node
	BlockChain *blockchain.BlockChain

	DataBase database.DatabaseInterface

	ChainParams *blockchain.Params

	// FeeEstimatator provides a feeEstimator. If it is not nil, the mempool
	// records all new transactions it observes into the feeEstimator.
	FeeEstimator map[byte]*FeeEstimator
}

// TxDesc is transaction message in mempool
type TxDesc struct {
	// transaction details
	Desc metadata.TxDesc

	StartingPriority int
}

// TxPool is transaction pool
type TxPool struct {
	// The following variables must only be used atomically.
	lastUpdated int64 // last time pool was updated

	mtx               sync.RWMutex
	config            Config
	pool              map[common.Hash]*TxDesc
	poolSerialNumbers map[common.Hash][][]byte

	txCoinHashHPool map[common.Hash][]common.Hash
	coinHashHPool   map[common.Hash]bool
	cMtx            sync.RWMutex

	//Candidate List in mempool
	candidateList []string
	candidateMtx  sync.RWMutex

	//Token ID List in Mempool
	tokenIDList []string
	tokenIDMtx  sync.RWMutex
}

/*
Init Txpool from config
*/
func (tp *TxPool) Init(cfg *Config) {
	tp.config = *cfg
	tp.pool = make(map[common.Hash]*TxDesc)
	tp.poolSerialNumbers = make(map[common.Hash][][]byte)

	tp.txCoinHashHPool = make(map[common.Hash][]common.Hash)
	tp.coinHashHPool = make(map[common.Hash]bool)
	tp.cMtx = sync.RWMutex{}
}

// ----------- transaction.MempoolRetriever's implementation -----------------
func (tp *TxPool) GetSerialNumbers() map[common.Hash][][]byte {
	return tp.poolSerialNumbers
}

func (tp *TxPool) GetTxsInMem() map[common.Hash]metadata.TxDesc {
	txsInMem := make(map[common.Hash]metadata.TxDesc)
	for hash, txDesc := range tp.pool {
		txsInMem[hash] = txDesc.Desc
	}
	return txsInMem
}

// ----------- end of transaction.MempoolRetriever's implementation -----------------

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
func (tp *TxPool) addTx(tx metadata.Transaction, height uint64, fee uint64) *TxDesc {
	txD := &TxDesc{
		Desc: metadata.TxDesc{
			Tx:     tx,
			Added:  time.Now(),
			Height: height,
			Fee:    fee,
		},
		StartingPriority: 1, //@todo we will apply calc function for it.
	}
	Logger.log.Info(tx.Hash().String())
	tp.pool[*tx.Hash()] = txD
	tp.poolSerialNumbers[*tx.Hash()] = txD.Desc.Tx.ListNullifiers()
	atomic.StoreInt64(&tp.lastUpdated, time.Now().Unix())

	// Record this tx for fee estimation if enabled. only apply for normal tx
	if tx.GetType() == common.TxNormalType {
		if tp.config.FeeEstimator != nil {
			shardID := common.GetShardIDFromLastByte(tx.(*transaction.Tx).PubKeyLastByteSender)
			if temp, ok := tp.config.FeeEstimator[shardID]; ok {
				temp.ObserveTransaction(txD)
			}

		}
	}
	txHash := tx.Hash()
	if txHash != nil {
		tp.AddTxCoinHashH(*txHash)
	}
	// add candidate into candidate list ONLY with staking transaction
	if tx.GetMetadata() != nil {
		if tx.GetMetadata().GetType() == metadata.ShardStakingMeta || tx.GetMetadata().GetType() == metadata.BeaconStakingMeta {
			pubkey := base58.Base58Check{}.Encode(tx.GetSigPubKey(), byte(0x00))
			tp.AddCandiateToList(pubkey)
		}
	}
	if tx.GetType() == common.TxCustomTokenType {
		customTokenTx := tx.(*transaction.TxCustomToken)
		if customTokenTx.TxTokenData.Type == transaction.CustomTokenInit {
			tokenID := customTokenTx.TxTokenData.PropertyID.String()
			tp.AddTokenIDToList(tokenID)
		}
	}
	return txD
}

/*
// maybeAcceptTransaction is the internal function which implements the public
// See the comment for MaybeAcceptTransaction for more details.
// This function MUST be called with the mempool lock held (for writes).
1. Validate tx version
2.1 Validate size of transaction (can't greater than max size of block)
2.2 Validate fee with tx size
3. Validate type of tx
4. Validate with other txs in mempool
5. Validate sanity data of tx
6. Validate data in tx: privacy proof, metadata,...
7. Validate tx with blockchain: douple spend, ...
8. Check tx existed in mempool
9. Not accept a salary tx
10. Check Duplicate stake public key in pool ONLY with staking transaction
*/
func (tp *TxPool) maybeAcceptTransaction(tx metadata.Transaction) (*common.Hash, *TxDesc, error) {
	txHash := tx.Hash()

	// Don't accept the transaction if it already exists in the pool.
	if tp.isTxInPool(txHash) {
		str := fmt.Sprintf("already have transaction %+v", txHash.String())
		err := MempoolTxError{}
		err.Init(RejectDuplicateTx, errors.New(str))
		return nil, nil, err
	}

	// that make sure transaction is accepted when passed any rules
	var shardID byte
	var err error

	// get shardID of tx
	shardID = common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())

	bestHeight := tp.config.BlockChain.BestState.Shard[shardID].BestBlock.Header.Height
	// nextBlockHeight := bestHeight + 1
	// check version
	ok := tx.CheckTxVersion(MaxVersion)
	if !ok {
		err := MempoolTxError{}
		err.Init(RejectVersion, fmt.Errorf("transaction %+v's version is invalid", txHash.String()))
		return nil, nil, err
	}

	// check actual size
	actualSize := tx.GetTxActualSize()
	fmt.Printf("Transaction %+v's size %+v \n", txHash, actualSize)
	if actualSize >= common.MaxBlockSize || actualSize >= common.MaxTxSize {
		err := MempoolTxError{}
		err.Init(RejectInvalidSize, fmt.Errorf("transaction %+v's size is invalid, more than %+v Kilobyte", txHash.String(), common.MaxBlockSize))
		return nil, nil, err
	}

	// check fee of tx
	minFeePerKbTx := tp.config.BlockChain.GetFeePerKbTx()
	txFee := tx.GetTxFee()
	ok = tx.CheckTransactionFee(minFeePerKbTx)
	if !ok {
		err := MempoolTxError{}
		err.Init(RejectVersion, fmt.Errorf("transaction %+v has %d fees which is under the required amount of %d", tx.Hash().String(), txFee, minFeePerKbTx))
		return nil, nil, err
	}
	// end check with policy

	ok = tx.ValidateType()
	if !ok {
		return nil, nil, errors.New("wrong tx type")
	}
	// check tx with all txs in current mempool
	err = tx.ValidateTxWithCurrentMempool(tp)
	if err != nil {
		return nil, nil, err
	}

	// sanity data
	//if validated, errS := tx.ValidateSanityData(tp.config.BlockChain); !validated {
	//	err := MempoolTxError{}
	//	err.Init(RejectSansityTx, fmt.Errorf("transaction's sansity %v is error %v", txHash.String(), errS.Error()))
	//	return nil, nil, err
	//}

	// ValidateTransaction tx by it self // TODO validate performance later 0xkraken
	validated := tx.ValidateTxByItself(tx.IsPrivacy(), tp.config.BlockChain.GetDatabase(), tp.config.BlockChain, shardID)
	if !validated {
		err := MempoolTxError{}
		err.Init(RejectInvalidTx, errors.New("invalid tx"))
		return nil, nil, err
	}

	// validate tx with data of blockchain
	err = tx.ValidateTxWithBlockChain(tp.config.BlockChain, shardID, tp.config.BlockChain.GetDatabase())
	// err = tp.ValidateTxWithBlockChain(tx, shardID)
	if err != nil {
		return nil, nil, err
	}

	if tx.GetType() == common.TxCustomTokenType {
		customTokenTx := tx.(*transaction.TxCustomToken)
		if customTokenTx.TxTokenData.Type == transaction.CustomTokenInit {
			//tokenID := customTokenTx.TxTokenData.PropertyID.String()
			//tp.tokenIDMtx.Lock()
			//found := common.IndexOfStr(tokenID, tp.tokenIDList)
			//tp.tokenIDMtx.Unlock()
			//if found > -1 {
			//	return nil, nil, errors.New("Init Transaction of this Token is in pool already")
			//}
		}
	}

	// A standalone transaction must not be a salary transaction.
	if tx.IsSalaryTx() {
		err := MempoolTxError{}
		err.Init(RejectSalaryTx, fmt.Errorf("%+v is salary tx", txHash.String()))
		return nil, nil, err
	}

	// check duplicate stake public key ONLY with staking transaction
	if tx.GetMetadata() != nil {
		if tx.GetMetadata().GetType() == metadata.ShardStakingMeta || tx.GetMetadata().GetType() == metadata.BeaconStakingMeta {
			pubkey := base58.Base58Check{}.Encode(tx.GetSigPubKey(), byte(0x00))
			tempPubkey := metadata.GetValidStaker(tp.candidateList, []string{pubkey})
			if len(tempPubkey) == 0 {
				str := fmt.Sprintf("This public key already stake and still in pool %+v", pubkey)
				err := MempoolTxError{}
				err.Init(RejectDuplicateStakeTx, errors.New(str))
				return nil, nil, err
			}
		}
	}

	txD := tp.addTx(tx, bestHeight, txFee)
	return tx.Hash(), txD, nil
}

// remove transaction for pool
func (tp *TxPool) removeTx(tx *metadata.Transaction) error {
	Logger.log.Infof((*tx).Hash().String())
	if _, exists := tp.pool[*(*tx).Hash()]; exists {
		delete(tp.pool, *(*tx).Hash())
		atomic.StoreInt64(&tp.lastUpdated, time.Now().Unix())
		return nil
	} else {
		return errors.New("not exist tx in pool")
	}
	return nil
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
func (tp *TxPool) MaybeAcceptTransaction(tx metadata.Transaction) (*common.Hash, *TxDesc, error) {
	tp.mtx.Lock()
	hash, txDesc, err := tp.maybeAcceptTransaction(tx)
	if err != nil {
		Logger.log.Error(err)
	}
	tp.mtx.Unlock()

	return hash, txDesc, err
}

// This function is safe for concurrent access.
func (tp *TxPool) MaybeAcceptTransactionForBlockProducing(tx metadata.Transaction) (*metadata.TxDesc, error) {
	tp.mtx.Lock()
	defer tp.mtx.Unlock()
	_, txDesc, err := tp.maybeAcceptTransaction(tx)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}
	tempTxDesc := &txDesc.Desc
	return tempTxDesc, err
}

// RemoveTx safe remove transaction for pool
func (tp *TxPool) RemoveTx(tx metadata.Transaction) error {
	tp.mtx.Lock()
	fmt.Println("...................................")
	fmt.Println("txHash To Be Remove", tx.Hash())
	fmt.Println("...................................")
	err := tp.removeTx(&tx)
	// remove tx coin hash from pool
	txHash := tx.Hash()
	if txHash != nil {
		tp.RemoveTxCoinHashH(*txHash)
	}
	tp.mtx.Unlock()
	return err
}

// GetTx get transaction info by hash
func (tp *TxPool) GetTx(txHash *common.Hash) (metadata.Transaction, error) {
	tp.mtx.Lock()
	Logger.log.Info(txHash.String())
	txDesc, exists := tp.pool[*txHash]
	tp.mtx.Unlock()
	if exists {
		return txDesc.Desc.Tx, nil
	}

	return nil, errors.New("transaction is not in the pool")
}

// // MiningDescs returns a slice of mining descriptors for all the transactions
// // in the pool.
func (tp *TxPool) MiningDescs() []*metadata.TxDesc {
	fmt.Println()
	fmt.Println("Current Transaction in pool", tp.pool)
	fmt.Println()
	descs := []*metadata.TxDesc{}
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
		size += tx.Desc.Tx.GetTxActualSize()
	}
	tp.mtx.RUnlock()

	return size
}

// Get Max fee
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
List all tx ids in mempool
*/
func (tp *TxPool) ListTxs() []string {
	result := make([]string, 0)
	for _, tx := range tp.pool {
		result = append(result, tx.Desc.Tx.Hash().String())
	}
	return result
}

// PrePoolTxCoinHashH -
func (tp *TxPool) PrePoolTxCoinHashH(txHashH common.Hash, coinHashHs []common.Hash) error {
	tp.cMtx.Lock()
	defer tp.cMtx.Unlock()
	tp.txCoinHashHPool[txHashH] = coinHashHs
	return nil
}

// addTxCoinHashH - add hash of output coin
//// which use to check double spend in memppol
func (tp *TxPool) AddTxCoinHashH(txHashH common.Hash) error {
	tp.cMtx.Lock()
	defer tp.cMtx.Unlock()
	inCoinHs, ok := tp.txCoinHashHPool[txHashH]
	if ok {
		for _, inCoinH := range inCoinHs {
			tp.coinHashHPool[inCoinH] = true
		}
	}
	return nil
}

// ValidateCoinHashH - check outputcoin which is
// used by a tx in mempool
func (tp *TxPool) ValidateCoinHashH(coinHashH common.Hash) error {
	tp.cMtx.Lock()
	defer tp.cMtx.Unlock()
	_, ok := tp.coinHashHPool[coinHashH]
	if ok {
		return errors.New("Coin is in used")
	}
	return nil
}

// removeTxCoinHashH remove hash of output coin
// which use to check double spend in memppol
func (tp *TxPool) RemoveTxCoinHashH(txHashH common.Hash) error {
	tp.cMtx.Lock()
	defer tp.cMtx.Unlock()
	if coinHashHs, okTxHashH := tp.txCoinHashHPool[txHashH]; okTxHashH {
		for _, coinHashH := range coinHashHs {
			if _, okCoinHashH := tp.coinHashHPool[coinHashH]; okCoinHashH {
				delete(tp.coinHashHPool, coinHashH)
			}
		}
		delete(tp.txCoinHashHPool, txHashH)
	}
	return nil
}

func (tp *TxPool) AddCandiateToList(candidate string) {
	tp.candidateMtx.Lock()
	defer tp.candidateMtx.Unlock()
	// fmt.Println("Mempool/addCanđiateToList: ", candidate)
	tp.candidateList = append(tp.candidateList, candidate)
}

func (tp *TxPool) RemoveCandidateList(candidate []string) {
	tp.candidateMtx.Lock()
	defer tp.candidateMtx.Unlock()
	newList := []string{}
	for _, value := range candidate {
		flag := false
		for _, currentCandidate := range tp.candidateList {
			if strings.Compare(value, currentCandidate) == 0 {
				flag = true
				break
			}
		}
		if !flag {
			newList = append(newList, value)
		}
	}
	tp.candidateList = newList
}
func (tp *TxPool) AddTokenIDToList(tokenID string) {
	tp.tokenIDMtx.Lock()
	defer tp.tokenIDMtx.Unlock()
	// fmt.Println("Mempool/addCanđiateToList: ", candidate)
	tp.tokenIDList = append(tp.tokenIDList, tokenID)
}

func (tp *TxPool) RemoveTokenIDList(tokenID []string) {
	tp.tokenIDMtx.Lock()
	defer tp.tokenIDMtx.Unlock()
	newList := []string{}
	for _, value := range tokenID {
		flag := false
		for _, currentToken := range tp.tokenIDList {
			if strings.Compare(value, currentToken) == 0 {
				flag = true
				break
			}
		}
		if !flag {
			newList = append(newList, value)
		}
	}
	tp.tokenIDList = newList
}

/*
	pool              map[common.Hash]*TxDesc
	poolSerialNumbers map[common.Hash][][]byte

	txCoinHashHPool map[common.Hash][]common.Hash
	coinHashHPool   map[common.Hash]bool
	cMtx            sync.RWMutex

	//Candidate List in mempool
	candidateList []string
	candidateMtx  sync.RWMutex

	//Token ID List in Mempool
	tokenIDList []string
	tokenIDMtx  sync.RWMutex
*/
func (tp *TxPool) EmptyPool() bool {
	tp.cMtx.Lock()
	tp.candidateMtx.Lock()
	tp.tokenIDMtx.Lock()

	defer tp.cMtx.Unlock()
	defer tp.candidateMtx.Unlock()
	defer tp.tokenIDMtx.Unlock()

	if len(tp.pool) == 0 && len(tp.poolSerialNumbers) == 0 && len(tp.txCoinHashHPool) == 0 && len(tp.coinHashHPool) == 0 && len(tp.candidateList) == 0 && len(tp.tokenIDList) == 0 {
		return true
	}

	for key := range tp.pool {
		delete(tp.pool, key)
	}

	for key := range tp.poolSerialNumbers {
		delete(tp.poolSerialNumbers, key)
	}

	for key := range tp.txCoinHashHPool {
		delete(tp.txCoinHashHPool, key)
	}

	for key := range tp.coinHashHPool {
		delete(tp.coinHashHPool, key)
	}
	tp.candidateList = []string{}
	tp.tokenIDList = []string{}

	if len(tp.pool) == 0 && len(tp.poolSerialNumbers) == 0 && len(tp.txCoinHashHPool) == 0 && len(tp.coinHashHPool) == 0 && len(tp.candidateList) == 0 && len(tp.tokenIDList) == 0 {
		return true
	}
	fmt.Println("len(tp.pool)", len(tp.pool))
	fmt.Println("len(tp.poolSerialNumbers)", len(tp.poolSerialNumbers))
	fmt.Println("len(tp.txCoinHashHPool)", len(tp.txCoinHashHPool))
	fmt.Println("len(tp.coinHashHPool)", len(tp.coinHashHPool))
	fmt.Println("len(tp.candidateList)", len(tp.candidateList))
	fmt.Println("len(tp.tokenIDList)", len(tp.tokenIDList))
	return false
}
