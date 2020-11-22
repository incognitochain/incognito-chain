package transaction

import (
	"errors"
	"fmt"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
)

func (tx Tx) validateNormalTxSanityDatav2(bcr metadata.ChainRetriever, beaconHeight uint64) (bool, error) {
	//check version
	if tx.Version > txVersion {
		return false, NewTransactionErr(RejectTxVersion, fmt.Errorf("tx version is %d. Wrong version tx. Only support for version >= %d", tx.Version, txVersion))
	}
	// check LockTime before now
	if int64(tx.LockTime) > time.Now().Unix() {
		return false, NewTransactionErr(RejectInvalidLockTime, fmt.Errorf("wrong tx locktime %d", tx.LockTime))
	}

	// check tx size
	actualTxSize := tx.GetTxActualSize()
	if actualTxSize > common.MaxTxSize {
		//fmt.Print(actualTxSize, common.MaxTxSize)
		return false, NewTransactionErr(RejectTxSize, fmt.Errorf("tx size %d kB is too large", actualTxSize))
	}

	// check sanity of Proof
	validateSanityOfProof, err := tx.validateSanityDataOfProof(bcr, beaconHeight)
	if err != nil || !validateSanityOfProof {
		return false, err
	}

	if len(tx.SigPubKey) != common.SigPubKeySize {
		return false, NewTransactionErr(RejectTxPublickeySigSize, fmt.Errorf("wrong tx Sig PK size %d", len(tx.SigPubKey)))
	}
	// check Type is normal or salary tx
	switch tx.Type {
	case common.TxNormalType, common.TxRewardType, common.TxCustomTokenPrivacyType, common.TxReturnStakingType: //is valid
	default:
		return false, NewTransactionErr(RejectTxType, fmt.Errorf("wrong tx type with %s", tx.Type))
	}

	//if txN.Type != common.TxNormalType && txN.Type != common.TxRewardType && txN.Type != common.TxCustomTokenType && txN.Type != common.TxCustomTokenPrivacyType { // only 1 byte
	//	return false, errors.New("wrong tx type")
	//}

	// check info field
	if len(tx.Info) > 512 {
		return false, NewTransactionErr(RejectTxInfoSize, fmt.Errorf("wrong tx info length %d bytes, only support info with max length <= %d bytes", len(tx.Info), 512))
	}

	return true, nil
}

func (tx Tx) validateNormalTxSanityDataWithBlkChain(bcr metadata.ChainRetriever, beaconHeight uint64) (bool, error) {
	//check version
	if tx.Version > txVersion {
		return false, NewTransactionErr(RejectTxVersion, fmt.Errorf("tx version is %d. Wrong version tx. Only support for version >= %d", tx.Version, txVersion))
	}
	// check LockTime before now
	if int64(tx.LockTime) > time.Now().Unix() {
		return false, NewTransactionErr(RejectInvalidLockTime, fmt.Errorf("wrong tx locktime %d", tx.LockTime))
	}

	// check tx size
	actualTxSize := tx.GetTxActualSize()
	if actualTxSize > common.MaxTxSize {
		//fmt.Print(actualTxSize, common.MaxTxSize)
		return false, NewTransactionErr(RejectTxSize, fmt.Errorf("tx size %d kB is too large", actualTxSize))
	}

	// check sanity of Proof
	validateSanityOfProof, err := tx.validateSanityDataOfProof(bcr, beaconHeight)
	if err != nil || !validateSanityOfProof {
		return false, err
	}

	if len(tx.SigPubKey) != common.SigPubKeySize {
		return false, NewTransactionErr(RejectTxPublickeySigSize, fmt.Errorf("wrong tx Sig PK size %d", len(tx.SigPubKey)))
	}
	// check Type is normal or salary tx
	switch tx.Type {
	case common.TxNormalType, common.TxRewardType, common.TxCustomTokenPrivacyType, common.TxReturnStakingType: //is valid
	default:
		return false, NewTransactionErr(RejectTxType, fmt.Errorf("wrong tx type with %s", tx.Type))
	}

	//if txN.Type != common.TxNormalType && txN.Type != common.TxRewardType && txN.Type != common.TxCustomTokenType && txN.Type != common.TxCustomTokenPrivacyType { // only 1 byte
	//	return false, errors.New("wrong tx type")
	//}

	// check info field
	if len(tx.Info) > 512 {
		return false, NewTransactionErr(RejectTxInfoSize, fmt.Errorf("wrong tx info length %d bytes, only support info with max length <= %d bytes", len(tx.Info), 512))
	}

	return true, nil
}

func (tx *Tx) ValidateSanityDataV2() (bool, error) {
	return false, nil
}

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
