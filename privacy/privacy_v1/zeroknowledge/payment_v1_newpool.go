package zkp

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy/env"
	errhandler "github.com/incognitochain/incognito-chain/privacy/errorhandler"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_util"
	"github.com/pkg/errors"
)

func (proof *PaymentProof) LoadCommitmentFromStateDB(db *statedb.StateDB, tokenID *common.Hash, shardID byte) error {
	Logger.Log.Infof("[testperformance] LoadCommitmentFromStateDB, tokenID %v, shardID %v", tokenID.String(), shardID)
	cmInputSum := make([]*operation.Point, len(proof.oneOfManyProof))
	for i := 0; i < len(proof.oneOfManyProof); i++ {
		// Calculate cm input sum
		cmInputSum[i] = new(operation.Point).Add(proof.commitmentInputSecretKey, proof.commitmentInputValue[i])
		cmInputSum[i].Add(cmInputSum[i], proof.commitmentInputSND[i])
		cmInputSum[i].Add(cmInputSum[i], proof.commitmentInputShardID)

		// get commitments list from CommitmentIndices
		commitments := make([]*operation.Point, privacy_util.CommitmentRingSize)
		for j := 0; j < privacy_util.CommitmentRingSize; j++ {
			index := proof.commitmentIndices[i*privacy_util.CommitmentRingSize+j]
			commitmentBytes, err := statedb.GetCommitmentByIndex(db, *tokenID, index, shardID)
			Logger.Log.Debugf("[TEST] commitment at index %v: %v\n", index, commitmentBytes)
			if err != nil {
				Logger.Log.Errorf("VERIFICATION PAYMENT PROOF 1: Error when get commitment by index from database", index, err)
				return errhandler.NewPrivacyErr(errhandler.VerifyOneOutOfManyProofFailedErr, err)
			}
			recheckIndex, err := statedb.GetCommitmentIndex(db, *tokenID, commitmentBytes, shardID)
			if err != nil || recheckIndex.Uint64() != index {
				Logger.Log.Errorf("VERIFICATION PAYMENT PROOF 2: Error when get commitment by index from database", index, err)
				return errhandler.NewPrivacyErr(errhandler.VerifyOneOutOfManyProofFailedErr, err)
			}
			commitments[j], err = new(operation.Point).FromBytesS(commitmentBytes)
			if err != nil {
				Logger.Log.Errorf("VERIFICATION PAYMENT PROOF: Cannot decompress commitment from database", index, err)
				return errhandler.NewPrivacyErr(errhandler.VerifyOneOutOfManyProofFailedErr, err)
			}
			commitments[j].Sub(commitments[j], cmInputSum[i])
		}

		proof.oneOfManyProof[i].Statement.Commitments = commitments

	}
	return nil
}

// Validate all of conditions
func (proof PaymentProof) VerifySanityData(
	vEnv env.ValidationEnviroment,
) (
	bool,
	error,
) {
	senderSID := vEnv.ShardID()
	if IsNewZKP(vEnv.BeaconHeight()) {
		expectedCMShardID := operation.PedCom.CommitAtIndex(
			new(operation.Scalar).FromUint64(uint64(senderSID)),
			fixedRandomnessShardID,
			operation.PedersenShardIDIndex,
		)
		if !operation.IsPointEqual(expectedCMShardID, proof.GetCommitmentInputShardID()) {
			return false, errors.New("ComInputShardID must be committed with the fixed randomness")
		}
	}

	return false, nil
}

func (proof PaymentProof) VerifyV2(
	vEnv env.ValidationEnviroment,
	pubKey key.PublicKey,
	fee uint64,
) (
	bool,
	error,
) {
	if !vEnv.IsPrivacy() {
		return proof.verifyNoPrivacyV2(pubKey, fee, vEnv)
	}
	return proof.verifyHasPrivacyV2(pubKey, fee, vEnv)
}

func (proof PaymentProof) verifyNoPrivacyV2(
	pubKey key.PublicKey,
	fee uint64,
// shardID byte,
// tokenID *common.Hash,
	vEnv env.ValidationEnviroment,
) (
	bool,
	error,
) {
	var sumInputValue, sumOutputValue uint64
	sumInputValue = 0
	sumOutputValue = 0

	pubKeyLastByteSender := pubKey[len(pubKey)-1]
	senderShardID := common.GetShardIDFromLastByte(pubKeyLastByteSender)
	cmShardIDSender := new(operation.Point)
	cmShardIDSender.ScalarMult(operation.PedCom.G[operation.PedersenShardIDIndex], new(operation.Scalar).FromBytes([operation.Ed25519KeySize]byte{senderShardID}))

	isNewZKP := IsNewZKP(vEnv.BeaconHeight())

	for i := 0; i < len(proof.inputCoins); i++ {
		if isNewZKP {
			// Check input coins' Serial number is created from input coins' input and sender's spending key
			valid, err := proof.serialNumberNoPrivacyProof[i].Verify(nil)
			if !valid {
				Logger.Log.Errorf("Verify serial number no privacy proof failed")
				return false, errhandler.NewPrivacyErr(errhandler.VerifySerialNumberNoPrivacyProofFailedErr, err)
			}
		} else {
			// Check input coins' Serial number is created from input coins' input and sender's spending key
			valid, err := proof.serialNumberNoPrivacyProof[i].VerifyOld(nil)
			if !valid {
				valid2, err2 := proof.serialNumberNoPrivacyProof[i].Verify(nil)
				if !valid2 {
					err3 := errors.Errorf("Verify Old and New serial number no privacy proof failed, error %v %v", err, err2)
					Logger.Log.Error(err3)
					return false, errhandler.NewPrivacyErr(errhandler.VerifySerialNumberNoPrivacyProofFailedErr, err3)
				}
			}
		}

		// Check input coins' cm is calculated correctly
		cmSK := proof.inputCoins[i].GetPublicKey()
		cmValue := new(operation.Point).ScalarMult(operation.PedCom.G[operation.PedersenValueIndex], new(operation.Scalar).FromUint64(proof.inputCoins[i].GetValue()))
		cmSND := new(operation.Point).ScalarMult(operation.PedCom.G[operation.PedersenSndIndex], proof.inputCoins[i].GetSNDerivator())
		cmRandomness := new(operation.Point).ScalarMult(operation.PedCom.G[operation.PedersenRandomnessIndex], proof.inputCoins[i].GetRandomness())
		cmTmp := new(operation.Point).Add(cmSK, cmValue)
		cmTmp.Add(cmTmp, cmSND)
		cmTmp.Add(cmTmp, cmShardIDSender)
		cmTmp.Add(cmTmp, cmRandomness)

		if !operation.IsPointEqual(cmTmp, proof.inputCoins[i].GetCommitment()) {
			Logger.Log.Errorf("Input coins %v commitment wrong!\n", i)
			return false, errhandler.NewPrivacyErr(errhandler.VerifyCoinCommitmentInputFailedErr, nil)
		}

		// Calculate sum of input values
		sumInputValue += proof.inputCoins[i].GetValue()
	}

	for i := 0; i < len(proof.outputCoins); i++ {
		// Check output coins' cm is calculated correctly
		shardID, err := proof.outputCoins[i].GetShardID()
		if err != nil {
			Logger.Log.Errorf("Cannot parse shardID of outputcoin error: %v", err)
			return false, err
		}
		cmSK := proof.outputCoins[i].CoinDetails.GetPublicKey()
		cmValue := new(operation.Point).ScalarMult(operation.PedCom.G[operation.PedersenValueIndex], new(operation.Scalar).FromUint64(proof.outputCoins[i].CoinDetails.GetValue()))
		cmSND := new(operation.Point).ScalarMult(operation.PedCom.G[operation.PedersenSndIndex], proof.outputCoins[i].CoinDetails.GetSNDerivator())
		cmShardID := new(operation.Point).ScalarMult(operation.PedCom.G[operation.PedersenShardIDIndex], new(operation.Scalar).FromBytes([operation.Ed25519KeySize]byte{shardID}))
		cmRandomness := new(operation.Point).ScalarMult(operation.PedCom.G[operation.PedersenRandomnessIndex], proof.outputCoins[i].CoinDetails.GetRandomness())

		cmTmp := new(operation.Point).Add(cmSK, cmValue)
		cmTmp.Add(cmTmp, cmSND)
		cmTmp.Add(cmTmp, cmShardID)
		cmTmp.Add(cmTmp, cmRandomness)

		if !operation.IsPointEqual(cmTmp, proof.outputCoins[i].GetCommitment()) {
			Logger.Log.Errorf("Output coins %v commitment wrong!\n", i)
			return false, errhandler.NewPrivacyErr(errhandler.VerifyCoinCommitmentOutputFailedErr, nil)
		}
	}

	//Calculate sum of output values and check overflow output's value
	if len(proof.outputCoins) > 0 {
		sumOutputValue = proof.outputCoins[0].CoinDetails.GetValue()
		for i := 1; i < len(proof.outputCoins); i++ {
			outValue := proof.outputCoins[i].CoinDetails.GetValue()
			sumTmp := sumOutputValue + outValue
			if sumTmp < sumOutputValue || sumTmp < outValue {
				return false, errhandler.NewPrivacyErr(errhandler.UnexpectedErr, fmt.Errorf("Overflow output value %v\n", outValue))
			}
			sumOutputValue += outValue
		}
	}

	// check overflow fee value
	tmp := sumOutputValue + fee
	if tmp < sumOutputValue || tmp < fee {
		return false, errhandler.NewPrivacyErr(errhandler.UnexpectedErr, fmt.Errorf("Overflow fee value %v\n", fee))
	}
	if (vEnv.TxType() == common.TxRewardType) || (vEnv.TxType() == common.TxReturnStakingType) || (vEnv.TxAction() == common.TxActInit) {
		return true, nil
	}
	// check if sum of input values equal sum of output values
	if sumInputValue != sumOutputValue+fee {
		Logger.Log.Debugf("sumInputValue: %v\n", sumInputValue)
		Logger.Log.Debugf("sumOutputValue: %v\n", sumOutputValue)
		Logger.Log.Debugf("fee: %v\n", fee)
		Logger.Log.Errorf("Sum of inputs is not equal sum of output!\n")
		return false, errhandler.NewPrivacyErr(errhandler.VerifyAmountNoPrivacyFailedErr, nil)
	}
	return true, nil
}

func (proof PaymentProof) verifyHasPrivacyV2(
	pubKey key.PublicKey,
	fee uint64,
	vEnv env.ValidationEnviroment,
) (
	bool,
	error,
) {

	isNewZKP := IsNewZKP(vEnv.BeaconHeight())
	shardID := byte(vEnv.ShardID())
	for i := 0; i < len(proof.oneOfManyProof); i++ {
		Logger.Log.Debugf("[TEST] input coins %v\n ShardID %v fee %v", i, shardID, fee)
		Logger.Log.Debugf("[TEST] commitments indices %v\n", proof.commitmentIndices[i*privacy_util.CommitmentRingSize:i*privacy_util.CommitmentRingSize+8])
		// Verify for the proof one-out-of-N commitments is a commitment to the coins being spent

		if isNewZKP {
			if IsNewOneOfManyProof(vEnv.ConfimedTime()) {
				valid, err := proof.oneOfManyProof[i].Verify()
				if !valid {
					Logger.Log.Errorf("VERIFICATION PAYMENT PROOF: One out of many failed")
					return false, errhandler.NewPrivacyErr(errhandler.VerifyOneOutOfManyProofFailedErr, err)
				}
			}
			// Verify for the Proof that input coins' serial number is derived from the committed derivator
			valid, err := proof.serialNumberProof[i].Verify(nil)
			if !valid {
				Logger.Log.Errorf("VERIFICATION PAYMENT PROOF: Serial number privacy failed")
				return false, errhandler.NewPrivacyErr(errhandler.VerifySerialNumberPrivacyProofFailedErr, err)
			}
		} else {
			if IsNewOneOfManyProof(vEnv.ConfimedTime()) {
				valid, err := proof.oneOfManyProof[i].VerifyOld()
				if !valid {
					valid2, err2 := proof.oneOfManyProof[i].Verify()
					if !valid2 {
						err3 := errors.Errorf("VERIFICATION PAYMENT PROOF Old and New: One out of many failed, error %v %v", err, err2)
						Logger.Log.Error(err3)
						return false, errhandler.NewPrivacyErr(errhandler.VerifyOneOutOfManyProofFailedErr, err3)
					}
				}
			}
			// Verify for the Proof that input coins' serial number is derived from the committed derivator
			valid, err := proof.serialNumberProof[i].VerifyOld(nil)
			if !valid {
				valid2, err2 := proof.serialNumberProof[i].Verify(nil)
				if !valid2 {
					err3 := errors.Errorf("Verify Old and New serial number no privacy proof failed, error %v %v", err, err2)
					Logger.Log.Error(err3)
					return false, errhandler.NewPrivacyErr(errhandler.VerifySerialNumberNoPrivacyProofFailedErr, err3)
				}
			}
		}
	}

	// Check output coins' cm is calculated correctly
	for i := 0; i < len(proof.outputCoins); i++ {
		cmTmp := new(operation.Point).Add(proof.outputCoins[i].CoinDetails.GetPublicKey(), proof.commitmentOutputValue[i])
		cmTmp.Add(cmTmp, proof.commitmentOutputSND[i])
		cmTmp.Add(cmTmp, proof.commitmentOutputShardID[i])

		if !operation.IsPointEqual(cmTmp, proof.outputCoins[i].GetCommitment()) {
			Logger.Log.Errorf("VERIFICATION PAYMENT PROOF: Commitment for output coins are not computed correctly")
			return false, errhandler.NewPrivacyErr(errhandler.VerifyCoinCommitmentOutputFailedErr, nil)
		}
	}

	if isNewZKP {
		valid, err := proof.aggregatedRangeProof.Verify()
		if !valid {
			Logger.Log.Errorf("VERIFICATION PAYMENT PROOF: Multi-range failed")
			return false, errhandler.NewPrivacyErr(errhandler.VerifyAggregatedProofFailedErr, err)
		}
	} else {
		valid, err := proof.aggregatedRangeProof.VerifyOld()
		if !valid {
			valid2, err2 := proof.aggregatedRangeProof.Verify()
			if !valid2 {
				err3 := errors.Errorf("VERIFICATION PAYMENT PROOF Old and New: Multi-range failed, error %v %v", err, err2)
				Logger.Log.Error(err3)
				return false, errhandler.NewPrivacyErr(errhandler.VerifyAggregatedProofFailedErr, err3)
			}
		}
	}

	// Verify the proof that sum of all input values is equal to sum of all output values
	comInputValueSum := new(operation.Point).Identity()
	for i := 0; i < len(proof.commitmentInputValue); i++ {
		comInputValueSum.Add(comInputValueSum, proof.commitmentInputValue[i])
	}

	comOutputValueSum := new(operation.Point).Identity()
	for i := 0; i < len(proof.commitmentOutputValue); i++ {
		comOutputValueSum.Add(comOutputValueSum, proof.commitmentOutputValue[i])
	}

	if fee > 0 {
		comOutputValueSum.Add(comOutputValueSum, new(operation.Point).ScalarMult(operation.PedCom.G[operation.PedersenValueIndex], new(operation.Scalar).FromUint64(uint64(fee))))
	}

	Logger.Log.Infof("comInputValueSum: %v\n", comInputValueSum.ToBytesS())
	Logger.Log.Infof("comOutputValueSum: %v\n", comOutputValueSum.ToBytesS())

	if !operation.IsPointEqual(comInputValueSum, comOutputValueSum) {
		Logger.Log.Debugf("comInputValueSum: ", comInputValueSum)
		Logger.Log.Debugf("comOutputValueSum: ", comOutputValueSum)
		Logger.Log.Error("VERIFICATION PAYMENT PROOF: Sum of input coins' value is not equal to sum of output coins' value")
		return false, errhandler.NewPrivacyErr(errhandler.VerifyAmountPrivacyFailedErr, nil)
	}

	return true, nil
}

