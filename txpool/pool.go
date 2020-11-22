package txpool

import (
	"fmt"
	"runtime"
	"time"

	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/patrickmn/go-cache"
)

type TxInfo struct {
	Fee   uint64
	Size  uint64
	VTime time.Duration
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
	// SID       byte
	action    chan func(*TxsPool)
	Verifier  TxVerifier
	Data      TxsData
	Cacher    *cache.Cache
	Inbox     chan metadata.Transaction
	isRunning bool
	cQuit     chan bool
}

func NewTxsPool(
	txVerifier TxVerifier,
	inbox chan metadata.Transaction,
) *TxsPool {
	return &TxsPool{
		action:    make(chan func(*TxsPool)),
		Verifier:  txVerifier,
		Data:      TxsData{},
		Cacher:    new(cache.Cache),
		Inbox:     inbox,
		isRunning: false,
		cQuit:     make(chan bool),
	}
}

func (tp *TxsPool) Start() {
	if tp.isRunning {
		return
	}
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
	ok, err := tp.Verifier.ValidateAuthentications(tx)
	if (err != nil) || (!ok) {
		return ok, err, 0
	}
	ok, err = tp.Verifier.ValidateDataCorrectness(tx)
	if (err != nil) || (!ok) {
		return ok, err, 0
	}
	ok, err = tp.Verifier.ValidateTxZKProof(tx)
	return ok, err, time.Since(start)
}

func (tp *TxsPool) GetTxsTranferForNewBlock(
	sView interface{},
	bcView interface{},
	maxSize uint64,
	maxTime time.Duration,
) []metadata.Transaction {
	poolData := tp.snapshotPool()
	res := []metadata.Transaction{}
	_ = poolData
	curSize := uint64(0)
	curTime := 0 * time.Millisecond
	for txHash, tx := range poolData.TxByHash {
		if (curSize+poolData.TxInfos[txHash].Size > maxSize) || (curTime+poolData.TxInfos[txHash].VTime > maxTime) {
			continue
		}
		ok, err := tp.Verifier.ValidateWithBlockChain(tx, sView, bcView)
		if err != nil {
			fmt.Printf("Validate tx %v return error %v\n", txHash, err)
		}
		if ok {
			res = append(res, tx)
		}
		tp.Verifier.ValidateTxAndAddToListTxs(
			tx,
			res,
			sView,
			bcView,
			func(txA, txB metadata.Transaction) bool {
				return txA.GetTxFee() > txB.GetTxFee()
			},
		)
	}
	return res
}

func (tp *TxsPool) CheckValidatedTxs(
	txs []metadata.Transaction,
) (
	valid []metadata.Transaction,
	needValidate []metadata.Transaction,
) {
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

// func (tp *TxsPool) removeTxs(tp)
