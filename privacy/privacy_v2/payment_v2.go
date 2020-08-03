package privacy_v2

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/wallet"
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
	inputCoins           []coin.PlainCoin
	outputCoins          []*coin.CoinV2
}

func (proof *PaymentProofV2) SetVersion() { proof.Version = 2 }
func (proof *PaymentProofV2) GetVersion() uint8 { return 2 }

func (proof PaymentProofV2) GetInputCoins() []coin.PlainCoin { return proof.inputCoins }
func (proof PaymentProofV2) GetOutputCoins() []coin.Coin {
	res := make([]coin.Coin, len(proof.outputCoins))
	for i := 0; i < len(proof.outputCoins); i += 1 {
		res[i] = proof.outputCoins[i]
	}
	return res
}

func (proof *PaymentProofV2) SetInputCoins(v []coin.PlainCoin) error {
	var err error
	proof.inputCoins = make([]coin.PlainCoin, len(v))
	for i := 0; i < len(v); i += 1 {
		b := v[i].Bytes()
		if proof.inputCoins[i], err = coin.NewPlainCoinFromByte(b); err != nil {
			Logger.Log.Errorf("Proofv2 cannot create inputCoins from new plain coin from bytes: err %v", err)
			return err
		}
	}
	return nil
}

func (proof *PaymentProofV2) SetOutputCoinsV2(v []*coin.CoinV2) error {
	var err error
	proof.outputCoins = make([]*coin.CoinV2, len(v))
	for i := 0; i < len(v); i += 1 {
		b := v[i].Bytes()
		proof.outputCoins[i] = new(coin.CoinV2)
		if err = proof.outputCoins[i].SetBytes(b); err != nil {
			Logger.Log.Errorf("Proofv2 cannot set byte to outputCoins : err %v", err)
			return err
		}
	}
	return nil
}

// v should be all coinv2 or else it would crash
func (proof *PaymentProofV2) SetOutputCoins(v []coin.Coin) error {
	var err error
	proof.outputCoins = make([]*coin.CoinV2, len(v))
	for i := 0; i < len(v); i += 1 {
		proof.outputCoins[i] = new(coin.CoinV2)
		b := v[i].Bytes()
		if err = proof.outputCoins[i].SetBytes(b); err != nil {
			Logger.Log.Errorf("Proofv2 cannot set byte to outputCoins : err %v", err)
			return err
		}
	}
	return nil
}

func (proof PaymentProofV2) GetAggregatedRangeProof() agg_interface.AggregatedRangeProof {
	return proof.aggregatedRangeProof
}

func (proof *PaymentProofV2) Init() {
	aggregatedRangeProof := &bulletproofs.AggregatedRangeProof{}
	aggregatedRangeProof.Init()
	proof.Version = 2
	proof.aggregatedRangeProof = aggregatedRangeProof
	proof.inputCoins = []coin.PlainCoin{}
	proof.outputCoins = []*coin.CoinV2{}
}

func (proof PaymentProofV2) MarshalJSON() ([]byte, error) {
	data := proof.Bytes()
	//temp := base58.Base58Check{}.Encode(data, common.ZeroByte)
	temp := base64.StdEncoding.EncodeToString(data)
	return json.Marshal(temp)
}

func (proof *PaymentProofV2) UnmarshalJSON(data []byte) error {
	dataStr := common.EmptyString
	errJson := json.Unmarshal(data, &dataStr)
	if errJson != nil {
		return errJson
	}
	temp, err := base64.StdEncoding.DecodeString(dataStr)
	if err != nil {
		return err
	}

	errSetBytes := proof.SetBytes(temp)
	if errSetBytes != nil {
		return errSetBytes
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
		return errhandler.NewPrivacyErr(errhandler.InvalidInputToSetBytesErr, errors.New("Proof bytes is zero"))
	}
	if proofbytes[0] != proof.GetVersion() {
		return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Proof bytes version is incorrect"))
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

	//InputCoins  []*coin.PlainCoinV1
	if offset >= len(proofbytes) {
		return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range input coins"))
	}
	lenInputCoinsArray := int(proofbytes[offset])
	offset += 1
	proof.inputCoins = make([]coin.PlainCoin, lenInputCoinsArray)
	for i := 0; i < lenInputCoinsArray; i++ {
		if offset >= len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range input coins"))
		}
		var err error

		lenInputCoin := int(proofbytes[offset])
		offset += 1

		if offset+lenInputCoin > len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range input coins"))
		}
		coinBytes := proofbytes[offset : offset+lenInputCoin]
		proof.inputCoins[i], err = coin.NewPlainCoinFromByte(coinBytes)
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
	proof.outputCoins = make([]*coin.CoinV2, lenOutputCoinsArray)
	for i := 0; i < lenOutputCoinsArray; i++ {
		proof.outputCoins[i] = new(coin.CoinV2)
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
	return proof.GetOutputCoins()[0].IsEncrypted()
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
	duplicatePublicKeys := make(map[string]bool)
	outputCoins := proof.GetOutputCoins()
	// cmsValues := proof.aggregatedRangeProof.GetCommitments()
	for i, outputCoin := range outputCoins {
		if outputCoin.GetPublicKey()==nil || !outputCoin.GetPublicKey().PointValid() {
			return false, errors.New("validate sanity Public key of output coin failed")
		}

		//check duplicate output addresses
		pubkeyStr := string(outputCoin.GetPublicKey().ToBytesS())
		if _, ok := duplicatePublicKeys[pubkeyStr]; ok {
			return false, errors.New("Cannot have duplicate publickey ")
		}
		duplicatePublicKeys[pubkeyStr] = true

		if !outputCoin.GetCommitment().PointValid() {
			return false, errors.New("validate sanity Coin commitment of output coin failed")
		}

		//re-compute the commitment if the output coin's address is the burning address
		if wallet.IsPublicKeyBurningAddress(outputCoins[i].GetPublicKey().ToBytesS()){
			value := outputCoin.GetValue()
			rand := outputCoin.GetRandomness()
			commitment := operation.PedCom.CommitAtIndex(new(operation.Scalar).FromUint64(value), rand, coin.PedersenValueIndex)
			if !operation.IsPointEqual(commitment, outputCoin.GetCommitment()){
				return false, errors.New("validate sanity Coin commitment of burned coin failed")
			}
		}
	}
	return true, nil
}

func Prove(inputCoins []coin.PlainCoin, outputCoins []*coin.CoinV2, hasPrivacy bool, paymentInfo []*key.PaymentInfo) (*PaymentProofV2, error) {
	var err error

	proof := new(PaymentProofV2)
	proof.SetVersion()
	aggregateproof := new(bulletproofs.AggregatedRangeProof)
	aggregateproof.Init()
	proof.aggregatedRangeProof = aggregateproof
	if err = proof.SetInputCoins(inputCoins); err != nil {
		Logger.Log.Errorf("Cannot set input coins in payment_v2 proof: err %v", err)
		return nil, err
	}
	if err = proof.SetOutputCoinsV2(outputCoins); err != nil {
		Logger.Log.Errorf("Cannot set output coins in payment_v2 proof: err %v", err)
		return nil, err
	}

	// Prepare range proofs
	n := len(outputCoins)
	outputValues := make([]uint64, n)
	outputRands := make([]*operation.Scalar, n)
	for i := 0; i < n; i += 1 {
		outputValues[i] = outputCoins[i].GetValue()
		outputRands[i] = outputCoins[i].GetRandomness()
	}

	wit := new(bulletproofs.AggregatedRangeWitness)
	wit.Set(outputValues, outputRands)
	proof.aggregatedRangeProof, err = wit.Prove()
	if err != nil {
		return nil, err
	}

	// After Prove, we should hide all information in coin details.
	for i, outputCoin := range proof.outputCoins {
		if !wallet.IsPublicKeyBurningAddress(outputCoin.GetPublicKey().ToBytesS()){
			if err = outputCoin.ConcealOutputCoin(paymentInfo[i].PaymentAddress.GetPublicView()); err != nil {
				return nil, err
			}

			// OutputCoin.GetKeyImage should be nil even though we do not have it
			// Because otherwise the RPC server will return the Bytes of [1 0 0 0 0 ...] (the default byte)
			proof.outputCoins[i].SetKeyImage(nil)
		}

	}

	for _, inputCoin := range proof.GetInputCoins(){
		coin, ok := inputCoin.(*coin.CoinV2)
		if !ok {
			return nil, errors.New("Input coin of PaymentProofV2 must be CoinV2")
		}
		coin.ConcealInputCoin()
	}

	return proof, nil
}

// TODO PRIVACY (recheck before devnet)
func (proof PaymentProofV2) verifyHasPrivacy(isBatch bool) (bool, error) {
	cmsValues := proof.aggregatedRangeProof.GetCommitments()
	if len(proof.GetOutputCoins())!=len(cmsValues){
		return false, errors.New("Commitment length mismatch")
	}
	// Verify the proof that output values and sum of them do not exceed v_max
	for i := 0; i < len(proof.outputCoins); i += 1 {

		if !proof.outputCoins[i].IsEncrypted() {
			if wallet.IsPublicKeyBurningAddress(proof.outputCoins[i].GetPublicKey().ToBytesS()) {
				continue
			}
			return false, errors.New("Verify has privacy should have every coin encrypted")
		}
		//check if output coins' commitment is the same as in the proof
		if !operation.IsPointEqual(cmsValues[i], proof.outputCoins[i].GetCommitment()){
			return false, errors.New("Coin & Proof Commitments mismatch")
		}
	}
	if isBatch == false {
		valid, err := proof.aggregatedRangeProof.Verify()
		if !valid {
			Logger.Log.Errorf("VERIFICATION PAYMENT PROOF V2: Multi-range failed")
			return false, errhandler.NewPrivacyErr(errhandler.VerifyAggregatedProofFailedErr, err)
		}
	}
	return true, nil
}

func (proof PaymentProofV2) verifyHasNoPrivacy(fee uint64) (bool, error) {
	sumInput, sumOutput := uint64(0), uint64(0)
	for i := 0; i < len(proof.inputCoins); i += 1 {
		if proof.inputCoins[i].IsEncrypted() {
			return false, errors.New("Verify has no privacy should not have any coin encrypted")
		}
		sumInput += proof.inputCoins[i].GetValue()
	}
	for i := 0; i < len(proof.outputCoins); i += 1 {
		if proof.outputCoins[i].IsEncrypted() {
			return false, errors.New("Verify has no privacy should not have any coin encrypted")
		}
		sumOutput += proof.outputCoins[i].GetValue()
	}
	tmpSum := sumOutput + fee
	if tmpSum < sumOutput || tmpSum < fee {
		return false, errors.New(fmt.Sprintf("Overflown sumOutput+fee: output value = %v, fee = %v, tmpSum = %v\n", sumOutput, fee, tmpSum))
	}

	if sumInput != tmpSum {
		return false, errors.New("VerifyHasNo privacy has sum input different from sum output with fee")
	}
	return true, nil
}

func (proof PaymentProofV2) Verify(hasPrivacy bool, pubKey key.PublicKey, fee uint64, shardID byte, tokenID *common.Hash, isBatch bool, additionalData interface{}) (bool, error) {
	inputCoins := proof.GetInputCoins()
	dupMap := make(map[string]bool)
	for _,coin := range inputCoins{
		identifier := base64.StdEncoding.EncodeToString(coin.GetKeyImage().ToBytesS())
		_, exists := dupMap[identifier]
		if exists{
			return false, errors.New("Duplicate input coin in PaymentProofV2")
		}
		dupMap[identifier] = true
	}

	return proof.verifyHasPrivacy(isBatch)
}
