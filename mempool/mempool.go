package mempool

import (
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/metrics"
	"github.com/incognitochain/incognito-chain/pubsub"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/incognitochain/incognito-chain/incognitokey"

	"github.com/incognitochain/incognito-chain/databasemp"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/transaction"
)

// config is a descriptor containing the memory pool configuration.
type Config struct {
	BlockChain            *blockchain.BlockChain // Block chain of node
	DataBase              database.DatabaseInterface
	DataBaseMempool       databasemp.DatabaseInterface
	ChainParams           *blockchain.Params
	FeeEstimator          map[byte]*FeeEstimator // FeeEstimatator provides a feeEstimator. If it is not nil, the mempool records all new transactions it observes into the feeEstimator.
	TxLifeTime            uint                   // Transaction life time in pool
	MaxTx                 uint64                 //Max transaction pool may have
	IsLoadFromMempool     bool                   //Reset mempool database when run node
	PersistMempool        bool
	RelayShards           []byte
	UserKeyset            *incognitokey.KeySet
	PubSubManager         *pubsub.PubSubManager
	RoleInCommittees      int //Current Role of Node
	RoleInCommitteesEvent pubsub.EventChannel
}

// TxDesc is transaction message in mempool
type TxDesc struct {
	Desc            metadata.TxDesc // transaction details
	StartTime       time.Time       //Unix Time that transaction enter mempool
	IsFowardMessage bool
}

// TxPool is transaction pool
type TxPool struct {
	// The following variables must only be used atomically.
	lastUpdated            int64 // last time pool was updated
	mtx                    sync.RWMutex
	config                 Config
	pool                   map[common.Hash]*TxDesc
	poolSerialNumbersHashH map[common.Hash][]common.Hash // [txHash]:list hash serialNumbers of input coin
	Scantime               time.Duration
	CandidatePool          map[common.Hash]string //Candidate List in mempool
	candidateMtx           sync.RWMutex
	TokenIDPool            map[common.Hash]string //Token ID List in Mempool
	tokenIDMtx             sync.RWMutex
	DuplicateTxs           map[common.Hash]uint64 //For testing
	roleMtx                sync.RWMutex
	CPendingTxs            chan<- metadata.Transaction // channel to deliver txs to block gen
	IsBlockGenStarted      bool
	IsUnlockMempool        bool
	IsTest                 bool
}

/*
Init Txpool from config
*/
func (tp *TxPool) Init(cfg *Config) {
	tp.config = *cfg
	tp.Scantime = 1 * time.Hour
	tp.pool = make(map[common.Hash]*TxDesc)
	tp.poolSerialNumbersHashH = make(map[common.Hash][]common.Hash)
	tp.TokenIDPool = make(map[common.Hash]string)
	tp.CandidatePool = make(map[common.Hash]string)
	tp.DuplicateTxs = make(map[common.Hash]uint64)
	tp.config.RoleInCommittees = -1
	tp.IsBlockGenStarted = false
	tp.IsUnlockMempool = true
	_, subChanRole, _ := tp.config.PubSubManager.RegisterNewSubscriber(pubsub.ShardRoleTopic)
	tp.config.RoleInCommitteesEvent = subChanRole
	tp.IsTest = false
}
func (tp *TxPool) InitChannelMempool(cPendingTxs chan metadata.Transaction) {
	tp.CPendingTxs = cPendingTxs
}
func (tp *TxPool) InitDatabaseMempool(db databasemp.DatabaseInterface) {
	tp.config.DataBaseMempool = db
}
func (tp *TxPool) AnnouncePersisDatabaseMempool() {
	if tp.config.PersistMempool {
		Logger.log.Critical("Turn on Mempool Persistence Database")
	} else {
		Logger.log.Critical("Turn off Mempool Persistence Database")
	}
}
func (tp *TxPool) LoadOrResetDatabaseMP() {
	if !tp.config.IsLoadFromMempool {
		err := tp.ResetDatabaseMP()
		if err != nil {
			Logger.log.Errorf("Fail to reset mempool database, error: %+v \n", err)
		} else {
			Logger.log.Critical("Successfully Reset from database")
		}
	} else {
		txDescs, err := tp.LoadDatabaseMP()
		if err != nil {
			Logger.log.Errorf("Fail to load mempool database, error: %+v \n", err)
		} else {
			Logger.log.Criticalf("Successfully load %+v from database \n", len(txDescs))
		}
	}
	//return []TxDesc{}
}
func createTxDescMempool(tx metadata.Transaction, height uint64, fee uint64, feeToken uint64) *TxDesc {
	txDesc := &TxDesc{
		Desc: metadata.TxDesc{
			Tx:       tx,
			Height:   height,
			Fee:      fee,
			FeeToken: feeToken,
		},
		StartTime:       time.Now(),
		IsFowardMessage: false,
	}
	return txDesc
}
// ----------- transaction.MempoolRetriever's implementation -----------------
func (tp *TxPool) GetSerialNumbersHashH() map[common.Hash][]common.Hash {
	return tp.poolSerialNumbersHashH
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
// #1: tx
// #2: store into db
// #3: default nil, contain input coins hash, which are used for creating this tx
*/
func (tp *TxPool) addTx(txD *TxDesc, isStore bool) {
	tx := txD.Desc.Tx
	txHash := tx.Hash()
	if isStore {
		err := tp.AddTransactionToDatabaseMP(txHash, *txD)
		if err != nil {
			Logger.log.Errorf("Fail to add tx %+v to mempool database %+v \n", *txHash, err)
		} else {
			Logger.log.Criticalf("Add tx %+v to mempool database success \n", *txHash)
		}
	}
	tp.pool[*txHash] = txD
	//==================================================
	tp.poolSerialNumbersHashH[*txHash] = txD.Desc.Tx.ListSerialNumbersHashH()
	atomic.StoreInt64(&tp.lastUpdated, time.Now().Unix())
	// Record this tx for fee estimation if enabled, apply for normal tx and privacy token tx
	if tp.config.FeeEstimator != nil {
		var shardID byte
		flag := false
		switch tx.GetType() {
		case common.TxNormalType:
			{
				shardID = common.GetShardIDFromLastByte(tx.(*transaction.Tx).PubKeyLastByteSender)
				flag = true
			}
		case common.TxCustomTokenPrivacyType:
			{
				shardID = common.GetShardIDFromLastByte(tx.(*transaction.TxCustomTokenPrivacy).PubKeyLastByteSender)
				flag = true
			}
		}
		if flag {
			if temp, ok := tp.config.FeeEstimator[shardID]; ok {
				temp.ObserveTransaction(txD)
			}
		}
	}
	// add candidate into candidate list ONLY with staking transaction
	if tx.GetMetadata() != nil {
		metadataType := tx.GetMetadata().GetType()
		switch metadataType {
		case metadata.ShardStakingMeta, metadata.BeaconStakingMeta:
			{
				publicKey := base58.Base58Check{}.Encode(tx.GetSigPubKey(), common.ZeroByte)
				tp.addCandidateToList(*txHash, publicKey)
			}
		default:
			{
				Logger.log.Debug("Metadata Type:", metadataType)
			}
		}
	}
	if tx.GetType() == common.TxCustomTokenType {
		customTokenTx := tx.(*transaction.TxCustomToken)
		if customTokenTx.TxTokenData.Type == transaction.CustomTokenInit {
			tokenID := customTokenTx.TxTokenData.PropertyID.String()
			tp.addTokenIDToList(*txHash, tokenID)
		}
	}
	Logger.log.Infof("Add Transaction %+v Successs \n", txHash.String())
}

/*
// maybeAcceptTransaction is the internal function which implements the public
// See the comment for MaybeAcceptTransaction for more details.
// This function MUST be called with the mempool lock held (for writes).
1. Validate tx version
2.1 Validate size of transaction (can't greater than max size of block)
2.2 Not accept a salary tx
2.3 Validate fee with tx size
3. Validate type of tx
4. Validate with other txs in mempool
5. Validate sanity data of tx
6. Validate data in tx: privacy proof, metadata,...
7. Validate tx with blockchain: douple spend, ...
8. Check tx existed in mempool
10. Check Duplicate stake public key in pool ONLY with staking transaction

Param#2: isStore: store transaction to persistence storage only work for transaction come from user (not for validation process)
*/
func (tp *TxPool) validateTransaction(tx metadata.Transaction) error {
	var shardID byte
	var err error
	var now time.Time
	txHash := tx.Hash()
	txType := tx.GetType()
	if txType == common.TxNormalType {
		if tx.IsPrivacy() {
			txType = metrics.TxNormalPrivacy
		} else {
			txType = metrics.TxNormalNoPrivacy
		}
	}
	now = time.Now()
	// Don't accept the transaction if it already exists in the pool.
	if tp.isTxInPool(txHash) {
		str := fmt.Sprintf("already have transaction %+v", txHash.String())
		go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
			metrics.Measurement:      metrics.TxPoolDuplicateTxs,
			metrics.MeasurementValue: float64(1),
		})
		err := MempoolTxError{}
		err.Init(RejectDuplicateTx, errors.New(str))
		return err
	}
	go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
		metrics.Measurement:      metrics.TxPoolValidationDetails,
		metrics.MeasurementValue: float64(time.Since(now).Seconds()),
		metrics.TagValue:         metrics.Condition1,
		metrics.Tag:              metrics.ValidateConditionTag,
	})
	// check version
	now = time.Now()
	ok := tx.CheckTxVersion(MaxVersion)
	if !ok {
		err := MempoolTxError{}
		err.Init(RejectVersion, fmt.Errorf("transaction %+v's version is invalid", txHash.String()))
		return err
	}
	go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
		metrics.Measurement:      metrics.TxPoolValidationDetails,
		metrics.MeasurementValue: float64(time.Since(now).Seconds()),
		metrics.TagValue:         metrics.Condition2,
		metrics.Tag:              metrics.ValidateConditionTag,
	})
	// check actual size
	now = time.Now()
	actualSize := tx.GetTxActualSize()
	Logger.log.Debugf("Transaction %+v 's size %+v \n", *txHash, actualSize)
	if actualSize >= common.MaxBlockSize || actualSize >= common.MaxTxSize {
		err := MempoolTxError{}
		err.Init(RejectInvalidSize, fmt.Errorf("transaction %+v's size is invalid, more than %+v Kilobyte", txHash.String(), common.MaxBlockSize))
		return err
	}
	go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
		metrics.Measurement:      metrics.TxPoolValidationDetails,
		metrics.MeasurementValue: float64(time.Since(now).Seconds()),
		metrics.TagValue:         metrics.Condition3,
		metrics.Tag:              metrics.ValidateConditionTag,
	})
	// A standalone transaction must not be a salary transaction.
	now = time.Now()
	if tx.IsSalaryTx() {
		err := MempoolTxError{}
		err.Init(RejectSalaryTx, fmt.Errorf("%+v is salary tx", txHash.String()))
		return err
	}
	go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
		metrics.Measurement:      metrics.TxPoolValidationDetails,
		metrics.MeasurementValue: float64(time.Since(now).Seconds()),
		metrics.TagValue:         metrics.Condition4,
		metrics.Tag:              metrics.ValidateConditionTag,
	})
	// check fee of tx
	now = time.Now()
	limitFee := tp.config.FeeEstimator[shardID].limitFee
	txFee := tx.GetTxFee()
	ok = tx.CheckTransactionFee(limitFee)
	if !ok {
		err := MempoolTxError{}
		err.Init(RejectInvalidFee, fmt.Errorf("transaction %+v has %d fees which is under the required amount of %d", txHash.String(), txFee, limitFee*tx.GetTxActualSize()))
		return err
	}
	go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
		metrics.Measurement:      metrics.TxPoolValidationDetails,
		metrics.MeasurementValue: float64(time.Since(now).Seconds()),
		metrics.TagValue:         metrics.Condition5,
		metrics.Tag:              metrics.ValidateConditionTag,
	})
	// end check with policy
	now = time.Now()
	ok = tx.ValidateType()
	if !ok {
		return err
	}
	go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
		metrics.Measurement:      metrics.TxPoolValidationDetails,
		metrics.MeasurementValue: float64(time.Since(now).Seconds()),
		metrics.TagValue:         metrics.Condition6,
		metrics.Tag:              metrics.ValidateConditionTag,
	})
	// check tx with all txs in current mempool
	now = time.Now()
	err = tx.ValidateTxWithCurrentMempool(tp)
	if err != nil {
		return err
	}
	go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
		metrics.Measurement:      metrics.TxPoolValidationDetails,
		metrics.MeasurementValue: float64(time.Since(now).Seconds()),
		metrics.TagValue:         metrics.Condition7,
		metrics.Tag:              metrics.ValidateConditionTag,
	})
	// sanity data
	now = time.Now()
	if validated, errS := tx.ValidateSanityData(tp.config.BlockChain); !validated {
		err := MempoolTxError{}
		err.Init(RejectSansityTx, fmt.Errorf("transaction's sansity %v is error %v", txHash.String(), errS.Error()))
		return err
	}
	go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
		metrics.Measurement:      metrics.TxPoolValidationDetails,
		metrics.MeasurementValue: float64(time.Since(now).Seconds()),
		metrics.TagValue:         metrics.Condition8,
		metrics.Tag:              metrics.ValidateConditionTag,
	})
	// ValidateTransaction tx by it self
	shardID = common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
	now = time.Now()
	validated, errValidateTxByItself := tx.ValidateTxByItself(tx.IsPrivacy(), tp.config.DataBase, tp.config.BlockChain, shardID)
	go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
		metrics.Measurement:      metrics.TxPoolValidationDetails,
		metrics.MeasurementValue: float64(time.Since(now).Seconds()),
		metrics.TagValue:         metrics.VTBITxTypeMetic,
		metrics.Tag:              metrics.ValidateConditionTag,
	})
	if !validated {
		err := MempoolTxError{}
		messageError := "Invalid tx - "
		if errValidateTxByItself != nil {
			messageError += errValidateTxByItself.Error()
		}
		err.Init(RejectInvalidTx, errors.New(messageError))
		return err
	}

	// validate tx with data of blockchain
	now = time.Now()
	err = tx.ValidateTxWithBlockChain(tp.config.BlockChain, shardID, tp.config.DataBase)
	if err != nil {
		return err
	}
	go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
		metrics.Measurement:      metrics.TxPoolValidationDetails,
		metrics.MeasurementValue: float64(time.Since(now).Seconds()),
		metrics.TagValue:         metrics.Condition9,
		metrics.Tag:              metrics.ValidateConditionTag,
	})
	now = time.Now()
	if tx.GetType() == common.TxCustomTokenType {
		customTokenTx := tx.(*transaction.TxCustomToken)
		if customTokenTx.TxTokenData.Type == transaction.CustomTokenInit {
			tokenID := customTokenTx.TxTokenData.PropertyID.String()
			tp.tokenIDMtx.RLock()
			found := common.IndexOfStrInHashMap(tokenID, tp.TokenIDPool)
			tp.tokenIDMtx.RUnlock()
			if found > 0 {
				str := fmt.Sprintf("Init Transaction of this Token is in pool already %+v", tokenID)
				err := MempoolTxError{}
				err.Init(RejectDuplicateInitTokenTx, errors.New(str))
				return err
			}
		}
	}
	go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
		metrics.Measurement:      metrics.TxPoolValidationDetails,
		metrics.MeasurementValue: float64(time.Since(now).Seconds()),
		metrics.TagValue:         metrics.Condition10,
		metrics.Tag:              metrics.ValidateConditionTag,
	})
	// check duplicate stake public key ONLY with staking transaction
	now = time.Now()
	if tx.GetMetadata() != nil {
		if tx.GetMetadata().GetType() == metadata.ShardStakingMeta || tx.GetMetadata().GetType() == metadata.BeaconStakingMeta {
			pubkey := base58.Base58Check{}.Encode(tx.GetSigPubKey(), common.ZeroByte)
			tp.candidateMtx.RLock()
			found := common.IndexOfStrInHashMap(pubkey, tp.CandidatePool)
			tp.candidateMtx.RUnlock()
			if found > 0 {
				str := fmt.Sprintf("This public key already stake and still in pool %+v", pubkey)
				err := MempoolTxError{}
				err.Init(RejectDuplicateStakeTx, errors.New(str))
				return err
			}
		}
	}
	go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
		metrics.Measurement:      metrics.TxPoolValidationDetails,
		metrics.MeasurementValue: float64(time.Since(now).Seconds()),
		metrics.TagValue:         metrics.Condition11,
		metrics.Tag:              metrics.ValidateConditionTag,
	})
	return nil
}

/*
// maybeAcceptTransaction into pool
// #1: tx
// #2: store into db
// #3: default nil, contain input coins hash, which are used for creating this tx
*/
func (tp *TxPool) maybeAcceptTransaction(tx metadata.Transaction, isStore bool, isNewTransaction bool) (*common.Hash, *TxDesc, error) {
	txType := tx.GetType()
	if txType == common.TxNormalType {
		if tx.IsPrivacy() {
			txType = metrics.TxNormalPrivacy
		} else {
			txType = metrics.TxNormalNoPrivacy
		}
	}
	txSize := fmt.Sprintf("%d", tx.GetTxActualSize())
	startValidate := time.Now()
	err := tp.validateTransaction(tx)
	if err != nil {
		return nil, nil, err
	}
	elapsed := float64(time.Since(startValidate).Seconds())

	go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
		metrics.Measurement:      metrics.TxPoolValidated,
		metrics.MeasurementValue: elapsed,
		metrics.Tag:              metrics.TxSizeTag,
		metrics.TagValue:         txSize})
	go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
		metrics.Measurement:      metrics.TxPoolValidatedWithType,
		metrics.MeasurementValue: elapsed,
		metrics.Tag:              metrics.TxSizeTag,
		metrics.TagValue:         txSize})
	shardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
	bestHeight := tp.config.BlockChain.BestState.Shard[shardID].BestBlock.Header.Height
	txFee := tx.GetTxFee()
	txFeeToken := tx.GetTxFeeToken()
	txD := createTxDescMempool(tx, bestHeight, txFee, txFeeToken)
	startAdd := time.Now()
	tp.addTx(txD, isStore)
	if isNewTransaction {
		Logger.log.Infof("Add New Txs Into Pool %+v FROM SHARD %+v\n", *tx.Hash(), shardID)
		go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
			metrics.Measurement:      metrics.TxPoolAddedAfterValidation,
			metrics.MeasurementValue: float64(time.Since(startAdd).Seconds()),
			metrics.Tag:              metrics.TxSizeTag,
			metrics.TagValue:         txSize})
	}
	return tx.Hash(), txD, nil
}

// remove transaction for pool
func (tp *TxPool) removeTx(tx *metadata.Transaction) {
	//Logger.log.Infof((*tx).Hash().String())
	if _, exists := tp.pool[*(*tx).Hash()]; exists {
		delete(tp.pool, *(*tx).Hash())
		atomic.StoreInt64(&tp.lastUpdated, time.Now().Unix())
	}
}

// Check relay shard and public key role before processing transaction
func (tp *TxPool) checkRelayShard(tx metadata.Transaction) bool {
	senderShardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
	if common.IndexOfByte(senderShardID, tp.config.RelayShards) > -1 {
		return true
	}
	return false
}

func (tp *TxPool) checkPublicKeyRole(tx metadata.Transaction) bool {
	senderShardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
	tp.roleMtx.RLock()
	if tp.config.RoleInCommittees > -1 && byte(tp.config.RoleInCommittees) == senderShardID {
		tp.roleMtx.RUnlock()
		return true
	} else {
		tp.roleMtx.RUnlock()
		return false
	}
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
// #1: tx
// #2: default nil, contain input coins hash, which are used for creating this tx
func (tp *TxPool) MaybeAcceptTransaction(tx metadata.Transaction) (*common.Hash, *TxDesc, error) {
	tp.mtx.Lock()
	defer tp.mtx.Unlock()
	if tp.IsTest {
		err := MempoolTxError{}
		err.Init(UnexpectedTransactionError, errors.New("Not allowed test tx"))
		return &common.Hash{}, &TxDesc{}, nil
	}
	go func(txHash common.Hash) {
		tp.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.TransactionHashEnterNodeTopic, txHash))
	}(*tx.Hash())
	if !tp.checkRelayShard(tx) && !tp.checkPublicKeyRole(tx) {
		senderShardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
		err := MempoolTxError{}
		err.Init(UnexpectedTransactionError, errors.New("Unexpected Transaction From Shard "+fmt.Sprintf("%d", senderShardID)))
		Logger.log.Error(err)
		return &common.Hash{}, &TxDesc{}, err
	}
	txType := tx.GetType()
	if txType == common.TxNormalType {
		if tx.IsPrivacy() {
			txType = metrics.TxNormalPrivacy
		} else {
			txType = metrics.TxNormalNoPrivacy
		}
	}
	txSize := fmt.Sprintf("%d", tx.GetTxActualSize())
	txTypePrivacyOrNot := metrics.TxPrivacy
	go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
		metrics.Measurement:      metrics.TxPoolTxBeginEnter,
		metrics.MeasurementValue: float64(1),
		metrics.Tag:              metrics.TxTypeTag,
		metrics.TagValue:         txType})
	//==========
	if uint64(len(tp.pool)) >= tp.config.MaxTx {
		err := MempoolTxError{}
		err.Init(MaxPoolSizeError, errors.New("Pool reach max number of transaction"))
		return nil, nil, err
	}
	startAdd := time.Now()
	hash, txDesc, err := tp.maybeAcceptTransaction(tx, tp.config.PersistMempool, true)
	elapsed := float64(time.Since(startAdd).Seconds())
	//==========
	go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
		metrics.Measurement:      metrics.TxPoolEntered,
		metrics.MeasurementValue: elapsed,
		metrics.Tag:              metrics.TxSizeTag,
		metrics.TagValue:         txSize})
	go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
		metrics.Measurement:      metrics.TxPoolEnteredWithType,
		metrics.MeasurementValue: elapsed,
		metrics.Tag:              metrics.TxSizeTag,
		metrics.TagValue:         txSize})
	size := len(tp.pool)
	go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
		metrics.Measurement:      metrics.PoolSize,
		metrics.MeasurementValue: float64(size)})
	go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
		metrics.Measurement:      metrics.TxAddedIntoPoolType,
		metrics.MeasurementValue: float64(1),
		metrics.Tag:              metrics.TxTypeTag,
		metrics.TagValue:         txType,
	})
	if !tx.IsPrivacy() {
		txTypePrivacyOrNot = metrics.TxNoPrivacy
	}
	go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
		metrics.Measurement:      metrics.TxPoolPrivacyOrNot,
		metrics.MeasurementValue: float64(1),
		metrics.Tag:              metrics.TxPrivacyOrNotTag,
		metrics.TagValue:         txTypePrivacyOrNot,
	})
	if err != nil {
		Logger.log.Error(err)
	} else {
		if tp.IsBlockGenStarted {
			if tp.IsUnlockMempool {
				go func(tx metadata.Transaction) {
					tp.CPendingTxs <- tx
				}(tx)
			}
		}
		// Publish Message
		go tp.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.MempoolInfoTopic, tp.listTxs()))
	}
	return hash, txDesc, err
}
func (tp *TxPool) SendTransactionToBlockGen() {
	tp.mtx.RLock()
	defer tp.mtx.RUnlock()
	for _, txdesc := range tp.pool {
		tp.CPendingTxs <- txdesc.Desc.Tx
	}
	tp.IsUnlockMempool = true
}

func (tp *TxPool) MarkForwardedTransaction(txHash common.Hash) {
	if tp.IsTest {
		return
	}
	tp.pool[txHash].IsFowardMessage = true
}

// This function is safe for concurrent access.
func (tp *TxPool) MaybeAcceptTransactionForBlockProducing(tx metadata.Transaction) (*metadata.TxDesc, error) {
	tp.mtx.Lock()
	defer tp.mtx.Unlock()
	_, txDesc, err := tp.maybeAcceptTransaction(tx, false, false)
	if err != nil {
		Logger.log.Error(err)
		fmt.Printf("[db] maybe err: %+v\n", tx.GetMetadataType())
		return nil, err
	}
	tempTxDesc := &txDesc.Desc
	return tempTxDesc, err
}

// RemoveTx safe remove transaction for pool
func (tp *TxPool) RemoveTx(txs []metadata.Transaction, isInBlock bool) {
	tp.mtx.Lock()
	defer tp.mtx.Unlock()
	// remove transaction from database mempool
	for _, tx := range txs {
		var now time.Time
		start := time.Now()
		now = time.Now()
		txDesc, ok := tp.pool[*tx.Hash()]
		if !ok {
			continue
		}
		txType := tx.GetType()
		if txType == common.TxNormalType {
			if tx.IsPrivacy() {
				txType = metrics.TxNormalPrivacy
			} else {
				txType = metrics.TxNormalNoPrivacy
			}
		}
		startTime := txDesc.StartTime
		if tp.config.PersistMempool {
			tp.RemoveTransactionFromDatabaseMP(tx.Hash())
		}
		go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
			metrics.Measurement:      metrics.TxPoolRemovedTimeDetails,
			metrics.MeasurementValue: float64(time.Since(now).Seconds()),
			metrics.Tag:              metrics.ValidateConditionTag,
			metrics.TagValue:         metrics.Condition1,
		})
		tp.removeTx(&tx)
		// remove serialNumbersHashH
		now = time.Now()
		delete(tp.poolSerialNumbersHashH, *(tx.Hash()))
		go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
			metrics.Measurement:      metrics.TxPoolRemovedTimeDetails,
			metrics.MeasurementValue: float64(time.Since(now).Seconds()),
			metrics.Tag:              metrics.ValidateConditionTag,
			metrics.TagValue:         metrics.Condition2,
		})
		now = time.Now()
		if isInBlock {
			elapsed := float64(time.Since(startTime).Seconds())
			txSize := fmt.Sprintf("%d", tx.GetTxActualSize())
			go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
				metrics.Measurement:      metrics.TxPoolRemoveAfterInBlock,
				metrics.MeasurementValue: elapsed,
				metrics.Tag:              metrics.TxSizeTag,
				metrics.TagValue:         txSize,
			})
			go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
				metrics.Measurement:      metrics.TxPoolRemoveAfterInBlockWithType,
				metrics.MeasurementValue: elapsed,
				metrics.Tag:              metrics.TxSizeWithTypeTag,
				metrics.TagValue:         txType + txSize,
			})
		}
		go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
			metrics.Measurement:      metrics.TxPoolRemovedNumber,
			metrics.MeasurementValue: float64(1),
		})
		size := len(tp.pool)
		go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
			metrics.Measurement:      metrics.PoolSize,
			metrics.MeasurementValue: float64(size),
		})
		go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
			metrics.Measurement:      metrics.TxPoolRemovedTimeDetails,
			metrics.MeasurementValue: float64(time.Since(now).Seconds()),
			metrics.Tag:              metrics.ValidateConditionTag,
			metrics.TagValue:         metrics.Condition3,
		})
		go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
			metrics.Measurement:      metrics.TxPoolRemovedTime,
			metrics.MeasurementValue: float64(time.Since(start).Seconds()),
			metrics.Tag:              metrics.TxTypeTag,
			metrics.TagValue:         txType,
		})
	}
	return
}

// GetTx get transaction info by hash
func (tp *TxPool) GetTx(txHash *common.Hash) (metadata.Transaction, error) {
	tp.mtx.Lock()
	defer tp.mtx.Unlock()
	Logger.log.Info(txHash.String())
	txDesc, exists := tp.pool[*txHash]
	if exists {
		return txDesc.Desc.Tx, nil
	}
	err := MempoolTxError{}
	err.Init(TransactionNotFoundError, errors.New("Transaction "+txHash.String()+" Not Found!"))
	Logger.log.Error(err)
	return nil, err
}

// // MiningDescs returns a slice of mining descriptors for all the transactions
// // in the pool.
func (tp *TxPool) MiningDescs() []*metadata.TxDesc {
	tp.mtx.Lock()
	defer tp.mtx.Unlock()
	descs := []*metadata.TxDesc{}
	for _, desc := range tp.pool {
		descs = append(descs, &desc.Desc)
	}
	return descs
}

func (tp *TxPool) GetPool() map[common.Hash]*TxDesc {
	return tp.pool
}

func (tp *TxPool) LockPool() {
	tp.mtx.Lock()
}

func (tp *TxPool) UnlockPool() {
	tp.mtx.Unlock()
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
	defer tp.mtx.RUnlock()
	size := uint64(0)
	for _, tx := range tp.pool {
		size += tx.Desc.Tx.GetTxActualSize()
	}
	return size
}

// Get Max fee
func (tp *TxPool) MaxFee() uint64 {
	tp.mtx.RLock()
	defer tp.mtx.RUnlock()
	fee := uint64(0)
	for _, tx := range tp.pool {
		if tx.Desc.Fee > fee {
			fee = tx.Desc.Fee
		}
	}
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
	defer tp.mtx.RUnlock()
	haveTx := tp.isTxInPool(hash)
	return haveTx
}

/*
List all tx ids in mempool
*/
func (tp *TxPool) ListTxs() []string {
	tp.mtx.RLock()
	defer tp.mtx.RUnlock()
	return tp.listTxs()
}

/*
List all tx ids in mempool
*/
func (tp *TxPool) listTxs() []string {
	result := make([]string, 0)
	for _, tx := range tp.pool {
		result = append(result, tx.Desc.Tx.Hash().String())
	}
	return result
}

/*
List all tx ids in mempool
*/
func (tp *TxPool) ListTxsDetail() []metadata.Transaction {
	tp.mtx.RLock()
	defer tp.mtx.RUnlock()
	result := make([]metadata.Transaction, 0)
	for _, tx := range tp.pool {
		result = append(result, tx.Desc.Tx)
	}
	return result
}

// ValidateSerialNumberHashH - check serialNumberHashH which is
// used by a tx in mempool
func (tp *TxPool) ValidateSerialNumberHashH(serialNumber []byte) error {
	hash := common.HashH(serialNumber)
	for txHash, serialNumbersHashH := range tp.poolSerialNumbersHashH {
		_ = txHash
		for _, serialNumberHashH := range serialNumbersHashH {
			if serialNumberHashH.IsEqual(&hash) {
				return errors.New("Coin is in used")
			}
		}
	}
	return nil
}

func (tp *TxPool) addCandidateToList(txHash common.Hash, candidate string) {
	tp.candidateMtx.Lock()
	defer tp.candidateMtx.Unlock()
	tp.CandidatePool[txHash] = candidate
}

func (tp *TxPool) RemoveCandidateList(candidate []string) {
	tp.candidateMtx.Lock()
	defer tp.candidateMtx.Unlock()
	candidateToBeRemoved := []common.Hash{}
	for _, value := range candidate {
		for txHash, currentCandidate := range tp.CandidatePool {
			if strings.Compare(value, currentCandidate) == 0 {
				candidateToBeRemoved = append(candidateToBeRemoved, txHash)
				break
			}
		}
	}
	for _, txHash := range candidateToBeRemoved {
		delete(tp.CandidatePool, txHash)
	}
}

func (tp *TxPool) addTokenIDToList(txHash common.Hash, tokenID string) {
	tp.tokenIDMtx.Lock()
	defer tp.tokenIDMtx.Unlock()
	tp.TokenIDPool[txHash] = tokenID
}

func (tp *TxPool) RemoveTokenIDList(tokenID []string) {
	tp.tokenIDMtx.Lock()
	defer tp.tokenIDMtx.Unlock()
	tokenToBeRemoved := []common.Hash{}
	for _, value := range tokenID {
		for txHash, currentToken := range tp.TokenIDPool {
			if strings.Compare(value, currentToken) == 0 {
				tokenToBeRemoved = append(tokenToBeRemoved, txHash)
				break
			}
		}
	}
	for _, txHash := range tokenToBeRemoved {
		delete(tp.TokenIDPool, txHash)
	}
}

func (tp *TxPool) EmptyPool() bool {
	tp.candidateMtx.Lock()
	tp.tokenIDMtx.Lock()
	defer tp.candidateMtx.Unlock()
	defer tp.tokenIDMtx.Unlock()
	if len(tp.pool) == 0 && len(tp.poolSerialNumbersHashH) == 0 && len(tp.CandidatePool) == 0 && len(tp.TokenIDPool) == 0 {
		return true
	}
	tp.pool = make(map[common.Hash]*TxDesc)
	tp.poolSerialNumbersHashH = make(map[common.Hash][]common.Hash)
	tp.CandidatePool = make(map[common.Hash]string)
	tp.TokenIDPool = make(map[common.Hash]string)
	if len(tp.pool) == 0 && len(tp.poolSerialNumbersHashH) == 0 && len(tp.CandidatePool) == 0 && len(tp.TokenIDPool) == 0 {
		return true
	}
	return false
}

func (tp *TxPool) calPoolSize() uint64 {
	var totalSize uint64
	for _, txDesc := range tp.pool {
		size := txDesc.Desc.Tx.GetTxActualSize()
		totalSize += size
	}
	return totalSize
}

func (tp *TxPool) monitorPool() {
	if tp.config.TxLifeTime == 0 {
		return
	}
	for {
		<-time.Tick(tp.Scantime)
		tp.mtx.Lock()
		ttl := time.Duration(tp.config.TxLifeTime) * time.Second
		txsToBeRemoved := []*TxDesc{}
		for _, txDesc := range tp.pool {
			if time.Since(txDesc.StartTime) > ttl {
				txsToBeRemoved = append(txsToBeRemoved, txDesc)
			}
		}
		for _, txDesc := range txsToBeRemoved {
			txHash := *txDesc.Desc.Tx.Hash()
			startTime := txDesc.StartTime
			delete(tp.pool, txHash)
			delete(tp.poolSerialNumbersHashH, txHash)
			delete(tp.CandidatePool, txHash)
			delete(tp.TokenIDPool, txHash)
			txSize := txDesc.Desc.Tx.GetTxActualSize()
			go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
				metrics.Measurement:      metrics.TxPoolRemoveAfterLifeTime,
				metrics.MeasurementValue: float64(time.Since(startTime).Seconds()),
				metrics.Tag:              metrics.TxSizeTag,
				metrics.TagValue:         txSize,
			})
			size := len(tp.pool)
			go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
				metrics.Measurement:      metrics.PoolSize,
				metrics.MeasurementValue: float64(size)})
		}
		tp.mtx.Unlock()
	}
}

func (tp *TxPool) Start(cQuit chan struct{}) {
	go tp.monitorPool()
	for {
		select {
		case <-cQuit:
			return
		case msg := <-tp.config.RoleInCommitteesEvent:
			{
				shardID, ok := msg.Value.(int)
				if !ok {
					continue
				}
				go func() {
					tp.roleMtx.Lock()
					defer tp.roleMtx.Unlock()
					tp.config.RoleInCommittees = shardID
				}()
			}
		}
	}
}
