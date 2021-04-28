package blockchain

import (
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/txpool"
	"github.com/pkg/errors"
)

type TxsVerifier struct {
	txDB   *statedb.StateDB
	txPool txpool.TxPool

	whitelist map[string]interface{}
}

func (v *TxsVerifier) UpdateTransactionStateDB(
	newSDB *statedb.StateDB,
) {
	v.txDB = newSDB
}

func NewTxsVerifier(
	txDB *statedb.StateDB,
	tp txpool.TxPool,
	whitelist map[string]interface{},
) txpool.TxVerifier {
	return &TxsVerifier{
		txDB:   txDB,
		txPool: tp,

		whitelist: whitelist,
	}
}

func (v *TxsVerifier) LoadCommitment(
	tx metadata.Transaction,
	shardViewRetriever metadata.ShardViewRetriever,
) bool {
	sDB := v.txDB
	if shardViewRetriever != nil {
		sDB = shardViewRetriever.GetCopiedTransactionStateDB()
	}
	Logger.log.Infof("[debugtxs] %v %v %v\n", tx, tx.Hash().String(), sDB)
	Logger.log.Infof("[debugtxs] %v\n", tx.GetValidationEnv().IsPrivacy())
	err := tx.LoadCommitment(sDB.Copy())
	if err != nil {
		Logger.log.Errorf("Can not load commitment of this tx %v, error: %v\n", tx.Hash().String(), err)
		return false
	}
	return true
}

func (v *TxsVerifier) LoadCommitmentForTxs(
	txs []metadata.Transaction,
	shardViewRetriever metadata.ShardViewRetriever,
) bool {
	sDB := v.txDB
	if shardViewRetriever != nil {
		sDB = shardViewRetriever.GetCopiedTransactionStateDB()
	}
	for _, tx := range txs {
		err := tx.LoadCommitment(sDB.Copy())
		if err != nil {
			Logger.log.Errorf("[testNewPool] Can not load commitment of this tx %v, error: %v\n", tx.Hash().String(), err)
			return false
		}
	}
	return true
}

func (v *TxsVerifier) ValidateTxsSig(
	txs []metadata.Transaction,
	errCh chan error,
	doneCh chan interface{},
) {
	for _, tx := range txs {
		go func(target metadata.Transaction) {
			ok, err := target.VerifySigTx()
			if !ok || err != nil {
				if errCh != nil {
					errCh <- errors.Errorf("Signature of tx %v is not valid, result %v, error %v", target.Hash().String(), ok, err)
				}
			} else {
				if doneCh != nil {
					doneCh <- nil
				}
			}
		}(tx)
	}
}

func (v *TxsVerifier) ValidateWithoutChainstate(tx metadata.Transaction) (bool, error) {
	ok, err := tx.ValidateSanityDataByItSelf()
	if !ok || err != nil {
		return ok, err
	}
	return tx.ValidateTxCorrectness()
}

func (v *TxsVerifier) ValidateWithChainState(
	tx metadata.Transaction,
	chainRetriever metadata.ChainRetriever,
	shardViewRetriever metadata.ShardViewRetriever,
	beaconViewRetriever metadata.BeaconViewRetriever,
	beaconHeight uint64,
) (bool, error) {
	ok, err := tx.ValidateSanityDataWithBlockchain(
		chainRetriever,
		shardViewRetriever,
		beaconViewRetriever,
		beaconHeight,
	)
	if !ok || err != nil {
		return ok, err
	}
	return tx.ValidateDoubleSpendWithBlockChain(shardViewRetriever.GetCopiedTransactionStateDB())
}

func (v *TxsVerifier) FilterWhitelistTxs(txs []metadata.Transaction) []metadata.Transaction {
	j := 0
	res := make([]metadata.Transaction, len(txs))
	for i, tx := range txs {
		if _, ok := v.whitelist[tx.Hash().String()]; !ok {
			res[j] = txs[i]
			j++
		}
	}
	return res[:j]
}

func (v *TxsVerifier) FullValidateTransactions(
	chainRetriever metadata.ChainRetriever,
	shardViewRetriever metadata.ShardViewRetriever,
	beaconViewRetriever metadata.BeaconViewRetriever,
	txs []metadata.Transaction,
) (bool, error) {
	Logger.log.Infof("[testNewPool] Total txs %v\n", len(txs))
	if len(txs) == 0 {
		return true, nil
	}
	txs = v.FilterWhitelistTxs(txs)
	_, newTxs := v.txPool.CheckValidatedTxs(txs)
	// fmt.Println("Is Validated")
	errCh := make(chan error)
	doneCh := make(chan interface{}, len(txs)+2*len(newTxs))
	numOfValidGoroutine := 0
	totalMsgDone := 0
	timeout := time.After(10 * time.Second)
	v.LoadCommitmentForTxs(
		txs,
		shardViewRetriever,
	)
	v.ValidateTxsSig(
		newTxs,
		errCh,
		doneCh,
	)
	totalMsgDone += len(newTxs)
	v.validateTxsWithoutChainstate(
		newTxs,
		errCh,
		doneCh,
	)
	totalMsgDone += len(newTxs)
	v.validateTxsWithChainstate(
		txs,
		chainRetriever,
		shardViewRetriever,
		beaconViewRetriever,
		errCh,
		doneCh,
	)
	totalMsgDone += len(txs)
	// fmt.Println("[testNewPool] wait!")
	for {
		select {
		case err := <-errCh:
			Logger.log.Error(err)
			return false, err
		case <-doneCh:
			numOfValidGoroutine++
			Logger.log.Infof("[testNewPool] %v %v\n", numOfValidGoroutine, len(txs))
			if numOfValidGoroutine == totalMsgDone {
				ok, err := v.checkDoubleSpendInListTxs(txs)
				if (!ok) || (err != nil) {
					Logger.log.Error(err)
					return false, err
				}
				return true, nil
			}
		case <-timeout:
			Logger.log.Error("Timeout!!!")
			return false, errors.Errorf("Validate %v txs timeout", len(txs))
		}
	}
}

func (v *TxsVerifier) validateTxsWithoutChainstate(
	txs []metadata.Transaction,
	errCh chan error,
	doneCh chan interface{},
) {
	for _, tx := range txs {
		go func(target metadata.Transaction) {
			ok, err := v.ValidateWithoutChainstate(target)
			if !ok || err != nil {
				if errCh != nil {
					errCh <- errors.Errorf("[testNewPool] This list txs contains a invalid tx %v, validate result %v, error %v", target.Hash().String(), ok, err)
				}
			} else {
				if doneCh != nil {
					doneCh <- nil
				}
			}
		}(tx)
	}
}

func (v *TxsVerifier) validateTxsWithChainstate(
	txs []metadata.Transaction,
	cView metadata.ChainRetriever,
	sView metadata.ShardViewRetriever,
	bcView metadata.BeaconViewRetriever,
	errCh chan error,
	doneCh chan interface{},
) {
	// MAX := runtime.NumCPU() - 1
	// nWorkers := make(chan int, MAX)
	for _, tx := range txs {
		if tx.Hash().String() == "9d0e017131c8d28d66c190484b4ea804859da1f8346280a0f279119670c0307d" { //Mainnet whitelist
			continue
		}
		// nWorkers <- 1
		go func(target metadata.Transaction) {
			ok, err := v.ValidateWithChainState(
				target,
				cView,
				sView,
				bcView,
				sView.GetBeaconHeight(),
			)
			if !ok || err != nil {
				if errCh != nil {
					errCh <- errors.Errorf("[testNewPool] This list txs contains a invalid tx %v, validate result %v, error %v", target.Hash().String(), ok, err)
				}
			} else {
				if doneCh != nil {
					doneCh <- nil
				}
			}
			// <-nWorkers
		}(tx)
	}
}

func (v *TxsVerifier) checkDoubleSpendInListTxs(
	txs []metadata.Transaction,
) (
	bool,
	error,
) {
	mapForChkDbSpend := map[[privacy.Ed25519KeySize]byte]interface{}{}
	for _, tx := range txs {

		prf := tx.GetProof()
		if prf == nil {
			continue
		}
		iCoins := prf.GetInputCoins()
		oCoins := prf.GetOutputCoins()
		for _, iCoin := range iCoins {
			if _, ok := mapForChkDbSpend[iCoin.CoinDetails.GetSerialNumber().ToBytes()]; ok {
				return false, errors.Errorf("List txs contain double spend tx %v", tx.Hash().String())
			} else {
				mapForChkDbSpend[iCoin.CoinDetails.GetSerialNumber().ToBytes()] = nil
			}
		}
		for _, oCoin := range oCoins {
			if _, ok := mapForChkDbSpend[oCoin.CoinDetails.GetSNDerivator().ToBytes()]; ok {
				return false, errors.Errorf("List txs contain double spend tx %v", tx.Hash().String())
			} else {
				mapForChkDbSpend[oCoin.CoinDetails.GetSNDerivator().ToBytes()] = nil
			}
		}
		if tx.GetType() == common.TxCustomTokenPrivacyType {
			txNormal := tx.(*transaction.TxCustomTokenPrivacy).TxPrivacyTokenData.TxNormal
			normalPrf := txNormal.GetProof()
			if normalPrf == nil {
				continue
			}
			iCoins := normalPrf.GetInputCoins()
			oCoins := normalPrf.GetOutputCoins()
			for _, iCoin := range iCoins {
				if _, ok := mapForChkDbSpend[iCoin.CoinDetails.GetSerialNumber().ToBytes()]; ok {
					return false, errors.Errorf("List txs contain double spend tx %v", tx.Hash().String())
				} else {
					mapForChkDbSpend[iCoin.CoinDetails.GetSerialNumber().ToBytes()] = nil
				}
			}
			for _, oCoin := range oCoins {
				if _, ok := mapForChkDbSpend[oCoin.CoinDetails.GetSNDerivator().ToBytes()]; ok {
					return false, errors.Errorf("List txs contain double spend tx %v", tx.Hash().String())
				} else {
					mapForChkDbSpend[oCoin.CoinDetails.GetSNDerivator().ToBytes()] = nil
				}
			}
		}
	}
	return true, nil
}
