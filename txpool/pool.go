package txpool

import (
	"fmt"
	"runtime"
	"time"

	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
)

type TxInfo struct {
	Fee   uint64
	Size  uint64
	VTime time.Duration
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
}

func NewTxsPool(
	txVerifier TxVerifier,
	inbox chan metadata.Transaction,
) *TxsPool {
	return &TxsPool{
		action:   make(chan func(*TxsPool)),
		Verifier: txVerifier,
		Data: TxsData{
			TxByHash: map[string]metadata.Transaction{},
			TxInfos:  map[string]TxInfo{},
		},
		Cacher:    cache.New(10*time.Second, 10*time.Second),
		Inbox:     inbox,
		isRunning: false,
		cQuit:     make(chan bool),
		better: func(txA, txB metadata.Transaction) bool {
			return txA.GetTxFee() > txB.GetTxFee()
		},
	}
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
	fmt.Println("[testperformance] Start pool!!")
	tp.isRunning = true
	cValidTxs := make(chan txInfoTemp, 128)
	stopGetTxs := make(chan interface{})
	go tp.getTxs(stopGetTxs, cValidTxs)
	for {
		select {
		case <-tp.cQuit:
			stopGetTxs <- nil
			return
		case validTx := <-cValidTxs:
			txH := validTx.tx.Hash().String()
			tp.Data.TxByHash[txH] = validTx.tx
			tp.Data.TxInfos[txH] = TxInfo{
				Fee:   validTx.tx.GetTxFee(),
				Size:  validTx.tx.GetTxActualSize(),
				VTime: validTx.vt,
			}
		case f := <-tp.action:
			f(tp)
		}
	}
}

func (tp *TxsPool) Stop() {
	tp.cQuit <- true
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
	start := time.Now()
	if _, exist := tp.Cacher.Get(tx.Hash().String()); exist {
		return false, nil, 0
	}
	ok := tp.Verifier.LoadCommitment(tx, nil)
	if !ok {
		return false, errors.Errorf("Can not load commitment for this tx %v", tx.Hash().String()), 0
	}
	ok, err := tp.Verifier.ValidateWithoutChainstate(tx)
	return ok, err, time.Since(start)
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
	res := []metadata.Transaction{}
	txDetailCh := make(chan *TxInfoDetail)
	stopCh := make(chan interface{})
	go tp.getTxsFromPool(txDetailCh, stopCh)
	curSize := uint64(0)
	curTime := 0 * time.Millisecond
	mapForChkDbSpend := map[[privacy.Ed25519KeySize]byte]struct {
		Index  uint
		Detail TxInfoDetail
	}{}
	sDB := sView.GetCopiedTransactionStateDB()
	for {
		select {
		case txDetails := <-txDetailCh:
			if txDetails == nil {
				return res
			}
			if (curSize+txDetails.Size > maxSize) || (curTime+txDetails.VTime > maxTime) {
				continue
			}
			err := txDetails.Tx.LoadCommitment(sDB)
			if err != nil {
				fmt.Printf("Validate tx %v return error %v\n", txDetails.Hash, err)
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
				fmt.Printf("Validate tx %v return error %v\n", txDetails.Hash, err)
				continue
			}
			ok, removedInfo := tp.CheckDoubleSpend(mapForChkDbSpend, txDetails.Tx, res)
			if ok {
				curSize = curSize - removedInfo.Fee + txDetails.Fee
				curTime = curTime - removedInfo.VTime + txDetails.VTime
				insertTxIntoList(mapForChkDbSpend, *txDetails, res)
			}
		case <-timeOut:
			return res
		}
	}
}

func (tp *TxsPool) CheckDoubleSpend(
	dataHelper map[[privacy.Ed25519KeySize]byte]struct {
		Index  uint
		Detail TxInfoDetail
	},
	tx metadata.Transaction,
	txs []metadata.Transaction,
) (
	bool,
	TxInfo,
) {
	iCoins := tx.GetProof().GetInputCoins()
	oCoins := tx.GetProof().GetOutputCoins()
	removedInfos := TxInfo{
		Fee:   0,
		VTime: 0,
	}
	removeIdx := map[uint]interface{}{}
	for _, iCoin := range iCoins {
		if info, ok := dataHelper[iCoin.CoinDetails.GetSerialNumber().ToBytes()]; ok {
			if _, ok := removeIdx[info.Index]; ok {
				continue
			}
			if tp.better(info.Detail.Tx, tx) {
				return false, removedInfos
			} else {
				removeIdx[info.Index] = nil
			}
		}
	}
	for _, oCoin := range oCoins {
		if info, ok := dataHelper[oCoin.CoinDetails.GetSNDerivator().ToBytes()]; ok {
			if _, ok := removeIdx[info.Index]; ok {
				continue
			}
			if tp.better(info.Detail.Tx, tx) {
				return false, removedInfos
			} else {
				removeIdx[info.Index] = nil
			}
		}
	}
	if len(removeIdx) > 0 {
		for k, v := range dataHelper {
			if _, ok := removeIdx[v.Index]; ok {
				delete(dataHelper, k)
			}
		}
		for k := range removeIdx {
			txs = append(txs[:k], txs[k+1:]...)
		}
	}

	return true, removedInfos
}

func insertTxIntoList(
	dataHelper map[[privacy.Ed25519KeySize]byte]struct {
		Index  uint
		Detail TxInfoDetail
	},
	txDetail TxInfoDetail,
	txs []metadata.Transaction,
) {
	tx := txDetail.Tx
	iCoins := tx.GetProof().GetInputCoins()
	oCoins := tx.GetProof().GetOutputCoins()
	for _, iCoin := range iCoins {
		dataHelper[iCoin.CoinDetails.GetSerialNumber().ToBytes()] = struct {
			Index  uint
			Detail TxInfoDetail
		}{
			Index:  uint(len(txs)),
			Detail: txDetail,
		}
	}
	for _, oCoin := range oCoins {
		fmt.Printf("Coin detail %v\n", oCoin.CoinDetails)
		fmt.Printf("Coin detail SND %v\n", oCoin.CoinDetails.GetSNDerivator())
		fmt.Printf("Map %v\n", dataHelper)
		dataHelper[oCoin.CoinDetails.GetSNDerivator().ToBytes()] = struct {
			Index  uint
			Detail TxInfoDetail
		}{
			Index:  uint(len(txs)),
			Detail: txDetail,
		}
	}
	txs = append(txs, tx)
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
			nWorkers <- 1
			go func() {
				isValid, err, vTime := tp.ValidateNewTx(msg)
				<-nWorkers
				if err != nil {
					fmt.Printf("Validate tx %v return error %v:\n", msg.Hash().String(), err)
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

func (tp *TxsPool) getTxsFromPool(
	txCh chan *TxInfoDetail,
	stopC <-chan interface{},
) {
	tp.action <- func(tpTemp *TxsPool) {
		defer func() {
			close(txCh)
			fmt.Println("[testperformance] tx channel is closed")
		}()
		txDetails := &TxInfoDetail{}
		for k, v := range tpTemp.Data.TxByHash {
			select {
			case <-stopC:
				return
			default:
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
					fmt.Printf("[testperformance] Got %v, send to channel\n", txDetails)
					txCh <- txDetails
				}
			}
		}
	}

}

// func (tp *TxsPool) removeTxs(tp)
