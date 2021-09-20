package tx_ver2

import (
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2"
	"github.com/incognitochain/incognito-chain/transaction/tx_generic"
	"github.com/incognitochain/incognito-chain/transaction/utils"
)

func checkInputInDB(
	tx metadata.Transaction,
	db *statedb.StateDB,
) error {
	txEnv := tx.GetValidationEnv()
	tokenID := txEnv.TokenID()
	shardID := byte(txEnv.ShardID())
	proofConversion, ok := tx.GetProof().(*privacy_v2.ConversionProofVer1ToVer2)
	if !ok {
		utils.Logger.Log.Error("Error casting ConversionProofVer1ToVer2")
		return errors.New("Error casting ConversionProofVer1ToVer2")
	}
	inputCoins := proofConversion.GetInputCoins()
	for i := 0; i < len(inputCoins); i++ {
		// Check if commitment has existed
		if ok, err := statedb.HasCommitment(db, tokenID, inputCoins[i].GetCommitment().ToBytesS(), shardID); !ok || err != nil {
			if err != nil {
				errStr := fmt.Sprintf("TxConversion database inputCommitment got error: %v", err)
				return errors.New(errStr)
			}
			return errors.New("TxConversion not found existing inputCommitment in database error")
		}

		// Check if input coin has not been spent
		if ok, err := statedb.HasSerialNumber(db, tokenID, inputCoins[i].GetKeyImage().ToBytesS(), shardID); ok || err != nil {
			if err != nil {
				errStr := fmt.Sprintf("TxConversion database serialNumber got error: %v", err)
				return errors.New(errStr)
			}
			return errors.New("TxConversion found existing serialNumber in database error")
		}
	}
	return nil
}

func validateTxConvertCorrectness(
	tx metadata.Transaction,
	db *statedb.StateDB,
) (bool, error) {
	vEnv := tx.GetValidationEnv()
	utils.Logger.Log.Infof("Begin verifying TX %s", tx.Hash().String())
	if valid, err := tx_generic.VerifySigNoPrivacy(tx.GetSig(), tx.GetSigPubKey(), tx.Hash()[:]); !valid {
		if err != nil {
			utils.Logger.Log.Errorf("Error verifying signature conversion with tx hash %s: %+v \n", tx.Hash().String(), err)
			return false, utils.NewTransactionErr(utils.VerifyTxSigFailError, err)
		}
		utils.Logger.Log.Errorf("FAILED VERIFICATION SIGNATURE conversion with tx hash %s", tx.Hash().String())
		return false, utils.NewTransactionErr(utils.VerifyTxSigFailError, fmt.Errorf("FAILED VERIFICATION SIGNATURE ver1 with tx hash %s", tx.Hash().String()))
	}
	proofConversion, ok := tx.GetProof().(*privacy_v2.ConversionProofVer1ToVer2)
	if !ok {
		utils.Logger.Log.Error("Error casting ConversionProofVer1ToVer2")
		return false, errors.New("Error casting ConversionProofVer1ToVer2")
	}

	// Verify that input coins have been created and have not been spent
	// inputCoins := proofConversion.GetInputCoins()
	// for i := 0; i < len(inputCoins); i++ {
	// 	// Check if commitment has existed
	// 	if ok, err := statedb.HasCommitment(db, tokenID, inputCoins[i].GetCommitment().ToBytesS(), shardID); !ok || err != nil {
	// 		if err != nil {
	// 			errStr := fmt.Sprintf("TxConversion database inputCommitment got error: %v", err)
	// 			return false, errors.New(errStr)
	// 		}
	// 		return false, errors.New("TxConversion not found existing inputCommitment in database error")
	// 	}

	// 	// Check if input coin has not been spent
	// 	if ok, err := statedb.HasSerialNumber(db, tokenID, inputCoins[i].GetKeyImage().ToBytesS(), shardID); ok || err != nil {
	// 		if err != nil {
	// 			errStr := fmt.Sprintf("TxConversion database serialNumber got error: %v", err)
	// 			return false, errors.New(errStr)
	// 		}
	// 		return false, errors.New("TxConversion found existing serialNumber in database error")
	// 	}
	// }

	// Verify that output coins' one-time-address has not been obtained + not duplicate OTAs
	// outputCoins := proofConversion.GetOutputCoins()
	// mapOutputCoins := make(map[string]int)
	// for i := 0; i < len(outputCoins); i++ {
	// 	if ok, _, err := statedb.HasOnetimeAddress(db, *tokenID, outputCoins[i].GetPublicKey().ToBytesS()); ok || err != nil {
	// 		if err != nil {
	// 			errStr := fmt.Sprintf("TxConversion database onetimeAddress got error: %v", err)
	// 			return false, errors.New(errStr)
	// 		}
	// 		return false, errors.New("TxConversion found existing one-time-address in database error")
	// 	}
	// 	dst := make([]byte, hex.EncodedLen(len(outputCoins[i].GetPublicKey().ToBytesS())))
	// 	hex.Encode(dst, outputCoins[i].GetPublicKey().ToBytesS())
	// 	mapOutputCoins[string(dst)] = i
	// }
	// if len(mapOutputCoins) != len(outputCoins) {
	// 	return false, errors.New("TxConversion found duplicate one-time-address error")
	// }

	// Verify the conversion proof
	// valid, err := proofConversion.Verify(boolParams, tx.GetSigPubKey(), tx.GetTxFee(), shardID, tokenID, nil)
	valid, err := proofConversion.VerifyV2(vEnv, tx.GetTxFee())
	if !valid {
		if err != nil {
			utils.Logger.Log.Error(err)
		}
		return false, utils.NewTransactionErr(utils.TxProofVerifyFailError, err, tx.Hash().String())
	}

	utils.Logger.Log.Debugf("SUCCEEDED VERIFICATION PAYMENT PROOF")
	return true, nil
}
