package tx_ver1

import (
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	errhandler "github.com/incognitochain/incognito-chain/privacy/errorhandler"
	"github.com/incognitochain/incognito-chain/transaction/tx_generic"
	"github.com/incognitochain/incognito-chain/transaction/utils"
	"github.com/pkg/errors"
)

func (tx *TxToken) initEnv() metadata.ValidationEnviroment {
	valEnv := tx_generic.DefaultValEnv()
	// if txCustomTokenPrivacy.IsSalaryTx() {
	valEnv = tx_generic.WithAct(valEnv, common.TxActTranfer)
	// }
	if tx.IsPrivacy() {
		valEnv = tx_generic.WithPrivacy(valEnv)
	} else {
		valEnv = tx_generic.WithNoPrivacy(valEnv)
	}

	valEnv = tx_generic.WithType(valEnv, tx.GetType())
	sID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
	valEnv = tx_generic.WithShardID(valEnv, int(sID))
	valEnv = tx_generic.WithTokenID(valEnv, common.PRVCoinID)
	valEnv = tx_generic.WithSigPubkey(valEnv, tx.Tx.GetSigPubKey())
	tx.SetValidationEnv(valEnv)
	txNormalValEnv := valEnv.Clone()
	txTokenData := tx.GetTxTokenData()
	if txTokenData.Type == utils.CustomTokenInit {
		txNormalValEnv = tx_generic.WithAct(txNormalValEnv, common.TxActInit)
	} else {
		txNormalValEnv = tx_generic.WithAct(txNormalValEnv, common.TxActTranfer)
	}
	if txTokenData.TxNormal.IsPrivacy() {
		txNormalValEnv = tx_generic.WithPrivacy(txNormalValEnv)
	} else {
		txNormalValEnv = tx_generic.WithNoPrivacy(txNormalValEnv)
	}
	txNormalValEnv = tx_generic.WithTokenID(txNormalValEnv, txTokenData.PropertyID)
	txNormalValEnv = tx_generic.WithSigPubkey(txNormalValEnv, txTokenData.TxNormal.GetSigPubKey())
	tx.GetTxTokenData().TxNormal.SetValidationEnv(txNormalValEnv)
	return valEnv
}

func (tx *TxToken) ValidateSanityDataByItSelf() (bool, error) {
	isMint, _, _, _ := tx.GetTxMintData()
	bHeight := tx.GetValidationEnv().BeaconHeight()
	afterUpgrade := bHeight >= config.Param().BCHeightBreakPointPrivacyV2
	if afterUpgrade && !isMint {
		return false, utils.NewTransactionErr(utils.RejectTxVersion, errors.New("old version is no longer supported"))
	}
	if tx.GetType() != common.TxCustomTokenPrivacyType {
		return false, utils.NewTransactionErr(utils.InvalidSanityDataPrivacyTokenError, errors.New("txCustomTokenPrivacy.Tx should have type tp"))
	}
	if tx.GetTxNormal().GetType() != common.TxNormalType {
		return false, utils.NewTransactionErr(utils.InvalidSanityDataPrivacyTokenError, errors.New("txCustomTokenPrivacy.TxNormal should have type n"))
	}

	if tx.GetTxNormal().GetMetadata() != nil {
		return false, errors.Errorf("This tx field is just used for send token, can not have metadata")
	}
	if tx.GetTxTokenData().PropertyID.String() == common.PRVIDStr {
		return false, utils.NewTransactionErr(utils.InvalidSanityDataPrivacyTokenError, errors.New("TokenID must not be equal PRVID"))
	}

	if (tx.Tx.GetProof() != nil) && ((len(tx.Tx.GetProof().GetInputCoins()) != 0) || (len(tx.Tx.GetProof().GetOutputCoins()) != 0)) {
		ok, err := tx.Tx.ValidateSanityDataByItSelf()
		if !ok || err != nil {
			return ok, err
		}
	}
	ok, err := tx.ValidateSanityDataWithMetadata()
	if (!ok) || (err != nil) {
		return false, err
	}
	ok, err = tx.GetTxNormal().ValidateSanityDataByItSelf()
	if !ok || err != nil {
		return ok, err
	}

	return true, nil
}

func (tx *TxToken) ValidateSanityDataWithBlockchain(
	chainRetriever metadata.ChainRetriever,
	shardViewRetriever metadata.ShardViewRetriever,
	beaconViewRetriever metadata.BeaconViewRetriever,
	beaconHeight uint64,
) (
	bool,
	error,
) {
	if tx.GetMetadata() != nil {
		utils.Logger.Log.Debug("tx.Metadata.ValidateSanityData")
		isContinued, ok, err := tx.GetMetadata().ValidateSanityData(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight, tx)
		utils.Logger.Log.Debug("END tx.Metadata.ValidateSanityData")
		if err != nil || !ok || !isContinued {
			return ok, err
		}
	}
	return true, nil
}

func (tx *TxToken) LoadData(
	db *statedb.StateDB,
) error {
	tmpEmbededTx := tx.Tx
	tmpNormalTx := tx.GetTxNormal()

	embededTx, ok := tmpEmbededTx.(*Tx)
	if !ok {
		return fmt.Errorf("cannot case txBase to txVer1")
	}

	normalTx, ok := tmpNormalTx.(*Tx)
	if !ok {
		return fmt.Errorf("cannot case txNormal to txVer1")
	}

	tmpProof := embededTx.GetProof()

	if tmpProof != nil {
		prf, ok := tmpProof.(*privacy.ProofV1)
		if !ok {
			return fmt.Errorf("cannot cast payment proof v1")
		}
		if embededTx.GetValidationEnv().IsPrivacy() {
			txEnv := embededTx.GetValidationEnv()
			// tokenID := embededTx.GetTokenID()
			data, err := prf.LoadDataFromStateDB(db, &common.PRVCoinID, byte(tx.GetValidationEnv().ShardID()))
			if err != nil {
				utils.Logger.Log.Error(err)
				return err
			}
			embededTx.SetValidationEnv(tx_generic.WithDBData(txEnv, data))
		} else {
			for _, iCoin := range prf.GetInputCoins() {
				ok, err := embededTx.CheckCMExistence(
					iCoin.GetCommitment().ToBytesS(),
					db,
					byte(tx.GetValidationEnv().ShardID()),
					&common.PRVCoinID,
				)
				if !ok || err != nil {
					if err != nil {
						utils.Logger.Log.Error(err)
					}
					return utils.NewTransactionErr(utils.InputCommitmentIsNotExistedError, err)
				}
			}
		}
	}

	tokenID := tx.GetTokenID()
	if tx.GetTxTokenData().Type == utils.CustomTokenInit {
		if !tx.GetTxTokenData().Mintable {
			// check exist token
			if statedb.PrivacyTokenIDExisted(db, *tokenID) {
				return errors.Errorf("Privacy Token ID is existed")
			}
		}
	}

	tmpProof = normalTx.GetProof()

	if tmpProof != nil {
		prf, ok := tmpProof.(*privacy.ProofV1)
		if !ok {
			return fmt.Errorf("cannot cast payment proof v1")
		}
		if normalTx.GetValidationEnv().IsPrivacy() {
			txEnv := normalTx.GetValidationEnv()
			data, err := prf.LoadDataFromStateDB(db, tokenID, byte(tx.GetValidationEnv().ShardID()))
			if err != nil {
				utils.Logger.Log.Error(err)
				return err
			}
			normalTx.SetValidationEnv(tx_generic.WithDBData(txEnv, data))
		} else {
			for _, iCoin := range prf.GetInputCoins() {
				ok, err := normalTx.CheckCMExistence(
					iCoin.GetCommitment().ToBytesS(),
					db,
					byte(tx.GetValidationEnv().ShardID()),
					tokenID,
				)
				if !ok || err != nil {
					if err != nil {
						utils.Logger.Log.Error(err)
					}
					return utils.NewTransactionErr(utils.InputCommitmentIsNotExistedError, err)
				}
			}
		}
	}

	return nil
}

func (tx *TxToken) CheckData(
	db *statedb.StateDB,
) error {
	tmpEmbededTx := tx.Tx
	tmpNormalTx := tx.GetTxNormal()

	embededTx, ok := tmpEmbededTx.(*Tx)
	if !ok {
		return fmt.Errorf("cannot case txBase to txVer1")
	}

	normalTx, ok := tmpNormalTx.(*Tx)
	if !ok {
		return fmt.Errorf("cannot case txNormal to txVer1")
	}

	tmpProof := embededTx.GetProof()
	prf, ok := tmpProof.(*privacy.ProofV1)
	if !ok {
		return fmt.Errorf("cannot cast payment proof v1")
	}

	if prf != nil {
		if embededTx.GetValidationEnv().IsPrivacy() {
			txEnv := embededTx.GetValidationEnv()
			data := txEnv.DBData()
			err := prf.CheckCommitmentWithStateDB(data, db, &common.PRVCoinID, byte(tx.GetValidationEnv().ShardID()))
			if err != nil {
				utils.Logger.Log.Error(err)
				return err
			}
		} else {
			for _, iCoin := range prf.GetInputCoins() {
				ok, err := embededTx.CheckCMExistence(
					iCoin.GetCommitment().ToBytesS(),
					db,
					byte(tx.GetValidationEnv().ShardID()),
					&common.PRVCoinID,
				)
				if !ok || err != nil {
					if err != nil {
						utils.Logger.Log.Error(err)
					}
					return utils.NewTransactionErr(utils.InputCommitmentIsNotExistedError, err)
				}
			}
		}
	}

	tokenID := tx.GetTokenID()
	if tx.GetTxTokenData().Type == utils.CustomTokenInit {
		if !tx.GetTxTokenData().Mintable {
			// check exist token
			if statedb.PrivacyTokenIDExisted(db, *tokenID) {
				return errors.Errorf("Privacy Token ID is existed")
			}
		}
	}

	tmpProof = normalTx.GetProof()
	prf, ok = tmpProof.(*privacy.ProofV1)
	if !ok {
		return fmt.Errorf("cannot cast payment proof v1")
	}

	if prf != nil {
		if normalTx.GetValidationEnv().IsPrivacy() {
			txEnv := normalTx.GetValidationEnv()
			data := txEnv.DBData()
			err := prf.CheckCommitmentWithStateDB(data, db, tokenID, byte(tx.GetValidationEnv().ShardID()))
			if err != nil {
				utils.Logger.Log.Error(err)
				return err
			}
		} else {
			for _, iCoin := range prf.GetInputCoins() {
				ok, err := normalTx.CheckCMExistence(
					iCoin.GetCommitment().ToBytesS(),
					db,
					byte(tx.GetValidationEnv().ShardID()),
					tokenID,
				)
				if !ok || err != nil {
					if err != nil {
						utils.Logger.Log.Error(err)
					}
					return utils.NewTransactionErr(utils.InputCommitmentIsNotExistedError, err)
				}
			}
		}
	}

	return nil
}

func (tx *TxToken) ValidateTxCorrectness(
	transactionStateDB *statedb.StateDB,
) (
	bool,
	error,
) {
	utils.Logger.Log.Infof("VALIDATING TX %v; Env: Beacon %v, shard %v, confirmedTime %v", tx.Hash().String(), tx.GetValidationEnv().BeaconHeight(), tx.GetValidationEnv().ShardHeight(), tx.GetValidationEnv().ConfirmedTime())
	ok, err := tx.GetTxNormal().ValidateTxCorrectness(transactionStateDB)
	if (!ok) || (err != nil) {
		return ok, err
	}
	return tx.Tx.ValidateTxCorrectness(transactionStateDB)
}

// Todo decoupling this function
func (tx TxToken) validateSanityDataOfProofV2() (bool, error) {
	tmpProof := tx.GetProof()
	if tmpProof != nil {
		prf, ok := tmpProof.(*privacy.ProofV1)
		if !ok {
			return false, fmt.Errorf("cannot cast payment proof v1")
		}
		if len(tx.GetProof().GetInputCoins()) > 255 {
			return false, errors.New("Input coins in tx are very large:" + strconv.Itoa(len(prf.GetInputCoins())))
		}

		if len(prf.GetOutputCoins()) > 255 {
			return false, errors.New("Output coins in tx are very large:" + strconv.Itoa(len(prf.GetOutputCoins())))
		}

		// check doubling a input coin in tx
		serialNumbers := make(map[common.Hash]bool)
		for i, inCoin := range prf.GetInputCoins() {
			hashSN := common.HashH(inCoin.GetKeyImage().ToBytesS())
			if serialNumbers[hashSN] {
				utils.Logger.Log.Errorf("Double input in tx - txId %v - index %v", tx.Hash().String(), i)
				return false, errors.New("double input in tx")
			}
			serialNumbers[hashSN] = true
		}

		sndOutputs := make([]*privacy.Scalar, len(prf.GetOutputCoins()))
		for i, output := range prf.GetOutputCoins() {
			sndOutputs[i] = output.GetSNDerivator()
		}
		if privacy.CheckDuplicateScalarArray(sndOutputs) {
			utils.Logger.Log.Errorf("Duplicate output coins' snd")
			return false, utils.NewTransactionErr(utils.DuplicatedOutputSndError, errors.New("Duplicate output coins' snd"))
		}

		isPrivacy := tx.IsPrivacy()

		if isPrivacy {
			// check cmValue of output coins is equal to comValue in Bulletproof
			cmValueOfOutputCoins := prf.GetCommitmentOutputValue()
			cmValueInBulletProof := prf.GetAggregatedRangeProof().GetCommitments()
			if len(cmValueOfOutputCoins) != len(cmValueInBulletProof) {
				return false, errors.New("invalid cmValues in Bullet proof")
			}

			if len(prf.GetInputCoins()) != len(prf.GetSerialNumberProof()) || len(prf.GetInputCoins()) != len(prf.GetOneOfManyProof()) {
				return false, errors.New("the number of input coins must be equal to the number of serialnumber proofs and the number of one-of-many proofs")
			}

			for i := 0; i < len(cmValueOfOutputCoins); i++ {
				if !privacy.IsPointEqual(cmValueOfOutputCoins[i], cmValueInBulletProof[i]) {
					utils.Logger.Log.Errorf("cmValue in Bulletproof is not equal to commitment of output's Value - txId %v", tx.Hash().String())
					return false, fmt.Errorf("cmValue %v in Bulletproof is not equal to commitment of output's Value", i)
				}
			}

			if !prf.GetAggregatedRangeProof().ValidateSanity() {
				return false, errors.New("validate sanity Aggregated range proof failed")
			}

			for i := 0; i < len(prf.GetOneOfManyProof()); i++ {
				if !prf.GetOneOfManyProof()[i].ValidateSanity() {
					return false, errors.New("validate sanity One out of many proof failed")
				}
			}

			cmInputSNDs := prf.GetCommitmentInputSND()
			cmInputSK := prf.GetCommitmentInputSecretKey()
			for i := 0; i < len(prf.GetSerialNumberProof()); i++ {
				// check cmSK of input coin is equal to comSK in serial number proof
				if !privacy.IsPointEqual(cmInputSK, prf.GetSerialNumberProof()[i].GetComSK()) {
					utils.Logger.Log.Errorf("ComSK in SNproof is not equal to commitment of private key - txId %v", tx.Hash().String())
					return false, errhandler.NewPrivacyErr(errhandler.VerifySerialNumberPrivacyProofFailedErr, fmt.Errorf("comSK of SNProof %v is not comSK of input coins", i))
				}

				// check cmSND of input coins is equal to comInputSND in serial number proof
				if !privacy.IsPointEqual(cmInputSNDs[i], prf.GetSerialNumberProof()[i].GetComInput()) {
					utils.Logger.Log.Errorf("cmSND in SNproof is not equal to commitment of input's SND - txId %v", tx.Hash().String())
					return false, errhandler.NewPrivacyErr(errhandler.VerifySerialNumberPrivacyProofFailedErr, fmt.Errorf("cmSND in SNproof %v is not equal to commitment of input's SND", i))
				}

				// check SN of input coins is equal to the corresponding SN in serial number proof
				if !privacy.IsPointEqual(prf.GetInputCoins()[i].GetKeyImage(), prf.GetSerialNumberProof()[i].GetSN()) {
					utils.Logger.Log.Errorf("SN in SNProof is not equal to SN of input coin - txId %v", tx.Hash().String())
					return false, errhandler.NewPrivacyErr(errhandler.VerifySerialNumberPrivacyProofFailedErr, fmt.Errorf("SN in SNProof %v is not equal to SN of input coin", i))
				}

				if !prf.GetSerialNumberProof()[i].ValidateSanity() {
					return false, errors.New("validate sanity Serial number proof failed")
				}
			}

			// check input coins with privacy
			for i := 0; i < len(prf.GetInputCoins()); i++ {
				if !prf.GetInputCoins()[i].GetKeyImage().PointValid() {
					return false, errors.New("validate sanity Serial number of input coin failed")
				}
			}
			// check output coins with privacy
			for i := 0; i < len(prf.GetOutputCoins()); i++ {
				if !prf.GetOutputCoins()[i].GetPublicKey().PointValid() {
					return false, errors.New("validate sanity Public key of output coin failed")
				}
				if !prf.GetOutputCoins()[i].GetCommitment().PointValid() {
					return false, errors.New("validate sanity Coin commitment of output coin failed")
				}
				if !prf.GetOutputCoins()[i].GetSNDerivator().ScalarValid() {
					return false, errors.New("validate sanity SNDerivator of output coin failed")
				}
			}
			// check ComInputSK
			if !prf.GetCommitmentInputSecretKey().PointValid() {
				return false, errors.New("validate sanity ComInputSK of proof failed")
			}

			// check SigPubKey
			sigPubKeyPoint, err := new(privacy.Point).FromBytesS(tx.GetSigPubKey())
			if err != nil {
				utils.Logger.Log.Errorf("SigPubKey is invalid - txId %v", tx.Hash().String())
				return false, errors.New("SigPubKey is invalid")
			}
			if !privacy.IsPointEqual(cmInputSK, sigPubKeyPoint) {
				utils.Logger.Log.Errorf("SigPubKey is not equal to commitment of private key - txId %v", tx.Hash().String())
				return false, errors.New("SigPubKey is not equal to commitment of private key")
			}

			// check ComInputValue
			for i := 0; i < len(prf.GetCommitmentInputValue()); i++ {
				if !prf.GetCommitmentInputValue()[i].PointValid() {
					return false, errors.New("validate sanity ComInputValue of proof failed")
				}
			}
			//check ComInputSND
			for i := 0; i < len(prf.GetCommitmentInputSND()); i++ {
				if !prf.GetCommitmentInputSND()[i].PointValid() {
					return false, errors.New("validate sanity ComInputSND of proof failed")
				}
			}

			//check ComInputShardID
			if !prf.GetCommitmentInputShardID().PointValid() {
				return false, errors.New("validate sanity ComInputShardID of proof failed")
			}

			ok, err := prf.VerifySanityData(tx.GetValidationEnv())
			if err != nil {
				return false, err
			}
			if !ok {
				return false, nil
			}

			// check ComOutputShardID
			for i := 0; i < len(prf.GetCommitmentOutputShardID()); i++ {
				if !prf.GetCommitmentOutputShardID()[i].PointValid() {
					return false, errors.New("validate sanity ComOutputShardID of proof failed")
				}
			}
			//check ComOutputSND
			for i := 0; i < len(prf.GetCommitmentOutputShardID()); i++ {
				if !prf.GetCommitmentOutputShardID()[i].PointValid() {
					return false, errors.New("validate sanity ComOutputSND of proof failed")
				}
			}
			//check ComOutputValue
			for i := 0; i < len(prf.GetCommitmentOutputValue()); i++ {
				if !prf.GetCommitmentOutputValue()[i].PointValid() {
					return false, errors.New("validate sanity ComOutputValue of proof failed")
				}
			}
			if len(prf.GetCommitmentIndices()) != len(prf.GetInputCoins())*privacy.CommitmentRingSize {
				return false, errors.New("validate sanity CommitmentIndices of proof failed")

			}
		}

		if !isPrivacy {
			// check SigPubKey
			sigPubKeyPoint, err := new(privacy.Point).FromBytesS(tx.GetSigPubKey())
			if err != nil {
				utils.Logger.Log.Errorf("SigPubKey is invalid - txId %v", tx.Hash().String())
				return false, errors.New("SigPubKey is invalid")
			}
			inputCoins := prf.GetInputCoins()

			if len(inputCoins) != len(prf.GetSerialNumberNoPrivacyProof()) {
				return false, errors.New("the number of input coins must be equal to the number of serialnumbernoprivacy proofs")
			}

			for i := 0; i < len(inputCoins); i++ {
				// check PublicKey of input coin is equal to SigPubKey
				if !privacy.IsPointEqual(inputCoins[i].GetPublicKey(), sigPubKeyPoint) {
					utils.Logger.Log.Errorf("SigPubKey is not equal to public key of input coins - txId %v", tx.Hash().String())
					return false, errors.New("SigPubKey is not equal to public key of input coins")
				}
			}

			for i := 0; i < len(prf.GetSerialNumberNoPrivacyProof()); i++ {
				// check PK of input coin is equal to vKey in serial number proof
				if !privacy.IsPointEqual(prf.GetInputCoins()[i].GetPublicKey(), prf.GetSerialNumberNoPrivacyProof()[i].GetVKey()) {
					utils.Logger.Log.Errorf("VKey in SNNoPrivacyProof is not equal public key of sender - txId %v", tx.Hash().String())
					return false, errhandler.NewPrivacyErr(errhandler.VerifySerialNumberNoPrivacyProofFailedErr, fmt.Errorf("VKey of SNNoPrivacyProof %v is not public key of sender", i))
				}

				// check SND of input coins is equal to SND in serial number no privacy proof
				if !privacy.IsScalarEqual(prf.GetInputCoins()[i].GetSNDerivator(), prf.GetSerialNumberNoPrivacyProof()[i].GetInput()) {
					utils.Logger.Log.Errorf("SND in SNNoPrivacyProof is not equal to input's SND - txId %v", tx.Hash().String())
					return false, errhandler.NewPrivacyErr(errhandler.VerifySerialNumberNoPrivacyProofFailedErr, fmt.Errorf("SND in SNNoPrivacyProof %v is not equal to input's SND", i))
				}

				// check SND of input coins is equal to SND in serial number no privacy proof
				if !privacy.IsPointEqual(prf.GetInputCoins()[i].GetKeyImage(), prf.GetSerialNumberNoPrivacyProof()[i].GetOutput()) {
					utils.Logger.Log.Errorf("SN in SNNoPrivacyProof is not equal to SN in input coin - txId %v", tx.Hash().String())
					return false, errhandler.NewPrivacyErr(errhandler.VerifySerialNumberNoPrivacyProofFailedErr, fmt.Errorf("SN in SNNoPrivacyProof %v is not equal to SN in input coin", i))
				}

				if !prf.GetSerialNumberNoPrivacyProof()[i].ValidateSanity() {
					return false, errors.New("validate sanity Serial number no privacy proof failed")
				}
			}
			// check input coins without privacy
			for i := 0; i < len(prf.GetInputCoins()); i++ {
				if !prf.GetInputCoins()[i].GetCommitment().PointValid() {
					return false, errors.New("validate sanity CoinCommitment of input coin failed")
				}
				if !prf.GetInputCoins()[i].GetPublicKey().PointValid() {
					return false, errors.New("validate sanity PublicKey of input coin failed")
				}
				if !prf.GetInputCoins()[i].GetKeyImage().PointValid() {
					return false, errors.New("validate sanity Serial number of input coin failed")
				}
				if !prf.GetInputCoins()[i].GetRandomness().ScalarValid() {
					return false, errors.New("validate sanity Randomness of input coin failed")
				}
				if !prf.GetInputCoins()[i].GetSNDerivator().ScalarValid() {
					return false, errors.New("validate sanity SNDerivator of input coin failed")
				}
			}

			// check output coins without privacy
			for i := 0; i < len(prf.GetOutputCoins()); i++ {
				if !prf.GetOutputCoins()[i].GetCommitment().PointValid() {
					return false, errors.New("validate sanity CoinCommitment of output coin failed")
				}
				if !prf.GetOutputCoins()[i].GetPublicKey().PointValid() {
					return false, errors.New("validate sanity PublicKey of output coin failed")
				}
				if !prf.GetOutputCoins()[i].GetRandomness().ScalarValid() {
					return false, errors.New("validate sanity Randomness of output coin failed")
				}
				if !prf.GetOutputCoins()[i].GetSNDerivator().ScalarValid() {
					return false, errors.New("validate sanity SNDerivator of output coin failed")
				}
			}
		}
	}
	return true, nil
}
