package privacy_v2

import (
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	errhandler "github.com/incognitochain/incognito-chain/privacy/errorhandler"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2/bulletproofs"
	"github.com/incognitochain/incognito-chain/privacy/proof/agg_interface"
)

type PaymentProofV2 struct {
	Version              uint8
	aggregatedRangeProof *bulletproofs.AggregatedRangeProof
	inputCoins           []*coin.InputCoin
	outputCoins          []*coin.OutputCoin
}

func (proof PaymentProofV2) GetVersion() uint8 { return 2 }
func (proof PaymentProofV2) GetInputCoins() []*coin.InputCoin { return proof.inputCoins }
func (proof PaymentProofV2) GetOutputCoins() []*coin.OutputCoin { return proof.outputCoins }

func (proof *PaymentProofV2) SetVersion()      { proof.Version = 2 }
func (proof *PaymentProofV2) SetInputCoins(v []*coin.InputCoin) { proof.inputCoins = v }
func (proof *PaymentProofV2) SetOutputCoins(v []*coin.OutputCoin) { proof.outputCoins = v }

func (proof PaymentProofV2) GetAggregatedRangeProof() *agg_interface.AggregatedRangeProof {
	var a agg_interface.AggregatedRangeProof = proof.aggregatedRangeProof
	return &a
}

func (proof *PaymentProofV2) Bytes() []byte {
	var bytes []byte
	bytes = append(bytes, proof.GetVersion())

	comOutputMultiRangeProof := proof.aggregatedRangeProof.Bytes()
	var rangeProofLength uint32 = uint32(len(comOutputMultiRangeProof))
	bytes = append(bytes, common.Uint32ToBytes(rangeProofLength)...)
	bytes = append(bytes, comOutputMultiRangeProof...)

	// InputCoins
	bytes = append(bytes, byte(len(proof.inputCoins)))
	for i := 0; i < len(proof.inputCoins); i++ {
		inputCoins := proof.inputCoins[i].Bytes()
		bytes = append(bytes, byte(len(inputCoins)))
		bytes = append(bytes, inputCoins...)
	}

	// OutputCoins
	bytes = append(bytes, byte(len(proof.outputCoins)))
	for i := 0; i < len(proof.outputCoins); i++ {
		outputCoins := proof.outputCoins[i].Bytes()
		lenOutputCoins := len(outputCoins)
		lenOutputCoinsBytes := []byte{}
		if lenOutputCoins < 256 {
			lenOutputCoinsBytes = []byte{byte(lenOutputCoins)}
		} else {
			lenOutputCoinsBytes = common.IntToBytes(lenOutputCoins)
		}

		bytes = append(bytes, lenOutputCoinsBytes...)
		bytes = append(bytes, outputCoins...)
	}

	return bytes
}

func (proof *PaymentProofV2) SetBytes(proofbytes []byte) *errhandler.PrivacyError {
	if len(proofbytes) == 0 {
		return errhandler.NewPrivacyErr(errhandler.InvalidInputToSetBytesErr, nil)
	}
	if proofbytes[0] != proof.GetVersion() {
		return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, nil)
	}

	proof.SetVersion()
	offset := 1

	//ComOutputMultiRangeProofSize *aggregatedRangeProof
	if offset+common.Uint32Size >= len(proofbytes) {
		return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range aggregated range proof"))
	}
	lenComOutputMultiRangeUint32, _ := common.BytesToUint32(proofbytes[offset : offset+common.Uint32Size])
	lenComOutputMultiRangeProof := int(lenComOutputMultiRangeUint32)
	offset += common.Uint32Size

	if offset+lenComOutputMultiRangeProof >= len(proofbytes) {
		return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range aggregated range proof"))
	}
	if lenComOutputMultiRangeProof > 0 {
		bulletproof := &bulletproofs.AggregatedRangeProof{}
		bulletproof.Init()
		proof.aggregatedRangeProof = bulletproof
		err := proof.aggregatedRangeProof.SetBytes(proofbytes[offset : offset+lenComOutputMultiRangeProof])
		if err != nil {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, err)
		}
		offset += lenComOutputMultiRangeProof
	}

	//InputCoins  []*privacy.InputCoin
	if offset >= len(proofbytes) {
		return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range input coins"))
	}
	lenInputCoinsArray := int(proofbytes[offset])
	offset += 1
	proof.inputCoins = make([]*coin.InputCoin, lenInputCoinsArray)
	for i := 0; i < lenInputCoinsArray; i++ {
		if offset >= len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range input coins"))
		}
		lenInputCoin := int(proofbytes[offset])
		offset += 1

		proof.inputCoins[i] = new(coin.InputCoin)
		if offset+lenInputCoin >= len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range input coins"))
		}
		err := proof.inputCoins[i].SetBytes(proofbytes[offset : offset+lenInputCoin])
		if err != nil {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, err)
		}
		offset += lenInputCoin
	}

	//OutputCoins []*privacy.OutputCoin
	if offset >= len(proofbytes) {
		return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range output coins"))
	}
	lenOutputCoinsArray := int(proofbytes[offset])
	offset += 1
	proof.outputCoins = make([]*coin.OutputCoin, lenOutputCoinsArray)
	for i := 0; i < lenOutputCoinsArray; i++ {
		proof.outputCoins[i] = new(coin.OutputCoin)
		// try get 1-byte for len
		if offset >= len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range output coins"))
		}
		lenOutputCoin := int(proofbytes[offset])
		offset += 1

		if offset+lenOutputCoin >= len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range output coins"))
		}
		err := proof.outputCoins[i].SetBytes(proofbytes[offset : offset+lenOutputCoin])
		if err != nil {
			// 1-byte is wrong
			// try get 2-byte for len
			if offset+1 >= len(proofbytes) {
				return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range output coins"))
			}
			lenOutputCoin = common.BytesToInt(proofbytes[offset-1 : offset+1])
			offset += 1

			if offset+lenOutputCoin >= len(proofbytes) {
				return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range output coins"))
			}
			err1 := proof.outputCoins[i].SetBytes(proofbytes[offset : offset+lenOutputCoin])
			if err1 != nil {
				return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, err)
			}
		}
		offset += lenOutputCoin
	}

	//fmt.Printf("SETBYTES ------------------ %v\n", proof.Bytes())

	return nil
}

// Privacy TODO: update version 2 remember to fix this
func Prove(inp []*coin.InputCoin, out []*coin.OutputCoin, hasPrivacy bool) (*PaymentProofV2, error) {
	// Init proof
	proof := new(PaymentProofV2)
	proof.SetVersion()
	proof.aggregatedRangeProof.Init()
	proof.SetInputCoins(inp)
	proof.SetOutputCoins(out)

	// If not have privacy then don't need to prove range
	if !hasPrivacy {
		return proof, nil
	}

	// Prepare range proofs
	n := len(out)
	outputValues := make([]uint64, n)
	outputRands := make([]*operation.Scalar, n)
	for i := 0; i < n; i += 1 {
		outputValues[i] = out[i].CoinDetails.GetValue()
		outputRands[i] = operation.RandomScalar()
	}

	wit := new(bulletproofs.AggregatedRangeWitness)
	wit.Set(outputValues, outputRands)
	proof.aggregatedRangeProof, err = wit.Prove()
	if err != nil {
		return nil, err
	}
	return proof, nil
}

func (proof PaymentProof) verifyNoPrivacy(pubKey key.PublicKey, fee uint64, shardID byte, tokenID *common.Hash) (bool, error) {
	return true, nil
}

func (proof PaymentProof) verifyHasPrivacy(pubKey key.PublicKey, fee uint64, shardID byte, tokenID *common.Hash, isBatch bool) (bool, error) {
	return true, nil
}

func (proof PaymentProof) Verify(hasPrivacy bool, pubKey key.PublicKey, fee uint64, shardID byte, tokenID *common.Hash, isBatch bool, additionalData interface{}) (bool, error) {
	// has no privacy
	if !hasPrivacy {
		return proof.verifyNoPrivacy(pubKey, fee, shardID, tokenID)
	}

	return proof.verifyHasPrivacy(pubKey, fee, shardID, tokenID, isBatch)
}
