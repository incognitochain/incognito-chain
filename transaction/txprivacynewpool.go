package transaction

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
)

func (tx Tx) ValidateSanityDataByItSelf() (bool, error) {
	switch tx.Type {
	case common.TxNormalType, common.TxRewardType, common.TxCustomTokenPrivacyType, common.TxReturnStakingType: //is valid
	default:
		return false, NewTransactionErr(RejectTxType, fmt.Errorf("wrong tx type with %s", tx.Type))
	}

	// check info field
	if len(tx.Info) > 512 {
		return false, NewTransactionErr(RejectTxInfoSize, fmt.Errorf("wrong tx info length %d bytes, only support info with max length <= %d bytes", len(tx.Info), 512))
	}

	if ((tx.Metadata.GetType() == metadata.ReturnStakingMeta) != (tx.valEnv.TxType() == common.TxReturnStakingType)) ||
		((tx.Metadata.GetType() == metadata.WithDrawRewardResponseMeta) != (tx.valEnv.TxType() == common.TxRewardType)) {
		return false, errors.Errorf("Not mismatch Type, txType: %v, metadataType %v", tx.valEnv.TxType(), tx.Metadata.GetType())
	}

	// check tx size
	actualTxSize := tx.GetTxActualSize()
	if actualTxSize > common.MaxTxSize {
		//fmt.Print(actualTxSize, common.MaxTxSize)
		return false, NewTransactionErr(RejectTxSize, fmt.Errorf("tx size %d kB is too large", actualTxSize))
	}

	//check version
	if tx.Version > txVersion {
		return false, NewTransactionErr(RejectTxVersion, fmt.Errorf("tx version is %d. Wrong version tx. Only support for version >= %d", tx.Version, txVersion))
	}
	// check LockTime before now
	if int64(tx.LockTime) > time.Now().Unix() {
		return false, NewTransactionErr(RejectInvalidLockTime, fmt.Errorf("wrong tx locktime %d", tx.LockTime))
	}

	// check sanity of Proof
	validateSanityOfProof, err := tx.validateSanityDataOfProofV2()
	if err != nil || !validateSanityOfProof {
		return false, err
	}

	if len(tx.SigPubKey) != common.SigPubKeySize {
		return false, NewTransactionErr(RejectTxPublickeySigSize, fmt.Errorf("wrong tx Sig PK size %d", len(tx.SigPubKey)))
	}

	return true, nil
}

func (tx *Tx) ValidateSanityDataWithBlockchain(
	chainRetriever metadata.ChainRetriever,
	shardViewRetriever metadata.ShardViewRetriever,
	beaconViewRetriever metadata.BeaconViewRetriever,
	beaconHeight uint64,
) (
	bool,
	error,
) {
	// Validate SND???
	// Validate DoubleSpend???
	if tx.Metadata != nil {
		Logger.log.Debug("tx.Metadata.ValidateSanityData")
		isContinued, ok, err := tx.Metadata.ValidateSanityData(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight, tx)
		Logger.log.Debug("END tx.Metadata.ValidateSanityData")
		if err != nil || !ok || !isContinued {
			return ok, err
		}
	}
	return true, nil
}

// func (tx *Tx) ValidateSanityDataWithBlockchain(

func (tx *Tx) ValidateTransactionV2(transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash) (bool, error) {
	//hasPrivacy = false
	Logger.log.Debugf("VALIDATING TX........\n")
	if tx.IsSalaryTx() {
		return tx.ValidateTxSalary(transactionStateDB)
	}
	// hasPrivacy := tx.IsPrivacy()
	var valid bool
	var err error

	if tx.GetType() == common.TxReturnStakingType {
		return true, nil //
	}

	if tx.Proof != nil {
		if tokenID == nil {
			tokenID = &common.Hash{}
			err := tokenID.SetBytes(common.PRVCoinID[:])
			if err != nil {
				Logger.log.Error(err)
				return false, NewTransactionErr(TokenIDInvalidError, err, tokenID.String())
			}
		}

		sndOutputs := make([]*privacy.Scalar, len(tx.Proof.GetOutputCoins()))
		for i := 0; i < len(tx.Proof.GetOutputCoins()); i++ {
			sndOutputs[i] = tx.Proof.GetOutputCoins()[i].CoinDetails.GetSNDerivator()
		}

		if privacy.CheckDuplicateScalarArray(sndOutputs) {
			Logger.log.Errorf("Duplicate output coins' snd\n")
			return false, NewTransactionErr(DuplicatedOutputSndError, errors.New("Duplicate output coins' snd\n"))
		}

		/*----------- TODO Moving out --------------

		// if isNewTransaction {
		// 	for i := 0; i < len(tx.Proof.GetOutputCoins()); i++ {
		// 		// Check output coins' SND is not exists in SND list (Database)
		// 		if ok, err := CheckSNDerivatorExistence(tokenID, tx.Proof.GetOutputCoins()[i].CoinDetails.GetSNDerivator(), transactionStateDB); ok 	|| err != nil {
		// 			if err != nil {
		// 				Logger.log.Error(err)
		// 			}
		// 			Logger.log.Errorf("snd existed: %d\n", i)
		// 			return false, NewTransactionErr(SndExistedError, err, fmt.Sprintf("snd existed: %d\n", i))
		// 		}
		// 	}
		// }

		// if !tx.valEnv.IsPrivacy() {
		// 	// Check input coins' commitment is exists in cm list (Database)
		// 	for i := 0; i < len(tx.Proof.GetInputCoins()); i++ {
		// 		ok, err := tx.CheckCMExistence(tx.Proof.GetInputCoins()[i].CoinDetails.GetCoinCommitment().ToBytesS(), transactionStateDB, shardID, 	tokenID)
		// 		if !ok || err != nil {
		// 			if err != nil {
		// 				Logger.log.Error(err)
		// 			}
		// 			return false, NewTransactionErr(InputCommitmentIsNotExistedError, err)
		// 		}
		// 	}
		// }
		------------------------------------------ */
		// Verify the payment proof

		valid, err = tx.Proof.VerifyV2(tx.valEnv, tx.SigPubKey, tx.Fee, transactionStateDB, shardID, tokenID)
		if !valid {
			if err != nil {
				Logger.log.Error(err)
			}
			Logger.log.Error("FAILED VERIFICATION PAYMENT PROOF")
			err1, ok := err.(*privacy.PrivacyError)
			if ok {
				// parse error detail
				if err1.Code == privacy.ErrCodeMessage[privacy.VerifyOneOutOfManyProofFailedErr].Code {
					// if isNewTransaction {
					// 	return false, NewTransactionErr(VerifyOneOutOfManyProofFailedErr, err1, tx.Hash().String())
					// } else {
					// for old txs which be get from sync block or validate new block
					if tx.LockTime <= ValidateTimeForOneoutOfManyProof {
						// only verify by sign on block because of issue #504(that mean we should pass old tx, which happen before this issue)
						return true, nil
					} else {
						return false, NewTransactionErr(VerifyOneOutOfManyProofFailedErr, err1, tx.Hash().String())
					}
					// }
				}
			}
			return false, NewTransactionErr(TxProofVerifyFailError, err, tx.Hash().String())
		} else {
			Logger.log.Debugf("SUCCESSED VERIFICATION PAYMENT PROOF ")
		}
	}
	//@UNCOMMENT: metrics time
	//elapsed := time.Since(start)
	//Logger.log.Debugf("Validation normal tx %+v in %s time \n", *tx.Hash(), elapsed)

	return true, nil
}

func (txN Tx) validateSanityDataOfProofV2() (bool, error) {
	if txN.Proof != nil {
		if len(txN.Proof.GetInputCoins()) > 255 {
			return false, errors.New("Input coins in tx are very large:" + strconv.Itoa(len(txN.Proof.GetInputCoins())))
		}

		if len(txN.Proof.GetOutputCoins()) > 255 {
			return false, errors.New("Output coins in tx are very large:" + strconv.Itoa(len(txN.Proof.GetOutputCoins())))
		}

		// check doubling a input coin in tx
		serialNumbers := make(map[common.Hash]bool)
		for i, inCoin := range txN.Proof.GetInputCoins() {
			hashSN := common.HashH(inCoin.CoinDetails.GetSerialNumber().ToBytesS())
			if serialNumbers[hashSN] {
				Logger.log.Errorf("Double input in tx - txId %v - index %v", txN.Hash().String(), i)
				return false, errors.New("double input in tx")
			}
			serialNumbers[hashSN] = true
		}

		isPrivacy := txN.IsPrivacy()

		if isPrivacy {
			// check cmValue of output coins is equal to comValue in Bulletproof
			cmValueOfOutputCoins := txN.Proof.GetCommitmentOutputValue()
			cmValueInBulletProof := txN.Proof.GetAggregatedRangeProof().GetCmValues()
			if len(cmValueOfOutputCoins) != len(cmValueInBulletProof) {
				return false, errors.New("invalid cmValues in Bullet proof")
			}

			if len(txN.Proof.GetInputCoins()) != len(txN.Proof.GetSerialNumberProof()) || len(txN.Proof.GetInputCoins()) != len(txN.Proof.GetOneOfManyProof()) {
				return false, errors.New("the number of input coins must be equal to the number of serialnumber proofs and the number of one-of-many proofs")
			}

			for i := 0; i < len(cmValueOfOutputCoins); i++ {
				if !privacy.IsPointEqual(cmValueOfOutputCoins[i], cmValueInBulletProof[i]) {
					Logger.log.Errorf("cmValue in Bulletproof is not equal to commitment of output's Value - txId %v", txN.Hash().String())
					return false, fmt.Errorf("cmValue %v in Bulletproof is not equal to commitment of output's Value", i)
				}
			}

			if !txN.Proof.GetAggregatedRangeProof().ValidateSanity() {
				return false, errors.New("validate sanity Aggregated range proof failed")
			}

			for i := 0; i < len(txN.Proof.GetOneOfManyProof()); i++ {
				if !txN.Proof.GetOneOfManyProof()[i].ValidateSanity() {
					return false, errors.New("validate sanity One out of many proof failed")
				}
			}

			cmInputSNDs := txN.Proof.GetCommitmentInputSND()
			cmInputSK := txN.Proof.GetCommitmentInputSecretKey()
			for i := 0; i < len(txN.Proof.GetSerialNumberProof()); i++ {
				// check cmSK of input coin is equal to comSK in serial number proof
				if !privacy.IsPointEqual(cmInputSK, txN.Proof.GetSerialNumberProof()[i].GetComSK()) {
					Logger.log.Errorf("ComSK in SNproof is not equal to commitment of private key - txId %v", txN.Hash().String())
					return false, privacy.NewPrivacyErr(privacy.VerifySerialNumberPrivacyProofFailedErr, fmt.Errorf("comSK of SNProof %v is not comSK of input coins", i))
				}

				// check cmSND of input coins is equal to comInputSND in serial number proof
				if !privacy.IsPointEqual(cmInputSNDs[i], txN.Proof.GetSerialNumberProof()[i].GetComInput()) {
					Logger.log.Errorf("cmSND in SNproof is not equal to commitment of input's SND - txId %v", txN.Hash().String())
					return false, privacy.NewPrivacyErr(privacy.VerifySerialNumberPrivacyProofFailedErr, fmt.Errorf("cmSND in SNproof %v is not equal to commitment of input's SND", i))
				}

				// check SN of input coins is equal to the corresponding SN in serial number proof
				if !privacy.IsPointEqual(txN.Proof.GetInputCoins()[i].CoinDetails.GetSerialNumber(), txN.Proof.GetSerialNumberProof()[i].GetSN()) {
					Logger.log.Errorf("SN in SNProof is not equal to SN of input coin - txId %v", txN.Hash().String())
					return false, privacy.NewPrivacyErr(privacy.VerifySerialNumberPrivacyProofFailedErr, fmt.Errorf("SN in SNProof %v is not equal to SN of input coin", i))
				}

				if !txN.Proof.GetSerialNumberProof()[i].ValidateSanity() {
					return false, errors.New("validate sanity Serial number proof failed")
				}
			}

			// check input coins with privacy
			for i := 0; i < len(txN.Proof.GetInputCoins()); i++ {
				if !txN.Proof.GetInputCoins()[i].CoinDetails.GetSerialNumber().PointValid() {
					return false, errors.New("validate sanity Serial number of input coin failed")
				}
			}
			// check output coins with privacy
			for i := 0; i < len(txN.Proof.GetOutputCoins()); i++ {
				if !txN.Proof.GetOutputCoins()[i].CoinDetails.GetPublicKey().PointValid() {
					return false, errors.New("validate sanity Public key of output coin failed")
				}
				if !txN.Proof.GetOutputCoins()[i].CoinDetails.GetCoinCommitment().PointValid() {
					return false, errors.New("validate sanity Coin commitment of output coin failed")
				}
				if !txN.Proof.GetOutputCoins()[i].CoinDetails.GetSNDerivator().ScalarValid() {
					return false, errors.New("validate sanity SNDerivator of output coin failed")
				}
			}
			// check ComInputSK
			if !txN.Proof.GetCommitmentInputSecretKey().PointValid() {
				return false, errors.New("validate sanity ComInputSK of proof failed")
			}

			// check SigPubKey
			sigPubKeyPoint, err := new(privacy.Point).FromBytesS(txN.GetSigPubKey())
			if err != nil {
				Logger.log.Errorf("SigPubKey is invalid - txId %v", txN.Hash().String())
				return false, errors.New("SigPubKey is invalid")
			}
			if !privacy.IsPointEqual(cmInputSK, sigPubKeyPoint) {
				Logger.log.Errorf("SigPubKey is not equal to commitment of private key - txId %v", txN.Hash().String())
				return false, errors.New("SigPubKey is not equal to commitment of private key")
			}

			// check ComInputValue
			for i := 0; i < len(txN.Proof.GetCommitmentInputValue()); i++ {
				if !txN.Proof.GetCommitmentInputValue()[i].PointValid() {
					return false, errors.New("validate sanity ComInputValue of proof failed")
				}
			}
			//check ComInputSND
			for i := 0; i < len(txN.Proof.GetCommitmentInputSND()); i++ {
				if !txN.Proof.GetCommitmentInputSND()[i].PointValid() {
					return false, errors.New("validate sanity ComInputSND of proof failed")
				}
			}

			//check ComInputShardID
			if !txN.Proof.GetCommitmentInputShardID().PointValid() {
				return false, errors.New("validate sanity ComInputShardID of proof failed")
			}

			ok, err := txN.Proof.VerifySanityData(txN.valEnv)
			if err != nil {
				return false, err
			}
			if !ok {
				return false, nil
			}

			// check ComOutputShardID
			for i := 0; i < len(txN.Proof.GetCommitmentOutputShardID()); i++ {
				if !txN.Proof.GetCommitmentOutputShardID()[i].PointValid() {
					return false, errors.New("validate sanity ComOutputShardID of proof failed")
				}
			}
			//check ComOutputSND
			for i := 0; i < len(txN.Proof.GetCommitmentOutputShardID()); i++ {
				if !txN.Proof.GetCommitmentOutputShardID()[i].PointValid() {
					return false, errors.New("validate sanity ComOutputSND of proof failed")
				}
			}
			//check ComOutputValue
			for i := 0; i < len(txN.Proof.GetCommitmentOutputValue()); i++ {
				if !txN.Proof.GetCommitmentOutputValue()[i].PointValid() {
					return false, errors.New("validate sanity ComOutputValue of proof failed")
				}
			}
			if len(txN.Proof.GetCommitmentIndices()) != len(txN.Proof.GetInputCoins())*privacy.CommitmentRingSize {
				return false, errors.New("validate sanity CommitmentIndices of proof failed")

			}
		}

		if !isPrivacy {
			// check SigPubKey
			sigPubKeyPoint, err := new(privacy.Point).FromBytesS(txN.GetSigPubKey())
			if err != nil {
				Logger.log.Errorf("SigPubKey is invalid - txId %v", txN.Hash().String())
				return false, errors.New("SigPubKey is invalid")
			}
			inputCoins := txN.Proof.GetInputCoins()

			if len(inputCoins) != len(txN.Proof.GetSerialNumberNoPrivacyProof()) {
				return false, errors.New("the number of input coins must be equal to the number of serialnumbernoprivacy proofs")
			}

			for i := 0; i < len(inputCoins); i++ {
				// check PublicKey of input coin is equal to SigPubKey
				if !privacy.IsPointEqual(inputCoins[i].CoinDetails.GetPublicKey(), sigPubKeyPoint) {
					Logger.log.Errorf("SigPubKey is not equal to public key of input coins - txId %v", txN.Hash().String())
					return false, errors.New("SigPubKey is not equal to public key of input coins")
				}
			}

			for i := 0; i < len(txN.Proof.GetSerialNumberNoPrivacyProof()); i++ {
				// check PK of input coin is equal to vKey in serial number proof
				if !privacy.IsPointEqual(txN.Proof.GetInputCoins()[i].CoinDetails.GetPublicKey(), txN.Proof.GetSerialNumberNoPrivacyProof()[i].GetVKey()) {
					Logger.log.Errorf("VKey in SNNoPrivacyProof is not equal public key of sender - txId %v", txN.Hash().String())
					return false, privacy.NewPrivacyErr(privacy.VerifySerialNumberNoPrivacyProofFailedErr, fmt.Errorf("VKey of SNNoPrivacyProof %v is not public key of sender", i))
				}

				// check SND of input coins is equal to SND in serial number no privacy proof
				if !privacy.IsScalarEqual(txN.Proof.GetInputCoins()[i].CoinDetails.GetSNDerivator(), txN.Proof.GetSerialNumberNoPrivacyProof()[i].GetInput()) {
					Logger.log.Errorf("SND in SNNoPrivacyProof is not equal to input's SND - txId %v", txN.Hash().String())
					return false, privacy.NewPrivacyErr(privacy.VerifySerialNumberNoPrivacyProofFailedErr, fmt.Errorf("SND in SNNoPrivacyProof %v is not equal to input's SND", i))
				}

				// check SND of input coins is equal to SND in serial number no privacy proof
				if !privacy.IsPointEqual(txN.Proof.GetInputCoins()[i].CoinDetails.GetSerialNumber(), txN.Proof.GetSerialNumberNoPrivacyProof()[i].GetOutput()) {
					Logger.log.Errorf("SN in SNNoPrivacyProof is not equal to SN in input coin - txId %v", txN.Hash().String())
					return false, privacy.NewPrivacyErr(privacy.VerifySerialNumberNoPrivacyProofFailedErr, fmt.Errorf("SN in SNNoPrivacyProof %v is not equal to SN in input coin", i))
				}

				if !txN.Proof.GetSerialNumberNoPrivacyProof()[i].ValidateSanity() {
					return false, errors.New("validate sanity Serial number no privacy proof failed")
				}
			}
			// check input coins without privacy
			for i := 0; i < len(txN.Proof.GetInputCoins()); i++ {
				if !txN.Proof.GetInputCoins()[i].CoinDetails.GetCoinCommitment().PointValid() {
					return false, errors.New("validate sanity CoinCommitment of input coin failed")
				}
				if !txN.Proof.GetInputCoins()[i].CoinDetails.GetPublicKey().PointValid() {
					return false, errors.New("validate sanity PublicKey of input coin failed")
				}
				if !txN.Proof.GetInputCoins()[i].CoinDetails.GetSerialNumber().PointValid() {
					return false, errors.New("validate sanity Serial number of input coin failed")
				}
				if !txN.Proof.GetInputCoins()[i].CoinDetails.GetRandomness().ScalarValid() {
					return false, errors.New("validate sanity Randomness of input coin failed")
				}
				if !txN.Proof.GetInputCoins()[i].CoinDetails.GetSNDerivator().ScalarValid() {
					return false, errors.New("validate sanity SNDerivator of input coin failed")
				}
			}

			// check output coins without privacy
			for i := 0; i < len(txN.Proof.GetOutputCoins()); i++ {
				if !txN.Proof.GetOutputCoins()[i].CoinDetails.GetCoinCommitment().PointValid() {
					return false, errors.New("validate sanity CoinCommitment of output coin failed")
				}
				if !txN.Proof.GetOutputCoins()[i].CoinDetails.GetPublicKey().PointValid() {
					return false, errors.New("validate sanity PublicKey of output coin failed")
				}
				if !txN.Proof.GetOutputCoins()[i].CoinDetails.GetRandomness().ScalarValid() {
					return false, errors.New("validate sanity Randomness of output coin failed")
				}
				if !txN.Proof.GetOutputCoins()[i].CoinDetails.GetSNDerivator().ScalarValid() {
					return false, errors.New("validate sanity SNDerivator of output coin failed")
				}
			}
		}
	}
	return true, nil
}
