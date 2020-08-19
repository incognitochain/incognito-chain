package transaction

import (
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/wallet"
	"time"
)

func verifyTxCreatedByMiner(tx metadata.Transaction, mintdata *metadata.MintData, shardID byte, bcr metadata.ChainRetriever, accumulatedValues *metadata.AccumulatedValues, retriever metadata.ShardViewRetriever, viewRetriever metadata.BeaconViewRetriever) (bool, error)  {
	if tx.IsPrivacy() {
		return true, nil
	}
	proof := tx.GetProof()
	meta := tx.GetMetadata()

	inputCoins := make([]coin.PlainCoin, 0)
	outputCoins := make([]coin.Coin, 0)
	if tx.GetProof() != nil {
		inputCoins = tx.GetProof().GetInputCoins()
		outputCoins = tx.GetProof().GetOutputCoins()
	}
	if proof != nil && len(inputCoins) == 0 && len(outputCoins) > 0 { // coinbase tx
		if meta == nil {
			return false, nil
		}
		if !meta.IsMinerCreatedMetaType() {
			return false, nil
		}
	}
	if meta != nil {
		ok, err := meta.VerifyMinerCreatedTxBeforeGettingInBlock(mintdata, shardID, tx, bcr, accumulatedValues, retriever, viewRetriever)
		if err != nil {
			Logger.Log.Error(err)
			return false, NewTransactionErr(VerifyMinerCreatedTxBeforeGettingInBlockError, err)
		}
		return ok, nil
	}
	return true, nil
}
func getTxMintData(tx metadata.Transaction, tokenID *common.Hash) (bool, coin.Coin, *common.Hash, error) {
	outputCoins, err := tx.GetReceiverData()
	if err != nil {
		Logger.Log.Error("GetTxMintData: Cannot get receiver data")
		return false, nil, nil, err
	}
	if len(outputCoins) != 1 {
		Logger.Log.Error("GetTxMintData : Should only have one receiver")
		return false, nil, nil, errors.New("Error Tx mint has more than one receiver")
	}
	if inputCoins := tx.GetProof().GetInputCoins(); len(inputCoins) > 0 {
		return false, nil, nil, errors.New("Error this is not Tx mint")
	}
	return true, outputCoins[0], tokenID, nil
}

func getTxBurnData(tx metadata.Transaction) (bool, coin.Coin, *common.Hash, error) {
	outputCoins, err := tx.GetReceiverData()
	if err != nil {
		Logger.Log.Errorf("Cannot get receiver data, error %v", err)
		return false, nil, nil, err
	}
	if len(outputCoins) > 2 {
		Logger.Log.Error("GetAndCheckBurning receiver: More than 2 receivers")
		return false, nil, nil, err
	}
	for _, coin := range outputCoins {
		if wallet.IsPublicKeyBurningAddress(coin.GetPublicKey().ToBytesS()) {
			return true, coin, &common.PRVCoinID, nil
		}
	}
	return false, nil, nil, nil
}

func validateTxWithBlockChain(tx metadata.Transaction, chainRetriever metadata.ChainRetriever, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever, shardID byte, stateDB *statedb.StateDB) error {
	if tx.GetType() == common.TxRewardType || tx.GetType() == common.TxReturnStakingType {
		return nil
	}
	meta := tx.GetMetadata()
	if meta != nil {
		isContinued, err := meta.ValidateTxWithBlockChain(tx, chainRetriever, shardViewRetriever, beaconViewRetriever, shardID, stateDB)
		fmt.Printf("[transactionStateDB] validate metadata with blockchain: %d %h %t %v\n", tx.GetMetadataType(), tx.Hash(), isContinued, err)
		if err != nil {
			Logger.Log.Errorf("[db] validate metadata with blockchain: %d %s %t %v\n", tx.GetMetadataType(), tx.Hash().String(), isContinued, err)
			return NewTransactionErr(RejectTxMedataWithBlockChain, fmt.Errorf("validate metadata of tx %s with blockchain error %+v", tx.Hash().String(), err))
		}
		if !isContinued {
			return nil
		}
	}
	return tx.ValidateDoubleSpendWithBlockchain(shardID, stateDB, nil)
}

func validateTxByItself(tx metadata.Transaction, hasPrivacy bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, shardID byte, isNewTransaction bool) (bool, error) {
	prvCoinID := &common.Hash{}
	err := prvCoinID.SetBytes(common.PRVCoinID[:])
	if err != nil {
		return false, err
	}
	ok, err := tx.ValidateTransaction(hasPrivacy, transactionStateDB, bridgeStateDB, shardID, prvCoinID, false, isNewTransaction)
	if !ok {
		return false, err
	}
	meta := tx.GetMetadata()
	if meta != nil {
		if hasPrivacy && tx.GetVersion() == 1{
			return false, errors.New("Metadata can not exist in not privacy tx")
		}
		validateMetadata := meta.ValidateMetadataByItself()
		if validateMetadata {
			return validateMetadata, nil
		} else {
			return validateMetadata, NewTransactionErr(UnexpectedError, errors.New("Metadata is invalid"))
		}
	}
	return true, nil
}

func validateTransaction(tx metadata.Transaction, hasPrivacy bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isBatch bool, isNewTransaction bool) (bool, error) {
	switch tx.GetType() {
	case common.TxRewardType:
		return tx.ValidateTxSalary(transactionStateDB)
	case common.TxReturnStakingType:
		return tx.ValidateTxReturnStaking(transactionStateDB), nil
	case common.TxConversionType:
		return validateConversionVer1ToVer2(tx, transactionStateDB, shardID, tokenID)
	}
	return tx.Verify(hasPrivacy, transactionStateDB, bridgeStateDB, shardID, tokenID, isBatch, isNewTransaction)
}

func validateSanityMetadata(tx metadata.Transaction, chainRetriever metadata.ChainRetriever, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever, beaconHeight uint64) ( bool, error) {
	meta := tx.GetMetadata()
	if meta != nil {
		isValid, ok, err := meta.ValidateSanityData(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight, tx )
		if err != nil || !ok || !isValid {
			return ok, err
		}
	}
	return true, nil
}

func validateSanityTxWithoutMetadata(tx metadata.Transaction, chainRetriever metadata.ChainRetriever, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever, beaconHeight uint64) (bool, error) {
	//check version
	if tx.GetVersion() > TxVersion2Number {
		return false, NewTransactionErr(RejectTxVersion, fmt.Errorf("tx version is %d. Wrong version tx. Only support for version <= %d", tx.GetVersion(), currentTxVersion))
	}
	// check LockTime before now
	if tx.GetLockTime() > time.Now().Unix() {
		return false, NewTransactionErr(RejectInvalidLockTime, fmt.Errorf("wrong tx locktime %d", tx.GetLockTime()))
	}

	// check tx size
	actualTxSize := tx.GetTxActualSize()
	if actualTxSize > common.MaxTxSize {
		return false, NewTransactionErr(RejectTxSize, fmt.Errorf("tx size %d kB is too large", actualTxSize))
	}

	// check sanity of Proof
	if tx.GetProof() != nil {
		ok, err := tx.GetProof().ValidateSanity()
		if !ok || err != nil {
			s := ""
			if !ok {
				s = fmt.Sprintf("ValidateSanity Proof got error: ok = false; %s\n", err.Error())
			} else {
				s = fmt.Sprintf("ValidateSanity Proof got error: error %s\n", err.Error())
			}
			return false, errors.New(s)
		}
	}

	// check Type is normal or salary tx
	switch tx.GetType() {
	case common.TxNormalType, common.TxRewardType, common.TxCustomTokenPrivacyType, common.TxTokenConversionType, common.TxReturnStakingType, common.TxConversionType: //is valid
	default:
		return false, NewTransactionErr(RejectTxType, fmt.Errorf("wrong tx type with %s", tx.GetType()))
	}

	// check info field
	info := tx.GetInfo()
	if len(info) > MaxSizeInfo {
		return false, NewTransactionErr(RejectTxInfoSize, fmt.Errorf("wrong tx info length %d bytes, only support info with max length <= %d bytes", len(info), 512))
	}
	return true, nil
}

func getTxActualSizeInBytes(tx metadata.Transaction) uint64{
	if tx == nil {
		return uint64(0)
	}
	var sizeTx = uint64(0)
	txTokenBase, ok := tx.(TxTokenBase)
	if ok { //TxTokenBase
		sizeTx += getTxActualSizeInBytes(txTokenBase.Tx)

		if &txTokenBase.TxTokenData != nil {
			sizeTx += getTxActualSizeInBytes(txTokenBase.TxTokenData.TxNormal)
			sizeTx += uint64(len(txTokenBase.TxTokenData.PropertyName))
			sizeTx += uint64(len(txTokenBase.TxTokenData.PropertySymbol))
			sizeTx += uint64(len(txTokenBase.TxTokenData.PropertyID))
			sizeTx += 4 // Type
			sizeTx += 1 // Mintable
			sizeTx += 8 // Amount
		}
		meta := txTokenBase.GetMetadata()
		if meta != nil {
			sizeTx += meta.CalculateSize()
		}

		return sizeTx
	}else{ //TxBase
		sizeTx += uint64(1) //version
		sizeTx += uint64(len(tx.GetType()) + 1) //type string
		sizeTx += uint64(8) //locktime
		sizeTx += uint64(8) //fee
		sizeTx += uint64(len(tx.GetInfo())) //info

		sizeTx += uint64(len(tx.GetSigPubKey())) //sigpubkey
		sizeTx += uint64(len(tx.GetSig())) //signature
		sizeTx += uint64(1) //pubkeylastbytesender

		//paymentproof
		if tx.GetProof() != nil {
			sizeTx += uint64(len(tx.GetProof().Bytes()))
		}

		//metadata
		if tx.GetMetadata() != nil {
			sizeTx += tx.GetMetadata().CalculateSize()
		}

		return sizeTx
	}
}