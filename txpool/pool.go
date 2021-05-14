package txpool

import (
	"log"
	"runtime"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	zkp "github.com/incognitochain/incognito-chain/privacy/zeroknowledge"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
)

type TxInfo struct {
	Fee   uint64
	Size  uint64
	VTime time.Duration
}

type validateResult struct {
	err    error
	result bool
	cost   time.Duration
}

type TxInfoDetail struct {
	Hash  string
	Fee   uint64
	Size  uint64
	VTime time.Duration
	Tx    metadata.Transaction
}

type TxsData struct {
	TxByHash map[string]metadata.Transaction
	TxInfos  map[string]TxInfo
}

type txInfoTemp struct {
	tx metadata.Transaction
	vt time.Duration
}

type TxsPool struct {
	action    chan func(*TxsPool)
	Verifier  TxVerifier
	Data      TxsData
	Cacher    *cache.Cache
	Inbox     chan metadata.Transaction
	isRunning bool
	cQuit     chan bool
	better    func(txA, txB metadata.Transaction) bool
	ttl       time.Duration
}

func NewTxsPool(
	txVerifier TxVerifier,
	inbox chan metadata.Transaction,
	ttl time.Duration,
) *TxsPool {
	tp := &TxsPool{
		action:   make(chan func(*TxsPool)),
		Verifier: txVerifier,
		Data: TxsData{
			TxByHash: map[string]metadata.Transaction{},
			TxInfos:  map[string]TxInfo{},
		},
		Cacher:    cache.New(ttl, ttl),
		Inbox:     inbox,
		isRunning: false,
		cQuit:     make(chan bool),
		better: func(txA, txB metadata.Transaction) bool {
			return txA.GetTxFee() > txB.GetTxFee()
		},
		ttl: ttl,
	}
	tp.Cacher.OnEvicted(tp.removeTx)
	return tp
}

func (tp *TxsPool) UpdateTxVerifier(tv TxVerifier) {
	tp.Verifier = tv
}

func (tp *TxsPool) GetInbox() chan metadata.Transaction {
	return tp.Inbox
}

func (tp *TxsPool) IsRunning() bool {
	return tp.isRunning
}

func (tp *TxsPool) Start() {
	if tp.isRunning {
		return
	}
	Logger.Infof("Start transaction pool v1")
	tp.isRunning = true
	cValidTxs := make(chan txInfoTemp, 1024)
	stopGetTxs := make(chan interface{})
	go tp.getTxs(stopGetTxs, cValidTxs)
	total := 0
	for {
		select {
		case <-tp.cQuit:
			stopGetTxs <- nil
<<<<<<< HEAD
			tp.isRunning = false
=======
>>>>>>> 2757919da... Enable staking v2
			return
		case f := <-tp.action:
			Logger.Debugf("Total txs received %v, total txs in pool %v\n", total, len(tp.Data.TxInfos))
			f(tp)
			Logger.Debugf("Total txs in pool %v after func\n", len(tp.Data.TxInfos))
		case validTx := <-cValidTxs:
			total++
			txH := validTx.tx.Hash().String()
			tp.Data.TxByHash[txH] = validTx.tx
			tp.Data.TxInfos[txH] = TxInfo{
				Fee:   validTx.tx.GetTxFee(),
				Size:  validTx.tx.GetTxActualSize(),
				VTime: validTx.vt,
			}
		}
	}
}

func (tp *TxsPool) Stop() {
	tp.cQuit <- true
}

func (tp *TxsPool) removeTx(txHash string, arg interface{}) {
	tp.action <- func(tpTemp *TxsPool) {
		Logger.Debugf("Removing tx %v at %v", txHash, time.Now())
		delete(tpTemp.Data.TxByHash, txHash)
		delete(tpTemp.Data.TxInfos, txHash)
	}
}

func (tp *TxsPool) RemoveTxs(txHashes []string) {
	tp.action <- func(tpTemp *TxsPool) {
		for _, tx := range txHashes {
			delete(tpTemp.Data.TxByHash, tx)
			delete(tpTemp.Data.TxInfos, tx)
		}
	}
}

func (tp *TxsPool) ValidateNewTx(tx metadata.Transaction) (bool, error, time.Duration) {
	txHash := tx.Hash().String()
	start := time.Now()
	Logger.Debugf("[txTracing] Start validate tx %v at %v", txHash, start.UTC())
	t := time.NewTimer(2 * time.Second)
	defer t.Stop()
	errChan := make(chan validateResult)
	go func() {
		if ok := isTxForUser(tx); !ok {
			Logger.Debugf("This transaction %v can not be sent by user", tx.Hash().String())
			errChan <- validateResult{
				err:    errors.Errorf("This transaction %v can not be sent by user", tx.Hash().String()),
				result: false,
				cost:   0,
			}
			return
		}
		if _, exist := tp.Cacher.Get(tx.Hash().String()); exist {
			log.Printf("[txTracing] Not validate tx %v cuz it found in cache, cost %v", txHash, time.Since(start))
			errChan <- validateResult{
				err:    nil,
				result: false,
				cost:   0,
			}
			return
		} else {
			Logger.Debugf("Caching tx %v at %v", tx.Hash().String(), time.Now())
			tp.Cacher.Add(tx.Hash().String(), nil, tp.ttl)
		}
		if ok, err := tp.Verifier.LoadCommitment(tx, nil); !ok || err != nil {
			Logger.Debugf("[txTracing] validate tx %v failed, error %v, cost %v", txHash, err, time.Since(start))
			errChan <- validateResult{
				err:    err,
				result: false,
				cost:   time.Since(start),
			}
			return
		}
		ok, err := tp.Verifier.ValidateWithoutChainstate(tx)
		errChan <- validateResult{
			err:    err,
			result: ok,
			cost:   time.Since(start),
		}
	}()
	select {
	case <-t.C:
		return false, errors.Errorf("[stream] Trying send to client but timeout"), 0
	case err := <-errChan:
		Logger.Debugf("[txTracing] validate tx %v return %v, error %v, cost %v", txHash, err.result, err.err, err.cost)
		return err.result, err.err, err.cost
	}
}

func (tp *TxsPool) GetTxsTranferForNewBlock(
	cView metadata.ChainRetriever,
	sView metadata.ShardViewRetriever,
	bcView metadata.BeaconViewRetriever,
	maxSize uint64,
	maxTime time.Duration,
	getTxsDuration time.Duration,
) []metadata.Transaction {
	//TODO Timeout
	timeOut := time.After(getTxsDuration)
	Logger.Infof("Has %v time for crawling txs for shard %v\n", getTxsDuration, sView.GetShardID())
	st := time.Now()
	res := []metadata.Transaction{}
	txDetailCh := make(chan *TxInfoDetail, 1024)
	stopCh := make(chan interface{}, 2)
	go tp.getTxsFromPool(txDetailCh, stopCh)
	curSize := uint64(0)
	curTime := 0 * time.Millisecond
	mapForChkDbSpend := map[[privacy.Ed25519KeySize]byte]struct {
		Index  uint
		Detail TxInfoDetail
	}{}
	mapForChkDbStake := map[string]interface{}{}
	defer func() {
		Logger.Infof("Return list txs #res %v cursize %v curtime %v maxsize %v maxtime %v for shard %v \n", len(res), curSize, curTime, maxSize, maxTime, sView.GetShardID())
	}()
	limitTxAction := map[int]int{}
	for {
		select {
		case txDetails := <-txDetailCh:
			if txDetails == nil {
				removeNilTx(&res)
				return res
			}
			Logger.Debugf("[txTracing] Validate new tx %v with chainstate\n", txDetails.Tx.Hash().String())
			if (curSize+txDetails.Size > maxSize) || (curTime+txDetails.VTime > maxTime) {
				continue
			}
			if ok := checkTxAction(limitTxAction, txDetails.Tx); !ok {
				continue
			}
			if tp.isDoubleStake(mapForChkDbStake, txDetails.Tx) {
				Logger.Errorf("[txTracing] Tx %v is stake/unstake/stop auto stake twice\n", txDetails.Hash)
				continue
			}
			if ok, err := tp.Verifier.LoadCommitment(txDetails.Tx, sView); !ok || err != nil {
				Logger.Errorf("[txTracing] Validate tx %v return error %v\n", txDetails.Hash, err)
				continue
			}
			ok, err := tp.Verifier.ValidateWithChainState(
				txDetails.Tx,
				cView,
				sView,
				bcView,
				sView.GetBeaconHeight(),
			)
			if !ok || err != nil {
				Logger.Errorf("[txTracing]Validate tx %v return error %v\n", txDetails.Hash, err)
				continue
			}
			Logger.Debugf("Try to add tx %v into list txs #res %v\n", txDetails.Tx.Hash().String(), len(res))
			isDoubleSpend, needToReplace, removedInfo := tp.CheckDoubleSpend(mapForChkDbSpend, txDetails.Tx, &res)
			if isDoubleSpend && !needToReplace {
				continue
			}
			curSize = curSize - removedInfo.Size + txDetails.Size
			curTime = curTime - removedInfo.VTime + txDetails.VTime
			Logger.Debugf("Added tx %v, %v %v\n", txDetails.Tx.Hash().String(), needToReplace, removedInfo)
			res = insertTxIntoList(mapForChkDbSpend, *txDetails, res)
		case <-timeOut:
			stopCh <- nil
			Logger.Debugf("Crawling txs for new block shard %v timeout! %v\n", sView.GetShardID(), time.Since(st))
			removeNilTx(&res)
			return res
		}
	}
}

func (tp *TxsPool) isDoubleStake(
	dataHelper map[string]interface{},
	tx metadata.Transaction,
) bool {
	metaType := tx.GetMetadataType()
	pk := ""
	switch metaType {
	case metadata.ShardStakingMeta, metadata.BeaconStakingMeta:
		if meta, ok := tx.GetMetadata().(*metadata.StakingMetadata); ok {
			pk = meta.CommitteePublicKey
		}
	case metadata.UnStakingMeta:
		if meta, ok := tx.GetMetadata().(*metadata.UnStakingMetadata); ok {
			pk = meta.CommitteePublicKey
		}
	case metadata.StopAutoStakingMeta:
		if meta, ok := tx.GetMetadata().(*metadata.StopAutoStakingMetadata); ok {
			pk = meta.CommitteePublicKey
		}
	}
	if pk != "" {
		if _, existed := dataHelper[pk]; existed {
			return true
		}
		dataHelper[pk] = nil
	}
	return false
}

func (tp *TxsPool) CheckDoubleSpend(
	dataHelper map[[privacy.Ed25519KeySize]byte]struct {
		Index  uint
		Detail TxInfoDetail
	},
	tx metadata.Transaction,
	txs *[]metadata.Transaction,
) (
	bool,
	bool,
	TxInfo,
) {
	prf := tx.GetProof()
	removedInfos := TxInfo{
		Fee:   0,
		VTime: 0,
	}
	removeIdx := map[uint]interface{}{}
	isDoubleSpend := false
	needToReplace := false
	if prf != nil {
		isDoubleSpend, needToReplace, removeIdx, removedInfos = tp.checkPrfDoubleSpend(prf, dataHelper, removeIdx, tx, removedInfos)
		if isDoubleSpend && !needToReplace {
			return isDoubleSpend, needToReplace, removedInfos
		}
	}

	if tx.GetType() == common.TxCustomTokenPrivacyType {
		txNormal := tx.(*transaction.TxCustomTokenPrivacy).TxPrivacyTokenData.TxNormal
		normalPrf := txNormal.GetProof()
		if normalPrf != nil {
			isDoubleSpend, needToReplace, removeIdx, removedInfos = tp.checkPrfDoubleSpend(normalPrf, dataHelper, removeIdx, tx, removedInfos)
			if isDoubleSpend && !needToReplace {
				return isDoubleSpend, needToReplace, removedInfos
			}
		}
	}

	if len(removeIdx) > 0 {
		Logger.Debugf("%v %v Doublespend %v\n", len(removeIdx), len((*txs)), tx.Hash().String())
		for k, v := range dataHelper {
			if _, ok := removeIdx[v.Index]; ok {
				delete(dataHelper, k)
			}
		}
		for k := range removeIdx {
			(*txs)[k] = nil
		}
	}
	return isDoubleSpend, needToReplace, removedInfos
}

func (tp *TxsPool) checkPrfDoubleSpend(
	prf *zkp.PaymentProof,
	dataHelper map[[privacy.Ed25519KeySize]byte]struct {
		Index  uint
		Detail TxInfoDetail
	},
	removeIdx map[uint]interface{},
	tx metadata.Transaction,
	removedInfos TxInfo,
) (bool, bool, map[uint]interface{}, TxInfo) {
	needToReplace := false
	isDoubleSpend := false
	iCoins := prf.GetInputCoins()
	oCoins := prf.GetOutputCoins()
	for _, iCoin := range iCoins {
		if info, ok := dataHelper[iCoin.CoinDetails.GetSerialNumber().ToBytes()]; ok {
			isDoubleSpend = true
			if _, ok := removeIdx[info.Index]; ok {
				needToReplace = true
				continue
			}
			if tp.better(info.Detail.Tx, tx) {
				return true, false, removeIdx, removedInfos
			} else {
				removeIdx[info.Index] = nil
				removedInfos.Fee += info.Detail.Fee
				removedInfos.Size += info.Detail.Size
				removedInfos.VTime += info.Detail.VTime
				needToReplace = true
			}
		}
	}
	for _, oCoin := range oCoins {
		if info, ok := dataHelper[oCoin.CoinDetails.GetSNDerivator().ToBytes()]; ok {
			isDoubleSpend = true
			if _, ok := removeIdx[info.Index]; ok {
				continue
			}
			if tp.better(info.Detail.Tx, tx) {
				return true, false, removeIdx, removedInfos
			} else {
				removeIdx[info.Index] = nil
				removedInfos.Fee += info.Detail.Fee
				removedInfos.Size += info.Detail.Size
				removedInfos.VTime += info.Detail.VTime
				needToReplace = true
			}
		}
	}

	return isDoubleSpend, needToReplace, removeIdx, removedInfos
}

func insertTxIntoList(
	dataHelper map[[privacy.Ed25519KeySize]byte]struct {
		Index  uint
		Detail TxInfoDetail
	},
	txDetail TxInfoDetail,
	txs []metadata.Transaction,
) []metadata.Transaction {
	tx := txDetail.Tx
	prf := tx.GetProof()
	if prf != nil {
		insertPrfForCheck(prf, dataHelper, txDetail, len(txs))
	}
	if tx.GetType() == common.TxCustomTokenPrivacyType {
		txNormal := tx.(*transaction.TxCustomTokenPrivacy).TxPrivacyTokenData.TxNormal
		normalPrf := txNormal.GetProof()
		if normalPrf != nil {
			insertPrfForCheck(normalPrf, dataHelper, txDetail, len(txs))
		}
	}
	return append(txs, tx)
}

func insertPrfForCheck(
	prf *zkp.PaymentProof,
	dataHelper map[[privacy.Ed25519KeySize]byte]struct {
		Index  uint
		Detail TxInfoDetail
	},
	txDetail TxInfoDetail,
	idx int,
) {
	iCoins := prf.GetInputCoins()
	oCoins := prf.GetOutputCoins()
	for _, iCoin := range iCoins {
		dataHelper[iCoin.CoinDetails.GetSerialNumber().ToBytes()] = struct {
			Index  uint
			Detail TxInfoDetail
		}{
			Index:  uint(idx),
			Detail: txDetail,
		}
	}
	for _, oCoin := range oCoins {
		dataHelper[oCoin.CoinDetails.GetSNDerivator().ToBytes()] = struct {
			Index  uint
			Detail TxInfoDetail
		}{
			Index:  uint(idx),
			Detail: txDetail,
		}
	}
}

func (tp *TxsPool) CheckValidatedTxs(
	txs []metadata.Transaction,
) (
	valid []metadata.Transaction,
	needValidate []metadata.Transaction,
) {
	if !tp.isRunning {
		return []metadata.Transaction{}, txs
	}
	poolData := tp.snapshotPool()
	for _, tx := range txs {
		if _, ok := poolData.TxInfos[tx.Hash().String()]; ok {
			valid = append(valid, tx)
		} else {
			needValidate = append(needValidate, tx)
		}
	}
	return valid, needValidate
}

func (tp *TxsPool) getTxs(quit <-chan interface{}, cValidTxs chan txInfoTemp) {
	MAX := runtime.NumCPU() - 1
	nWorkers := make(chan int, MAX)
	for {
		select {
		case <-quit:
			return
		default:
			msg := <-tp.Inbox
			txHah := msg.Hash().String()
			workerID := len(nWorkers)
			Logger.Debugf("[txTracing] Received new tx %v, send to worker %v", txHah, workerID)
			nWorkers <- 1
			go func() {
				isValid, err, vTime := tp.ValidateNewTx(msg)
				<-nWorkers
				if err != nil {
					Logger.Errorf("Validate tx %v return error %v:\n", msg.Hash().String(), err)
				}
				if isValid {
					cValidTxs <- txInfoTemp{
						msg,
						vTime,
					}
				}
			}()
		}
	}
}

func (tp *TxsPool) snapshotPool() TxsData {
	cData := make(chan TxsData)
	tp.action <- func(tpTemp *TxsPool) {
		res := TxsData{
			TxByHash: map[string]metadata.Transaction{},
			TxInfos:  map[string]TxInfo{},
		}
		for k, v := range tpTemp.Data.TxByHash {
			res.TxByHash[k] = v
		}
		for k, v := range tpTemp.Data.TxInfos {
			res.TxInfos[k] = v
		}
		cData <- res
	}
	return <-cData
}

func (tp *TxsPool) getTxByHash(txID string) metadata.Transaction {
	cData := make(chan metadata.Transaction)
	tp.action <- func(tpTemp *TxsPool) {
		if tx, ok := tpTemp.Data.TxByHash[txID]; ok {
			cData <- tx
		} else {
			cData <- nil
		}
	}
	return <-cData
}

func (tp *TxsPool) getTxsFromPool(
	txCh chan *TxInfoDetail,
	stopC <-chan interface{},
) {
	tp.action <- func(tpTemp *TxsPool) {
		defer func() {
			close(txCh)
			Logger.Debug("tx channel is closed")
		}()
		for k, v := range tpTemp.Data.TxByHash {
			select {
			case <-stopC:
				return
			default:
				txDetails := &TxInfoDetail{}
				if info, ok := tpTemp.Data.TxInfos[k]; ok {
					txDetails.Hash = k
					txDetails.Fee = info.Fee
					txDetails.Size = info.Size
					txDetails.VTime = info.VTime
				} else {
					continue
				}
				if v != nil {
					txDetails.Tx = v
					Logger.Debugf("[debugperformance] Got %v, send to channel\n", txDetails.Hash)
					if txCh != nil {
						txCh <- txDetails
					}
				}
			}
		}
		if txCh != nil {
			txCh <- nil
		}
	}

}

func checkTxAction(
	remining map[int]int,
	tx metadata.Transaction,
) bool {
	act := metadata.GetMetaAction(tx.GetMetadataType())
	if act == metadata.NoAction {
		return true
	}
	if _, ok := remining[act]; !ok {
		remining[act] = metadata.GetLimitOfMeta(tx.GetMetadataType())
	}
	limit := remining[act]
	if limit < 1 {
		Logger.Errorf("[rejecttx] Total txs %v is larger than limit %v, reject this tx %v \n", act, limit, tx.Hash().String())
		return false
	}
	remining[act] = limit - 1
	return true
}

func isTxForUser(tx metadata.Transaction) bool {
	if tx.GetType() == common.TxReturnStakingType {
		return false
	}
	if tx.GetType() == common.TxRewardType {
		return false
	}
	if tx.GetType() == common.TxCustomTokenPrivacyType {
		tempTx, ok := tx.(*transaction.TxCustomTokenPrivacy)
		if !ok {
			return false
		}
		if tempTx.TxPrivacyTokenData.Mintable {
			return false
		}
	}
	return true
}

func removeNilTx(txs *[]metadata.Transaction) {
	j := 0
	for _, tx := range *txs {
		if tx == nil {
			continue
		}
		(*txs)[j] = tx
		j++
	}
	*txs = (*txs)[:j]
}
