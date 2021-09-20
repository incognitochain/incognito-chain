package tx_generic //nolint:revive

import (
	"fmt"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction/utils"
	"github.com/pkg/errors"
)

func (tx *TxBase) initEnv() metadata.ValidationEnviroment {
	valEnv := DefaultValEnv()
	if tx.IsSalaryTx() {
		valEnv = WithAct(valEnv, common.TxActInit)
	}
	if tx.IsPrivacy() {
		valEnv = WithPrivacy(valEnv)
	} else {
		valEnv = WithNoPrivacy(valEnv)
	}
	valEnv = WithType(valEnv, tx.GetType())
	sID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
	valEnv = WithShardID(valEnv, int(sID))
	valEnv = WithVersion(valEnv, tx.Version)
	valEnv = WithCA(valEnv, false)
	if tx.Version == utils.TxVersion2Number {
		proofAsV2, ok := tx.GetProof().(*privacy.ProofV2)
		if ok {
			if hasCA, err := proofAsV2.IsConfidentialAsset(); err != nil {
				valEnv = WithCA(valEnv, hasCA)
			}
		}
	}
	valEnv = WithTokenID(valEnv, common.PRVCoinID)
	tx.SetValidationEnv(valEnv)
	return valEnv
}

func (tx *TxBase) GetValidationEnv() metadata.ValidationEnviroment {
	return tx.valEnv
}

func (tx *TxBase) SetValidationEnv(vEnv metadata.ValidationEnviroment) {
	if vE, ok := vEnv.(*ValidationEnv); ok {
		tx.valEnv = vE
	} else {
		valEnv := DefaultValEnv()
		if tx.IsPrivacy() {
			valEnv = WithPrivacy(valEnv)
		} else {
			valEnv = WithNoPrivacy(valEnv)
		}
		valEnv = WithType(valEnv, tx.GetType())
		tx.valEnv = valEnv
	}
}

func (tx *TxBase) ValidateSanityDataByItSelf() (bool, error) {
	switch tx.Type {
	case common.TxNormalType, common.TxRewardType, common.TxCustomTokenPrivacyType, common.TxReturnStakingType: //is valid
	default:
		return false, utils.NewTransactionErr(utils.RejectTxType, fmt.Errorf("wrong tx type with %s", tx.Type))
	}

	// check info field
	if len(tx.Info) > 512 {
		return false, utils.NewTransactionErr(utils.RejectTxInfoSize, fmt.Errorf("wrong tx info length %d bytes, only support info with max length <= %d bytes", len(tx.Info), 512))
	}

	if tx.Metadata != nil {
		metaType := tx.Metadata.GetType()
		txType := tx.GetValidationEnv().TxType()
		if !metadata.IsAvailableMetaInTxType(metaType, txType) {
			return false, errors.Errorf("Not mismatch Type, txType: %v, metadataType %v", txType, metaType)
		}
	}

	// check tx size
	actualTxSize := tx.GetTxActualSize()
	if actualTxSize > common.MaxTxSize {
		//fmt.Print(actualTxSize, common.MaxTxSize)
		return false, utils.NewTransactionErr(utils.RejectTxSize, fmt.Errorf("tx size %d kB is too large", actualTxSize))
	}

	//check version
	if tx.Version > utils.TxVersion2Number {
		return false, utils.NewTransactionErr(utils.RejectTxVersion, fmt.Errorf("tx version is %d. Wrong version tx. Only support for version <= %d", tx.Version, utils.CurrentTxVersion))
	}
	// check LockTime before now
	if int64(tx.LockTime) > time.Now().Unix() {
		return false, utils.NewTransactionErr(utils.RejectInvalidLockTime, fmt.Errorf("wrong tx locktime %d", tx.LockTime))
	}

	if len(tx.SigPubKey) != common.SigPubKeySize {
		return false, utils.NewTransactionErr(utils.RejectTxPublickeySigSize, fmt.Errorf("wrong tx Sig PK size %d", len(tx.SigPubKey)))
	}

	metaData := tx.GetMetadata()
	proof := tx.GetProof()
	if metaData != nil {
		if !metaData.ValidateMetadataByItself() {
			return false, errors.Errorf("Metadata is not valid")
		}
	}

	if (proof == nil) || ((len(proof.GetInputCoins()) == 0) && (len(proof.GetOutputCoins()) == 0)) {
		if metaData == nil {
			utils.Logger.Log.Errorf("[invalidtxsanity] This tx %v has no proof, but metadata is nil", tx.Hash().String())
		} else {
			metaType := metaData.GetType()
			if !metadata.NoInputNoOutput(metaType) {
				utils.Logger.Log.Errorf("[invalidtxsanity] This tx %v has no proof, but metadata is invalid, metadata type %v", tx.Hash().String(), metaType)
			}
		}
	} else {
		if len(proof.GetInputCoins()) == 0 {
			if (metaData == nil) && (tx.GetValidationEnv().TxAction() != common.TxActInit) {
				return false, utils.NewTransactionErr(utils.RejectTxType, fmt.Errorf("This tx %v has no input, but metadata is nil", tx.Hash().String()))
			} else {
				if metaData != nil {
					metaType := metaData.GetType()
					if !metadata.NoInputHasOutput(metaType) {
						return false, utils.NewTransactionErr(utils.RejectTxType, fmt.Errorf("This tx %v has no proof, but metadata is invalid, metadata type %v", tx.Hash().String(), metaType))
					}
				}
			}
		}
		// if len(proof.GetOutputCoins()) == 0 {
		// 	if metaData != nil {
		// 		metaType := metaData.GetType()
		// 		if !metadata.HasInputNoOutput(metaType) {
		// 			return false, utils.NewTransactionErr(RejectTxType, fmt.Errorf("This tx %v has no proof, but metadata is invalid, metadata type %v", tx.Hash().String(), metaType))
		// 		}
		// 	}
		// }
		// check sanity of Proof
		validateSanityOfProof, err := tx.validateSanityDataOfProof()
		if err != nil || !validateSanityOfProof {
			return false, err
		}
	}

	return true, nil
}

func (tx *TxBase) validateSanityDataOfProof() (bool, error) {
	if tx.Proof != nil {
		return tx.Proof.ValidateSanity(tx.valEnv)
	}
	return false, errors.Errorf("Proof of tx %v is nil", tx.Hash().String())
}
