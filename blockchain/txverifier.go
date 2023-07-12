package blockchain

import (
	"fmt"
	"math"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/txpool"
	"github.com/pkg/errors"
)

type TxsVerifier struct {
	txDB   *statedb.StateDB
	txPool txpool.TxPool

	whitelist map[string]interface{}

	feeEstimator FeeEstimator
}

func (v *TxsVerifier) UpdateTransactionStateDB(
	newSDB *statedb.StateDB,
) {
	v.txDB = newSDB
}

func (v *TxsVerifier) UpdateFeeEstimator(
	estimator txpool.FeeEstimator,
) {
	v.feeEstimator = estimator
}

func NewTxsVerifier(
	txDB *statedb.StateDB,
	tp txpool.TxPool,
	whitelist map[string]interface{},
	estimator FeeEstimator,
) txpool.TxVerifier {
	return &TxsVerifier{
		txDB:   txDB,
		txPool: tp,

		feeEstimator: estimator,
		whitelist:    whitelist,
	}
}

func (v *TxsVerifier) LoadCommitment(
	tx metadata.Transaction,
	shardViewRetriever metadata.ShardViewRetriever,
) (bool, error) {
	sDB := v.txDB
	if shardViewRetriever != nil {
		sDB = shardViewRetriever.GetCopiedTransactionStateDB()
	}
	err := tx.LoadData(sDB.Copy())
	if err != nil {
		Logger.log.Errorf("Can not load commitment of this tx %v, error: %v\n", tx.Hash().String(), err)
		return false, err
	}
	return true, nil
}

func (v *TxsVerifier) PrepareDataForTxs(
	validTxs []metadata.Transaction,
	newTxs []metadata.Transaction,
	shardViewRetriever metadata.ShardViewRetriever,
) (bool, error) {
	sDB := v.txDB
	if shardViewRetriever != nil {
		sDB = shardViewRetriever.GetCopiedTransactionStateDB()
	}
	for _, tx := range validTxs {
		err := tx.CheckData(sDB.Copy())
		if err != nil {
			err = errors.Errorf("Can not load commitment of this tx %v, error: %v\n", tx.Hash().String(), err)
			return false, err
		}
	}
	for _, tx := range newTxs {
		err := tx.LoadData(sDB.Copy())
		if err != nil {
			err = errors.Errorf("Can not load commitment of this tx %v, error: %v\n", tx.Hash().String(), err)
			return false, err
		}
	}
	return true, nil
}

func GetFeeInfo(md metadata.Metadata, feeEstimator FeeEstimator) (minFeePerKB uint64, minFeePerTx uint64) {
	minFeePerKB = feeEstimator.GetLimitFeeForNativeToken()
	minFeePerTx = feeEstimator.GetMinFeePerTx()
	if md != nil {
		mdType := md.GetType()
		txFeeType := metadataCommon.GetSpecifiedFeeTxType(mdType)

		switch txFeeType {
		case metadataCommon.SpecifiedTxType1:
			minFeePerTx = feeEstimator.GetSpecifiedFeeTx()
		case metadataCommon.SpecifiedTxType2:
			minFeePerKB = feeEstimator.GetSpecifiedFeePerKBType2()
			minFeePerTx = feeEstimator.GetSpecifiedFeePerTxType2()
		}
	}

	return
}

func (v *TxsVerifier) checkFees(
	beaconHeight uint64,
	tx metadata.Transaction,
	beaconStateDB *statedb.StateDB,
	shardID byte,
) bool {
	Logger.log.Info("Beacon heigh for checkFees: ", beaconHeight, tx.Hash().String())
	txType := tx.GetType()
	limitFee, minFeePerTx := GetFeeInfo(tx.GetMetadata(), v.feeEstimator)

	if txType == common.TxCustomTokenPrivacyType {
		// check transaction fee for meta data
		meta := tx.GetMetadata()
		// verify at metadata level
		if meta != nil {
			ok := meta.CheckTransactionFee(tx, limitFee, minFeePerTx, int64(beaconHeight), beaconStateDB)
			if !ok {
				Logger.log.Errorf("Error: %+v", fmt.Errorf("transaction %+v: Invalid fee metadata",
					tx.Hash().String()))
			}
			return ok
		}
		// verify at transaction level
		tokenID := tx.GetTokenID()
		feeNativeToken := tx.GetTxFee()
		feePToken := tx.GetTxFeeToken()
		//convert fee in Ptoken to fee in native token (if feePToken > 0)
		if feePToken > 0 {
			feePTokenToNativeTokenTmp, err := metadata.ConvertPrivacyTokenToNativeToken(feePToken, tokenID, int64(beaconHeight), beaconStateDB)
			if err != nil {
				Logger.log.Errorf("ERROR: %+v", fmt.Errorf("transaction %+v: %+v %v can not convert to native token %+v",
					tx.Hash().String(), feePToken, tokenID, err))
				return false
			}

			feePTokenToNativeToken := uint64(math.Ceil(feePTokenToNativeTokenTmp))
			feeNativeToken += feePTokenToNativeToken
		}

		// check min fee of tx
		if feeNativeToken < minFeePerTx {
			Logger.log.Errorf("ERROR: %+v",
				fmt.Errorf("transaction %+v has %d fees PRV which is under the required min fee per tx %d",
					tx.Hash().String(), feeNativeToken, minFeePerTx))
			return false
		}

		// get limit fee in native token
		actualTxSize := tx.GetTxActualSize()
		// check fee in native token
		minFee := actualTxSize * limitFee
		if feeNativeToken < minFee {
			Logger.log.Errorf("ERROR: %+v", fmt.Errorf("transaction %+v has %d fees PRV which is under the required amount of %d, tx size %d",
				tx.Hash().String(), feeNativeToken, minFee, actualTxSize))
			return false
		}
	} else {
		// This is a normal tx -> only check like normal tx with PRV
		txFee := tx.GetTxFee()
		// txNormal := tx.(*transaction.Tx)
		if limitFee > 0 {
			meta := tx.GetMetadata()
			if meta != nil {
				ok := tx.GetMetadata().CheckTransactionFee(tx, limitFee, minFeePerTx, int64(beaconHeight), beaconStateDB)
				if !ok {
					Logger.log.Errorf("ERROR: %+v", fmt.Errorf("transaction %+v has %d fees which is under the required amount of %d",
						tx.Hash().String(), txFee, limitFee*tx.GetTxActualSize()))
				}
				return ok
			}

			// check min fee of tx
			if txFee < minFeePerTx {
				Logger.log.Errorf("ERROR: %+v",
					fmt.Errorf("transaction %+v has %d fees PRV which is under the required min fee per tx %d",
						tx.Hash().String(), txFee, minFeePerTx))
				return false
			}

			fullFee := limitFee * tx.GetTxActualSize()
			// ok := tx.CheckTransactionFee(limitFee)
			if txFee < fullFee {
				Logger.log.Errorf("ERROR: %+v", fmt.Errorf("transaction %+v has %d fees which is under the required amount of %d",
					tx.Hash().String(), txFee, limitFee*tx.GetTxActualSize()))
				return false
			}
		}
	}
	return true
}

func (v *TxsVerifier) ValidateWithoutChainstate(tx metadata.Transaction) (bool, error) {
	ok, err := tx.ValidateSanityDataByItSelf()
	if !ok || err != nil {
		return ok, err
	}
	return tx.ValidateTxCorrectness(v.txDB.Copy())
}

func (v *TxsVerifier) ValidateWithChainState(
	tx metadata.Transaction,
	chainRetriever metadata.ChainRetriever,
	shardViewRetriever metadata.ShardViewRetriever,
	beaconViewRetriever metadata.BeaconViewRetriever,
	beaconHeight uint64,
) (bool, error) {
	ok := v.checkFees(
		beaconViewRetriever.GetHeight(),
		tx,
		beaconViewRetriever.GetBeaconFeatureStateDB(),
		shardViewRetriever.GetShardID(),
	)
	if !ok {
		err := errors.Errorf(" This list txs contains a invalid tx %v, validate result %v, error %v", tx.Hash().String(), ok, errors.Errorf("Transaction fee %v PRV %v Token is invalid", tx.GetTxFee(), tx.GetTxFeeToken()))
		return ok, err
	}
	ok, err := tx.ValidateSanityDataWithBlockchain(
		chainRetriever,
		shardViewRetriever,
		beaconViewRetriever,
		beaconHeight,
	)
	if !ok || err != nil {
		return ok, err
	}
	txDB := shardViewRetriever.GetCopiedTransactionStateDB()
	if meta := tx.GetMetadata(); meta != nil {
		ok, err = meta.ValidateTxWithBlockChain(
			tx,
			chainRetriever,
			shardViewRetriever,
			beaconViewRetriever,
			shardViewRetriever.GetShardID(),
			txDB,
		)
		if err != nil {
			return false, err
		}
	}
	tokenID := tx.GetValidationEnv().TokenID()
	err = tx.ValidateDoubleSpendWithBlockchain(byte(tx.GetValidationEnv().ShardID()), txDB, &tokenID)
	return err == nil, err
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
	Logger.log.Infof("Total txs %v\n", len(txs))
	txs = v.FilterWhitelistTxs(txs)
	if len(txs) == 0 {
		return true, nil
	}
	txsTmp := v.filterSpamStake(txs)
	if len(txsTmp) != len(txs) {
		return false, errors.Errorf("This list txs contain double stake/unstake/stop auto stake for the same key")
	}
	validTxs, newTxs := v.txPool.CheckValidatedTxs(txs)
	errCh := make(chan error)
	doneCh := make(chan interface{}, len(txs)+len(newTxs))
	numOfValidGoroutine := 0
	totalMsgDone := 0
	timeout := time.After(config.Param().BlockTime.MinShardBlockInterval / 2)
	ok, err := v.PrepareDataForTxs(
		validTxs,
		newTxs,
		shardViewRetriever,
	)
	if (!ok) || (err != nil) {
		return false, errors.Errorf("Can not load commitment for this txs, errors %v", err)
	}
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
	for {
		select {
		case err := <-errCh:
			Logger.log.Error(err)
			return false, err
		case <-doneCh:
			numOfValidGoroutine++
			Logger.log.Debugf(" %v %v\n", numOfValidGoroutine, len(txs))
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
					errCh <- errors.Errorf("This list txs contains a invalid tx %v, validate result %v, error %v", target.Hash().String(), ok, err)
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
					errCh <- errors.Errorf("This list txs contains a invalid tx %v, validate result %v, error %v", target.Hash().String(), ok, err)
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

func (v *TxsVerifier) filterSpamStake(
	transactions []metadata.Transaction,
) []metadata.Transaction {
	res := []metadata.Transaction{}
	spam := map[string]interface{}{}
	for _, tx := range transactions {
		metaType := tx.GetMetadataType()
		pk := ""
		switch metaType {
		case metadata.AddStakingMeta:
			if meta, ok := tx.GetMetadata().(*metadata.AddStakingMetadata); ok {
				pk = meta.CommitteePublicKey
			}
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
			if _, existed := spam[pk]; existed {
				continue
			}
			spam[pk] = nil
		}
		res = append(res, tx)
	}
	return res
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
			if _, ok := mapForChkDbSpend[iCoin.GetKeyImage().ToBytes()]; ok {
				return false, errors.Errorf("List txs contain double spend tx %v", tx.Hash().String())
			} else {
				mapForChkDbSpend[iCoin.GetKeyImage().ToBytes()] = nil
			}
		}
		for _, oCoin := range oCoins {
			coinID := oCoin.GetCoinID()
			if _, ok := mapForChkDbSpend[coinID]; ok {
				if common.IsPublicKeyBurningAddress(oCoin.GetPublicKey().ToBytesS()) {
					continue
				}
				return false, errors.Errorf("List txs contain double spend tx %v", tx.Hash().String())
			} else {
				mapForChkDbSpend[coinID] = nil
			}
		}
		if (tx.GetType() == common.TxCustomTokenPrivacyType) || (tx.GetType() == common.TxTokenConversionType) {
			txNormal := tx.(transaction.TransactionToken).GetTxTokenData().TxNormal
			normalPrf := txNormal.GetProof()
			if normalPrf == nil {
				continue
			}
			iCoins := normalPrf.GetInputCoins()
			oCoins := normalPrf.GetOutputCoins()
			for _, iCoin := range iCoins {
				if _, ok := mapForChkDbSpend[iCoin.GetKeyImage().ToBytes()]; ok {
					return false, errors.Errorf("List txs contain double spend tx %v", tx.Hash().String())
				} else {
					mapForChkDbSpend[iCoin.GetKeyImage().ToBytes()] = nil
				}
			}
			for _, oCoin := range oCoins {
				coinID := oCoin.GetCoinID()
				if _, ok := mapForChkDbSpend[coinID]; ok {
					if common.IsPublicKeyBurningAddress(oCoin.GetPublicKey().ToBytesS()) {
						continue
					}
					return false, errors.Errorf("List txs contain double spend tx %v", tx.Hash().String())
				} else {
					mapForChkDbSpend[coinID] = nil
				}
			}
		}
	}
	return true, nil
}
