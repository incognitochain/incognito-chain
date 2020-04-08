package privacy_v2

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"

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

func (proof *PaymentProofV2) GetVersion() uint8                 { return 2 }
func (proof PaymentProofV2) GetInputCoins() []*coin.InputCoin   { return proof.inputCoins }
func (proof PaymentProofV2) GetOutputCoins() []*coin.OutputCoin { return proof.outputCoins }

func (proof *PaymentProofV2) SetVersion() { proof.Version = 2 }

func (proof *PaymentProofV2) SetInputCoins(v []*coin.InputCoin) error {
	n := len(v)
	inputCoins := make([]*coin.InputCoin, n)
	for i := 0; i < n; i += 1 {
		b := v[i].Bytes()
		inputCoins[i] = new(coin.InputCoin)
		err := inputCoins[i].SetBytes(b)
		if err != nil {
			return err
		}
	}
	proof.inputCoins = inputCoins
	return nil
}
func (proof *PaymentProofV2) SetOutputCoins(v []*coin.OutputCoin) error {
	n := len(v)
	outputCoins := make([]*coin.OutputCoin, n)
	for i := 0; i < n; i += 1 {
		b := v[i].Bytes()
		outputCoins[i] = new(coin.OutputCoin)
		err := outputCoins[i].SetBytes(b)
		if err != nil {
			return err
		}
	}
	proof.outputCoins = outputCoins
	return nil
}

func (proof PaymentProofV2) GetAggregatedRangeProof() agg_interface.AggregatedRangeProof {
	return proof.aggregatedRangeProof
}

func (proof *PaymentProofV2) Init() {
	aggregatedRangeProof := &bulletproofs.AggregatedRangeProof{}
	aggregatedRangeProof.Init()
	proof.aggregatedRangeProof = aggregatedRangeProof
	proof.inputCoins = []*coin.InputCoin{}
	proof.outputCoins = []*coin.OutputCoin{}
}

// MarshalJSON - override function
func (proof PaymentProofV2) MarshalJSON() ([]byte, error) {
	data := proof.Bytes()
	//temp := base58.Base58Check{}.Encode(data, common.ZeroByte)
	temp := base64.StdEncoding.EncodeToString(data)
	return json.Marshal(temp)
}

// UnmarshalJSON - override function
func (proof *PaymentProofV2) UnmarshalJSON(data []byte) error {
	dataStr := common.EmptyString
	errJson := json.Unmarshal(data, &dataStr)
	if errJson != nil {
		return errJson
	}
	//temp, _, err := base58.Base58Check{}.Decode(dataStr)
	temp, err := base64.StdEncoding.DecodeString(dataStr)
	if err != nil {
		return err
	}

	err = proof.SetBytes(temp)
	if err.(*errhandler.PrivacyError) != nil {
		return err
	}
	return nil
}

func (proof PaymentProofV2) Bytes() []byte {
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

	if offset+lenComOutputMultiRangeProof > len(proofbytes) {
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
		if offset+lenInputCoin > len(proofbytes) {
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

		if offset+lenOutputCoin > len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range output coins"))
		}
		err := proof.outputCoins[i].SetBytes(proofbytes[offset : offset+lenOutputCoin])
		if err != nil {
			// 1-byte is wrong
			// try get 2-byte for len
			if offset+1 > len(proofbytes) {
				return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range output coins"))
			}
			lenOutputCoin = common.BytesToInt(proofbytes[offset-1 : offset+1])
			offset += 1

			if offset+lenOutputCoin > len(proofbytes) {
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

func (proof *PaymentProofV2) IsPrivacy() bool {
	return true
}

func (proof PaymentProofV2) ValidateSanity() (bool, error) {
	if len(proof.GetInputCoins()) > 255 {
		return false, errors.New("Input coins in tx are very large:" + strconv.Itoa(len(proof.GetInputCoins())))
	}

	if len(proof.GetOutputCoins()) > 255 {
		return false, errors.New("Output coins in tx are very large:" + strconv.Itoa(len(proof.GetOutputCoins())))
	}

	if !proof.aggregatedRangeProof.ValidateSanity() {
		return false, errors.New("validate sanity Aggregated range proof failed")
	}

	// check output coins with privacy
	for i := 0; i < len(proof.GetOutputCoins()); i++ {
		if !proof.GetOutputCoins()[i].CoinDetails.GetPublicKey().PointValid() {
			return false, errors.New("validate sanity Public key of output coin failed")
		}
		if !proof.GetOutputCoins()[i].CoinDetails.GetCoinCommitment().PointValid() {
			return false, errors.New("validate sanity Coin commitment of output coin failed")
		}
	}
	return true, nil
}

// Privacy TODO: update version 2 remember to fix this
func Prove(inp *[]*coin.InputCoin, out *[]*coin.OutputCoin, hasPrivacy bool, paymentInfoPtr *[]*key.PaymentInfo) (*PaymentProofV2, error) {
	paymentInfo := *paymentInfoPtr
	inputCoins := *inp
	outputCoins := *out
	// Init proof
	proof := new(PaymentProofV2)
	proof.SetVersion()

	aggregateproof := new(bulletproofs.AggregatedRangeProof)
	aggregateproof.Init()
	proof.aggregatedRangeProof = aggregateproof

	if err := proof.SetInputCoins(inputCoins); err != nil {
		return nil, err
	}
	if err := proof.SetOutputCoins(outputCoins); err != nil {
		return nil, err
	}

	// If not have privacy then don't need to prove range
	if !hasPrivacy {
		return proof, nil
	}

	// Prepare range proofs
	n := len(outputCoins)
	outputValues := make([]uint64, n)
	outputRands := make([]*operation.Scalar, n)
	for i := 0; i < n; i += 1 {
		outputValues[i] = outputCoins[i].CoinDetails.GetValue()
		outputRands[i] = outputCoins[i].CoinDetails.GetRandomness()
	}

	wit := new(bulletproofs.AggregatedRangeWitness)
	wit.Set(outputValues, outputRands)
	var err error
	proof.aggregatedRangeProof, err = wit.Prove()
	if err != nil {
		return nil, err
	}

	// After Prove, we should hide all information in coin details.
	// encrypt coin details (Randomness)
	// hide information of output coins except coin commitments, public key
	for i := 0; i < len(proof.GetOutputCoins()); i++ {
		err = proof.GetOutputCoins()[i].Encrypt(paymentInfo[i].PaymentAddress.Tk)
		if err.(*errhandler.PrivacyError) != nil {
			return nil, err.(*errhandler.PrivacyError)
		}
		// proof.GetOutputCoins()[i].CoinDetails.SetSNDerivator(nil)
		proof.GetOutputCoins()[i].CoinDetails.SetSerialNumber(nil)
		proof.GetOutputCoins()[i].CoinDetails.SetValue(0)
		proof.GetOutputCoins()[i].CoinDetails.SetRandomness(nil)
	}

	for i := 0; i < len(proof.GetInputCoins()); i++ {
		proof.GetInputCoins()[i].CoinDetails.SetSerialNumber(nil)
		proof.GetInputCoins()[i].CoinDetails.SetCoinCommitment(nil)
		proof.GetInputCoins()[i].CoinDetails.SetValue(0)
		proof.GetInputCoins()[i].CoinDetails.SetSNDerivator(nil)
		proof.GetInputCoins()[i].CoinDetails.SetPublicKey(nil)
		proof.GetInputCoins()[i].CoinDetails.SetRandomness(nil)
	}

	return proof, nil
}

func (proof PaymentProofV2) verifyNoPrivacy(pubKey key.PublicKey, fee uint64, shardID byte, tokenID *common.Hash) (bool, error) {
	return true, nil
}

func (proof PaymentProofV2) verifyHasPrivacy(isBatch bool) (bool, error) {
	// Verify the proof that output values and sum of them do not exceed v_max
	if isBatch == false {
		valid, err := proof.aggregatedRangeProof.Verify()
		if !valid {
			Logger.Log.Errorf("VERIFICATION PAYMENT PROOF V2: Multi-range failed")
			return false, errhandler.NewPrivacyErr(errhandler.VerifyAggregatedProofFailedErr, err)
		}
	}
	return true, nil
}

func (proof PaymentProofV2) Verify(hasPrivacy bool, pubKey key.PublicKey, fee uint64, shardID byte, tokenID *common.Hash, isBatch bool, additionalData interface{}) (bool, error) {
	// has no privacy
	if !hasPrivacy {
		return true, nil
	}

	return proof.verifyHasPrivacy(isBatch)
}
