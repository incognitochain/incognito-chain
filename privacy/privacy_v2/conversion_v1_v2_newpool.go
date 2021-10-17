//nolint:revive // skip linter for this package name
package privacy_v2

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/env"
	errhandler "github.com/incognitochain/incognito-chain/privacy/errorhandler"
	"github.com/incognitochain/incognito-chain/privacy/operation"
)

func (proof ConversionProofVer1ToVer2) VerifyV2(vEnv env.ValidationEnviroment, fee uint64) (bool, error) {
	// Step to verify ConversionProofVer1ToVer2
	//  - verify if inputCommitments existed
	//	- verify sumInput = sumOutput + fee
	//	- verify if serial number of each input coin has been derived correctly
	//	- verify input coins' randomness
	//	- verify if output coins' commitment has been calculated correctly
	var err error
	hasPrivacy := vEnv.IsPrivacy()
	if hasPrivacy {
		return false, errors.New("ConversionProof does not have privacy, something is wrong")
	}
	sumInput, sumOutput := uint64(0), uint64(0)
	pubKey := vEnv.SigPubKey()
	tokenID := vEnv.TokenID()
	for i := 0; i < len(proof.inputCoins); i++ {
		if proof.inputCoins[i].IsEncrypted() {
			return false, errors.New("ConversionProof input should not be encrypted")
		}
		if !bytes.Equal(pubKey, proof.inputCoins[i].GetPublicKey().ToBytesS()) {
			return false, errors.New("ConversionProof inputCoins.publicKey should equal to pubkey")
		}
		// Check the consistency of input coins' commitment
		commitment := proof.inputCoins[i].GetCommitment()
		err := proof.inputCoins[i].CommitAll()
		if err != nil {
			return false, err
		}
		if !bytes.Equal(commitment.ToBytesS(), proof.inputCoins[i].GetCommitment().ToBytesS()) {
			return false, errors.New("ConversionProof inputCoins.commitment is not correct")
		}

		// Check input overflow (not really necessary)
		inputValue := proof.inputCoins[i].GetValue()
		tmpInputSum := sumInput + inputValue
		if tmpInputSum < sumInput || tmpInputSum < inputValue {
			return false, errors.New("Overflown sumOutput")
		}
		sumInput += inputValue
	}
	for i := 0; i < len(proof.outputCoins); i++ {
		if proof.outputCoins[i].IsEncrypted() {
			return false, errors.New("ConversionProof output should not be encrypted")
		}
		// check output commitment
		outputValue := proof.outputCoins[i].GetValue()
		randomness := proof.outputCoins[i].GetRandomness()
		var tmpCommitment *operation.Point
		if tokenID.String() == common.PRVIDStr {
			tmpCommitment = operation.PedCom.CommitAtIndex(new(operation.Scalar).FromUint64(outputValue), randomness, operation.PedersenValueIndex)
		} else {
			tmpAssetTag := operation.HashToPoint(tokenID[:])
			if !bytes.Equal(tmpAssetTag.ToBytesS(), proof.outputCoins[i].GetAssetTag().ToBytesS()) {
				return false, fmt.Errorf("something is wrong with assetTag %v of tokenID %v: %v", proof.outputCoins[i].GetAssetTag().ToBytesS(), tokenID.String(), err)
			}
			tmpCommitment, err = proof.outputCoins[i].ComputeCommitmentCA()
			if err != nil {
				return false, fmt.Errorf("cannot compute output coin commitment for token %v: %v", tokenID.String(), err)
			}
		}
		if !bytes.Equal(tmpCommitment.ToBytesS(), proof.outputCoins[i].GetCommitment().ToBytesS()) {
			return false, fmt.Errorf("commitment of coin %v is not valid", i)
		}

		// Check output overflow
		tmpOutputSum := sumOutput + outputValue
		if tmpOutputSum < sumOutput || tmpOutputSum < outputValue {
			return false, errors.New("Overflown sumOutput")
		}
		sumOutput += outputValue
	}
	tmpSum := sumOutput + fee
	if tmpSum < sumOutput || tmpSum < fee {
		return false, fmt.Errorf("Overflown sumOutput+fee: output value = %v, fee = %v, sum = %v\n", sumOutput, fee, tmpSum)
	}
	if sumInput != tmpSum {
		return false, errors.New("ConversionProof input should be equal to fee + sum output")
	}
	if len(proof.inputCoins) != len(proof.serialNumberNoPrivacyProof) {
		return false, errors.New("The number of input coins should be equal to the number of proofs")
	}

	for i := 0; i < len(proof.inputCoins); i++ {
		valid, err := proof.serialNumberNoPrivacyProof[i].Verify(nil)
		if !valid {
			Logger.Log.Errorf("Verify serial number no privacy proof failed")
			return false, errhandler.NewPrivacyErr(errhandler.VerifySerialNumberNoPrivacyProofFailedErr, err)
		}
	}
	return true, nil
}
