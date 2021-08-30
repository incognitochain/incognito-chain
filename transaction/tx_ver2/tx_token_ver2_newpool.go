package tx_ver2

import (
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/transaction/tx_generic"
	"github.com/incognitochain/incognito-chain/transaction/utils"
)

func (txToken *TxToken) LoadCommitment(*statedb.StateDB) error {
	return nil
}

func (txToken *TxToken) ValidateSanityDataByItSelf() (bool, error) {
	if txToken.GetType() != common.TxCustomTokenPrivacyType && txToken.GetType() != common.TxTokenConversionType {
		return false, utils.NewTransactionErr(utils.InvalidSanityDataPrivacyTokenError, errors.New("txCustomTokenPrivacy.Tx should have type tp"))
	}
	txn, ok := txToken.GetTxNormal().(*Tx)
	if !ok || txn == nil {
		return false, utils.NewTransactionErr(utils.InvalidSanityDataPrivacyTokenError, errors.New("TX token must have token component"))
	}
	if txToken.GetTxBase().GetProof() == nil && txn.GetProof() == nil {
		return false, errors.New("Tx Privacy Ver 2 must have a proof")
	}
	if txToken.GetTokenID().String() == common.PRVCoinID.String() {
		return false, utils.NewTransactionErr(utils.InvalidSanityDataPrivacyTokenError, errors.New("cannot transfer PRV via txtoken"))
	}
	check, err := txn.TxBase.ValidateSanityDataByItSelf()
	if !check || err != nil {
		return false, utils.NewTransactionErr(utils.InvalidSanityDataPrivacyTokenError, err)
	}
	check, err = txToken.Tx.TxBase.ValidateSanityDataByItSelf()
	if !check || err != nil {
		return false, utils.NewTransactionErr(utils.InvalidSanityDataPrivacyTokenError, err)
	}
	return true, nil
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

	// validate for pToken
	switch txToken.TokenData.Type {
	case utils.CustomTokenTransfer:
		if txToken.GetType() == common.TxTokenConversionType {
			valid, err := validateConversionVer1ToVer2(txn, transactionStateDB, byte(shardID), &tokenID)
			return valid, err
		}

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

		resTxTokenData, err := txn.ValidateTxCorrectness(transactionStateDB)
		if err != nil {
			return resTxTokenData, err
		}
		txFeeProof := txToken.Tx.GetProof()
		if txFeeProof == nil {
			return false, errors.New("Missing proof for PRV")
		}

		bpValid, err := txFeeProof.VerifyV2(txn.GetValidationEnv(), 0)

		return bpValid && resTxTokenData, err
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
	sID := common.GetShardIDFromLastByte(txToken.GetSenderAddrLastByte())
	valEnv = tx_generic.WithShardID(valEnv, int(sID))
	txToken.SetValidationEnv(valEnv)
	txNormalValEnv := valEnv.Clone()
	if txToken.GetTxTokenData().Type == utils.CustomTokenInit {
		txNormalValEnv = tx_generic.WithAct(txNormalValEnv, common.TxActInit)
	} else {
		txNormalValEnv = tx_generic.WithAct(txNormalValEnv, common.TxActTranfer)
	}
	if txToken.GetTxTokenData().TxNormal.IsPrivacy() {
		txNormalValEnv = tx_generic.WithPrivacy(txNormalValEnv)
	} else {
		txNormalValEnv = tx_generic.WithNoPrivacy(txNormalValEnv)
	}
	txToken.GetTxTokenData().TxNormal.SetValidationEnv(txNormalValEnv)
	return valEnv
}

func (txToken *TxToken) GetValidationEnv() metadata.ValidationEnviroment {
	return txToken.Tx.GetValidationEnv()
}

func (txToken *TxToken) SetValidationEnv(valEnv metadata.ValidationEnviroment) {
	txToken.Tx.SetValidationEnv(valEnv)
}
