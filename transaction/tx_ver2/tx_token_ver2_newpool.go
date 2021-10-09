package tx_ver2

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/transaction/tx_generic"
	"github.com/incognitochain/incognito-chain/transaction/utils"
)

func (txToken *TxToken) LoadData(transactionStateDB *statedb.StateDB) error {
	if txToken.TokenData.Type == utils.CustomTokenTransfer {
		txn := txToken.GetTxNormal()
		if txToken.GetType() == common.TxTokenConversionType {
			if err := checkInputInDB(txn, transactionStateDB); err != nil {
				return err
			}
		} else {
			err := txn.LoadData(transactionStateDB)
			if err != nil {
				return err
			}
		}
	}
	return txToken.Tx.LoadData(transactionStateDB)
}

func (txToken *TxToken) CheckData(transactionStateDB *statedb.StateDB) error {
	if txToken.TokenData.Type == utils.CustomTokenTransfer {
		txn := txToken.GetTxNormal()
		if txToken.GetType() == common.TxTokenConversionType {
			if err := checkInputInDB(txn, transactionStateDB); err != nil {
				return err
			}
		} else {
			err := txn.CheckData(transactionStateDB)
			if err != nil {
				return err
			}
		}
	}
	return txToken.Tx.CheckData(transactionStateDB)
}

func (txToken *TxToken) ValidateSanityDataWithMetadata() (bool, error) {
	metaData := txToken.GetMetadata()
	if metaData != nil {
		metaType := metaData.GetType()
		txType := txToken.GetValidationEnv().TxType()
		if !metadata.IsAvailableMetaInTxType(metaType, txType) {
			return false, errors.Errorf("Not mismatch Type, txType: %v, metadataType %v", txType, metaType)
		}
		if !metaData.ValidateMetadataByItself() {
			return false, errors.Errorf("Metadata is not valid")
		}
	}
	txn := txToken.GetTxNormal()
	if txn == nil {
		return false, errors.Errorf("Can not get tx normal for tx %v", txToken.Hash().String())
	}
	proof := txn.GetProof()
	if (proof == nil) || ((len(proof.GetInputCoins()) == 0) && (len(proof.GetOutputCoins()) == 0)) {
		if metaData == nil {
			utils.Logger.Log.Errorf("[invalidtxsanity] This tx %v has no proof, but metadata is nil", txToken.Hash().String())
		} else {
			metaType := metaData.GetType()
			if !metadata.NoInputNoOutput(metaType) {
				utils.Logger.Log.Errorf("[invalidtxsanity] This tx %v has no proof, but metadata is invalid, metadata type %v", txToken.Hash().String(), metaType)
			}
		}
	} else {
		if len(proof.GetInputCoins()) == 0 {
			if (metaData == nil) && (txToken.GetValidationEnv().TxAction() != common.TxActInit) {
				return false, utils.NewTransactionErr(utils.RejectTxType, fmt.Errorf("This tx %v has no input, but metadata is nil", txToken.Hash().String()))
			}
			if metaData != nil {
				metaType := metaData.GetType()
				if !metadata.NoInputHasOutput(metaType) {
					return false, utils.NewTransactionErr(utils.RejectTxType, fmt.Errorf("This tx %v has no proof, but metadata is invalid, metadata type %v", txToken.Hash().String(), metaType))
				}
			}

		}
	}
	proof = txToken.Tx.GetProof()
	if proof != nil {
		if len(proof.GetInputCoins()) == 0 {
			return false, utils.NewTransactionErr(utils.RejectTxType, fmt.Errorf("This tx %v for pay fee for tx %v, can not be mint tx", txToken.Tx.Hash().String(), txToken.Hash().String()))
		}
	}
	return true, nil
}

func (txToken *TxToken) ValidateSanityDataByItSelf() (bool, error) {
	if txToken.GetType() != common.TxCustomTokenPrivacyType && txToken.GetType() != common.TxTokenConversionType {
		return false, utils.NewTransactionErr(utils.InvalidSanityDataPrivacyTokenError, errors.New("txCustomTokenPrivacy.Tx should have type tp"))
	}
	txn, ok := txToken.GetTxNormal().(*Tx)
	if !ok || txn == nil {
		return false, utils.NewTransactionErr(utils.InvalidSanityDataPrivacyTokenError, errors.New("TX token must have token component"))
	}
	if txToken.GetTxBase().GetProof() != nil {
		check, err := txToken.Tx.TxBase.ValidateSanityDataByItSelf()
		if !check || err != nil {
			return false, utils.NewTransactionErr(utils.InvalidSanityDataPrivacyTokenError, err)
		}
	}
	if txn.GetProof() == nil {
		return false, errors.New("Tx Privacy Ver 2 must have a proof")
	}
	if txToken.GetTokenID().String() == common.PRVCoinID.String() {
		return false, utils.NewTransactionErr(utils.InvalidSanityDataPrivacyTokenError, errors.New("cannot transfer PRV via txtoken"))
	}
	check, err := txn.TxBase.ValidateSanityDataByItSelf()
	if !check || err != nil {
		return false, utils.NewTransactionErr(utils.InvalidSanityDataPrivacyTokenError, err)
	}
	return txToken.ValidateSanityDataWithMetadata()
}

func (txToken *TxToken) ValidateSanityDataWithBlockchain(
	chainRetriever metadata.ChainRetriever,
	shardViewRetriever metadata.ShardViewRetriever,
	beaconViewRetriever metadata.BeaconViewRetriever,
	beaconHeight uint64,
) (
	bool,
	error,
) {
	meta := txToken.GetMetadata()
	if meta != nil {
		utils.Logger.Log.Debug("tx.Metadata.ValidateSanityData")
		isContinued, ok, err := meta.ValidateSanityData(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight, txToken)
		utils.Logger.Log.Debug("END tx.Metadata.ValidateSanityData")
		if err != nil || !ok || !isContinued {
			return ok, err
		}
	}
	return true, nil
}

func (txToken *TxToken) ValidateTxCorrectness(transactionStateDB *statedb.StateDB) (bool, error) {
	var err error
	vEnv := txToken.GetValidationEnv()
	shardID := vEnv.ShardID()
	tokenID := vEnv.TokenID()
	txn := txToken.GetTxNormal()
	ok, err := txToken.verifySig(transactionStateDB, byte(shardID), &common.PRVCoinID)
	if !ok {
		utils.Logger.Log.Errorf("FAILED VERIFICATION SIGNATURE ver2 (token) with tx hash %s: %+v \n", txToken.Hash().String(), err)
		return false, utils.NewTransactionErr(utils.VerifyTxSigFailError, err)
	}

	if txToken.IsSalaryTx() {
		valid, err := txToken.ValidateTxSalary(transactionStateDB)
		return valid, err
	}
	// validate for pToken
	switch txToken.TokenData.Type {
	case utils.CustomTokenTransfer:
		var resTxToken bool
		if txToken.GetType() == common.TxTokenConversionType {
			resTxToken, err = validateTxConvertCorrectness(txn, transactionStateDB)
			if err != nil {
				return false, err
			}
		} else {

			// This transaction might be a tx burn, we must check its tokenId and assetTag
			if tokenID.String() != common.PRVIDStr || tokenID.String() != common.ConfidentialAssetID.String() {
				isBurned, burnedToken, _, err := txToken.GetTxBurnData()
				if err != nil {
					return false, err
				}
				if isBurned && !operation.IsPointEqual(burnedToken.GetAssetTag(), operation.HashToPoint(tokenID[:])) {
					return false, fmt.Errorf("invalid burned tokenId")
				}
			}

			resTxToken, err = txn.ValidateTxCorrectness(transactionStateDB)
			if err != nil {
				return resTxToken, err
			}
		}
		txFeeProof := txToken.Tx.GetProof()
		if txFeeProof == nil {
			return false, errors.New("Missing proof for PRV")
		}

		bpValid, err := txFeeProof.VerifyV2(txn.GetValidationEnv(), 0)

		return bpValid && resTxToken, err
	default:
		return false, errors.New("Cannot validate Tx Token. Unavailable type")
	}
}

func (txToken *TxToken) initEnv() metadata.ValidationEnviroment {

	valEnv := tx_generic.DefaultValEnv()
	// if txCustomTokenPrivacy.IsSalaryTx() {
	valEnv = tx_generic.WithAct(valEnv, common.TxActTranfer)
	// }
	if txToken.IsPrivacy() {
		valEnv = tx_generic.WithPrivacy(valEnv)
	} else {
		valEnv = tx_generic.WithNoPrivacy(valEnv)
	}

	valEnv = tx_generic.WithType(valEnv, txToken.GetType())
	valEnv = tx_generic.WithTokenID(valEnv, common.PRVCoinID)
	valEnv = tx_generic.WithSigPubkey(valEnv, txToken.Tx.GetSigPubKey())
	sID := common.GetShardIDFromLastByte(txToken.GetSenderAddrLastByte())
	valEnv = tx_generic.WithShardID(valEnv, int(sID))
	txNormalValEnv := valEnv.Clone()
	if txToken.TokenData.Type == utils.CustomTokenInit {
		txNormalValEnv = tx_generic.WithAct(txNormalValEnv, common.TxActInit)
		valEnv = tx_generic.WithAct(valEnv, common.TxActInit)
	} else {
		txNormalValEnv = tx_generic.WithAct(txNormalValEnv, common.TxActTranfer)
	}
	txn := txToken.GetTxNormal()
	proofAsV2, ok := txn.GetProof().(*privacy.ProofV2)
	if (proofAsV2 != nil) && (ok) {
		if hasCA, err := proofAsV2.IsConfidentialAsset(); err != nil {
			valEnv = tx_generic.WithCA(valEnv, hasCA)
		}
	}
	txToken.SetValidationEnv(valEnv)

	if txn.IsPrivacy() {
		txNormalValEnv = tx_generic.WithPrivacy(txNormalValEnv)
	} else {
		txNormalValEnv = tx_generic.WithNoPrivacy(txNormalValEnv)
	}
	txNormalValEnv = tx_generic.WithTokenID(txNormalValEnv, txToken.GetTxTokenData().PropertyID)
	txNormalValEnv = tx_generic.WithSigPubkey(txNormalValEnv, txn.GetSigPubKey())
	txn.SetValidationEnv(txNormalValEnv)
	return valEnv
}

func (txToken *TxToken) GetValidationEnv() metadata.ValidationEnviroment {
	return txToken.Tx.GetValidationEnv()
}

func (txToken *TxToken) SetValidationEnv(valEnv metadata.ValidationEnviroment) {
	if vE, ok := valEnv.(*tx_generic.ValidationEnv); ok {
		txToken.valEnv = vE
	} else {
		valEnv := tx_generic.DefaultValEnv()
		if txToken.IsPrivacy() {
			valEnv = tx_generic.WithPrivacy(valEnv)
		} else {
			valEnv = tx_generic.WithNoPrivacy(valEnv)
		}
		valEnv = tx_generic.WithType(valEnv, txToken.GetType())
		txToken.valEnv = valEnv
	}
}
