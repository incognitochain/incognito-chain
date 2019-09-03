package zkp

import (
	"encoding/json"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge/aggregaterange"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge/oneoutofmany"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge/serialnumbernoprivacy"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge/serialnumberprivacy"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge/utils"
)

// PaymentProof contains all of PoK for spending coin
type PaymentProof struct {
	// for input coins
	oneOfManyProof    []*oneoutofmany.OneOutOfManyProof
	serialNumberProof []*serialnumberprivacy.SNPrivacyProof
	// it is exits when tx has no privacy
	serialNumberNoPrivacyProof []*serialnumbernoprivacy.SNNoPrivacyProof

	// for output coins
	// for proving each value and sum of them are less than a threshold value
	aggregatedRangeProof *aggregaterange.AggregatedRangeProof

	inputCoins  []*privacy.InputCoin
	outputCoins []*privacy.OutputCoin

	commitmentOutputValue   []*privacy.EllipticPoint
	commitmentOutputSND     []*privacy.EllipticPoint
	commitmentOutputShardID []*privacy.EllipticPoint

	commitmentInputSecretKey *privacy.EllipticPoint
	commitmentInputValue     []*privacy.EllipticPoint
	commitmentInputSND       []*privacy.EllipticPoint
	commitmentInputShardID   *privacy.EllipticPoint

	commitmentIndices []uint64
}

// GET/SET function
func (paymentProof PaymentProof) GetOneOfManyProof() []*oneoutofmany.OneOutOfManyProof {
	return paymentProof.oneOfManyProof
}

func (paymentProof PaymentProof) GetSerialNumberProof() []*serialnumberprivacy.SNPrivacyProof {
	return paymentProof.serialNumberProof
}

func (paymentProof PaymentProof) GetSerialNumberNoPrivacyProof() []*serialnumbernoprivacy.SNNoPrivacyProof {
	return paymentProof.serialNumberNoPrivacyProof
}

func (paymentProof PaymentProof) GetAggregatedRangeProof() *aggregaterange.AggregatedRangeProof {
	return paymentProof.aggregatedRangeProof
}

func (paymentProof PaymentProof) GetCommitmentOutputValue() []*privacy.EllipticPoint {
	return paymentProof.commitmentOutputValue
}

func (paymentProof PaymentProof) GetCommitmentOutputSND() []*privacy.EllipticPoint {
	return paymentProof.commitmentOutputSND
}

func (paymentProof PaymentProof) GetCommitmentOutputShardID() []*privacy.EllipticPoint {
	return paymentProof.commitmentOutputShardID
}

func (paymentProof PaymentProof) GetCommitmentInputSecretKey() *privacy.EllipticPoint {
	return paymentProof.commitmentInputSecretKey
}

func (paymentProof PaymentProof) GetCommitmentInputValue() []*privacy.EllipticPoint {
	return paymentProof.commitmentInputValue
}

func (paymentProof PaymentProof) GetCommitmentInputSND() []*privacy.EllipticPoint {
	return paymentProof.commitmentInputSND
}

func (paymentProof PaymentProof) GetCommitmentInputShardID() *privacy.EllipticPoint {
	return paymentProof.commitmentInputShardID
}

func (paymentProof PaymentProof) GetCommitmentIndices() []uint64 {
	return paymentProof.commitmentIndices
}

func (paymentProof PaymentProof) GetInputCoins() []*privacy.InputCoin {
	return paymentProof.inputCoins
}

func (paymentProof *PaymentProof) SetInputCoins(v []*privacy.InputCoin) {
	paymentProof.inputCoins = v
}

func (paymentProof PaymentProof) GetOutputCoins() []*privacy.OutputCoin {
	return paymentProof.outputCoins
}

func (paymentProof *PaymentProof) SetOutputCoins(v []*privacy.OutputCoin) {
	paymentProof.outputCoins = v
}

// End GET/SET function

// Init
func (proof *PaymentProof) Init() {
	aggregatedRangeProof := &aggregaterange.AggregatedRangeProof{}
	aggregatedRangeProof.Init()
	proof.oneOfManyProof = []*oneoutofmany.OneOutOfManyProof{}
	proof.serialNumberProof = []*serialnumberprivacy.SNPrivacyProof{}
	proof.aggregatedRangeProof = aggregatedRangeProof
	proof.inputCoins = []*privacy.InputCoin{}
	proof.outputCoins = []*privacy.OutputCoin{}

	proof.commitmentOutputValue = []*privacy.EllipticPoint{}
	proof.commitmentOutputSND = []*privacy.EllipticPoint{}
	proof.commitmentOutputShardID = []*privacy.EllipticPoint{}

	proof.commitmentInputSecretKey = new(privacy.EllipticPoint)
	proof.commitmentInputValue = []*privacy.EllipticPoint{}
	proof.commitmentInputSND = []*privacy.EllipticPoint{}
	proof.commitmentInputShardID = new(privacy.EllipticPoint)

}

// MarshalJSON - override function
func (proof PaymentProof) MarshalJSON() ([]byte, error) {
	data := proof.Bytes()
	temp := base58.Base58Check{}.Encode(data, common.ZeroByte)
	return json.Marshal(temp)
}

// UnmarshalJSON - override function
func (proof *PaymentProof) UnmarshalJSON(data []byte) error {
	dataStr := common.EmptyString
	errJson := json.Unmarshal(data, &dataStr)
	if errJson != nil {
		return errJson
	}
	temp, _, err := base58.Base58Check{}.Decode(dataStr)
	if err != nil {
		return err
	}

	err = proof.SetBytes(temp)
	if err.(*privacy.PrivacyError) != nil {
		return err
	}
	return nil
}

func (proof *PaymentProof) Bytes() []byte {
	var bytes []byte
	hasPrivacy := len(proof.oneOfManyProof) > 0

	// OneOfManyProofSize
	bytes = append(bytes, byte(len(proof.oneOfManyProof)))
	for i := 0; i < len(proof.oneOfManyProof); i++ {
		oneOfManyProof := proof.oneOfManyProof[i].Bytes()
		bytes = append(bytes, common.IntToBytes(utils.OneOfManyProofSize)...)
		bytes = append(bytes, oneOfManyProof...)
	}

	// SerialNumberProofSize
	bytes = append(bytes, byte(len(proof.serialNumberProof)))
	for i := 0; i < len(proof.serialNumberProof); i++ {
		serialNumberProof := proof.serialNumberProof[i].Bytes()
		bytes = append(bytes, common.IntToBytes(utils.SnPrivacyProofSize)...)
		bytes = append(bytes, serialNumberProof...)
	}

	// SNNoPrivacyProofSize
	bytes = append(bytes, byte(len(proof.serialNumberNoPrivacyProof)))
	for i := 0; i < len(proof.serialNumberNoPrivacyProof); i++ {
		snNoPrivacyProof := proof.serialNumberNoPrivacyProof[i].Bytes()
		bytes = append(bytes, byte(utils.SnNoPrivacyProofSize))
		bytes = append(bytes, snNoPrivacyProof...)
	}

	//ComOutputMultiRangeProofSize
	if hasPrivacy {
		comOutputMultiRangeProof := proof.aggregatedRangeProof.Bytes()
		bytes = append(bytes, common.IntToBytes(len(comOutputMultiRangeProof))...)
		bytes = append(bytes, comOutputMultiRangeProof...)
	} else {
		bytes = append(bytes, []byte{0, 0}...)
	}

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
		bytes = append(bytes, byte(lenOutputCoins))
		bytes = append(bytes, outputCoins...)
	}

	// ComOutputValue
	bytes = append(bytes, byte(len(proof.commitmentOutputValue)))
	for i := 0; i < len(proof.commitmentOutputValue); i++ {
		comOutputValue := proof.commitmentOutputValue[i].Compress()
		bytes = append(bytes, byte(privacy.CompressedEllipticPointSize))
		bytes = append(bytes, comOutputValue...)
	}

	// ComOutputSND
	bytes = append(bytes, byte(len(proof.commitmentOutputSND)))
	for i := 0; i < len(proof.commitmentOutputSND); i++ {
		comOutputSND := proof.commitmentOutputSND[i].Compress()
		bytes = append(bytes, byte(privacy.CompressedEllipticPointSize))
		bytes = append(bytes, comOutputSND...)
	}

	// ComOutputShardID
	bytes = append(bytes, byte(len(proof.commitmentOutputShardID)))
	for i := 0; i < len(proof.commitmentOutputShardID); i++ {
		comOutputShardID := proof.commitmentOutputShardID[i].Compress()
		bytes = append(bytes, byte(privacy.CompressedEllipticPointSize))
		bytes = append(bytes, comOutputShardID...)
	}

	//ComInputSK 				*privacy.EllipticPoint
	if proof.commitmentInputSecretKey != nil {
		comInputSK := proof.commitmentInputSecretKey.Compress()
		bytes = append(bytes, byte(privacy.CompressedEllipticPointSize))
		bytes = append(bytes, comInputSK...)
	} else {
		bytes = append(bytes, byte(0))
	}

	//ComInputValue 		[]*privacy.EllipticPoint
	bytes = append(bytes, byte(len(proof.commitmentInputValue)))
	for i := 0; i < len(proof.commitmentInputValue); i++ {
		comInputValue := proof.commitmentInputValue[i].Compress()
		bytes = append(bytes, byte(privacy.CompressedEllipticPointSize))
		bytes = append(bytes, comInputValue...)
	}

	//ComInputSND 			[]*privacy.EllipticPoint
	bytes = append(bytes, byte(len(proof.commitmentInputSND)))
	for i := 0; i < len(proof.commitmentInputSND); i++ {
		comInputSND := proof.commitmentInputSND[i].Compress()
		bytes = append(bytes, byte(privacy.CompressedEllipticPointSize))
		bytes = append(bytes, comInputSND...)
	}

	//ComInputShardID 	*privacy.EllipticPoint
	if proof.commitmentInputShardID != nil {
		comInputShardID := proof.commitmentInputShardID.Compress()
		bytes = append(bytes, byte(privacy.CompressedEllipticPointSize))
		bytes = append(bytes, comInputShardID...)
	} else {
		bytes = append(bytes, byte(0))
	}

	// convert commitment index to bytes array
	for i := 0; i < len(proof.commitmentIndices); i++ {
		bytes = append(bytes, common.AddPaddingBigInt(big.NewInt(int64(proof.commitmentIndices[i])), common.Uint64Size)...)
	}
	//fmt.Printf("BYTES ------------------ %v\n", bytes)
	//fmt.Printf("LEN BYTES ------------------ %v\n", len(bytes))

	return bytes
}

func (proof *PaymentProof) SetBytes(proofbytes []byte) *privacy.PrivacyError {
	if len(proofbytes) == 0 {
		return privacy.NewPrivacyErr(privacy.InvalidInputToSetBytesErr, nil)
	}

	offset := 0

	// Set OneOfManyProofSize
	lenOneOfManyProofArray := int(proofbytes[offset])
	offset += 1
	proof.oneOfManyProof = make([]*oneoutofmany.OneOutOfManyProof, lenOneOfManyProofArray)
	for i := 0; i < lenOneOfManyProofArray; i++ {
		lenOneOfManyProof := common.BytesToInt(proofbytes[offset : offset+2])
		offset += 2
		proof.oneOfManyProof[i] = new(oneoutofmany.OneOutOfManyProof).Init()
		err := proof.oneOfManyProof[i].SetBytes(proofbytes[offset : offset+lenOneOfManyProof])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenOneOfManyProof
	}

	// Set serialNumberProofSize
	lenSerialNumberProofArray := int(proofbytes[offset])
	offset += 1
	proof.serialNumberProof = make([]*serialnumberprivacy.SNPrivacyProof, lenSerialNumberProofArray)
	for i := 0; i < lenSerialNumberProofArray; i++ {
		lenSerialNumberProof := common.BytesToInt(proofbytes[offset : offset+2])
		offset += 2
		proof.serialNumberProof[i] = new(serialnumberprivacy.SNPrivacyProof).Init()
		err := proof.serialNumberProof[i].SetBytes(proofbytes[offset : offset+lenSerialNumberProof])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenSerialNumberProof
	}

	// Set SNNoPrivacyProofSize
	lenSNNoPrivacyProofArray := int(proofbytes[offset])
	offset += 1
	proof.serialNumberNoPrivacyProof = make([]*serialnumbernoprivacy.SNNoPrivacyProof, lenSNNoPrivacyProofArray)
	for i := 0; i < lenSNNoPrivacyProofArray; i++ {
		lenSNNoPrivacyProof := int(proofbytes[offset])
		offset += 1
		proof.serialNumberNoPrivacyProof[i] = new(serialnumbernoprivacy.SNNoPrivacyProof).Init()
		err := proof.serialNumberNoPrivacyProof[i].SetBytes(proofbytes[offset : offset+lenSNNoPrivacyProof])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenSNNoPrivacyProof
	}

	//ComOutputMultiRangeProofSize *aggregatedRangeProof
	lenComOutputMultiRangeProof := common.BytesToInt(proofbytes[offset : offset+2])
	offset += 2
	if lenComOutputMultiRangeProof > 0 {
		aggregatedRangeProof := &aggregaterange.AggregatedRangeProof{}
		aggregatedRangeProof.Init()
		proof.aggregatedRangeProof = aggregatedRangeProof
		err := proof.aggregatedRangeProof.SetBytes(proofbytes[offset : offset+lenComOutputMultiRangeProof])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComOutputMultiRangeProof
	}

	//InputCoins  []*privacy.InputCoin
	lenInputCoinsArray := int(proofbytes[offset])
	offset += 1
	proof.inputCoins = make([]*privacy.InputCoin, lenInputCoinsArray)
	for i := 0; i < lenInputCoinsArray; i++ {
		lenInputCoin := int(proofbytes[offset])
		offset += 1
		proof.inputCoins[i] = new(privacy.InputCoin)
		err := proof.inputCoins[i].SetBytes(proofbytes[offset : offset+lenInputCoin])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenInputCoin
	}

	//OutputCoins []*privacy.OutputCoin
	lenOutputCoinsArray := int(proofbytes[offset])
	offset += 1
	proof.outputCoins = make([]*privacy.OutputCoin, lenOutputCoinsArray)
	for i := 0; i < lenOutputCoinsArray; i++ {
		lenOutputCoin := int(proofbytes[offset])
		offset += 1
		proof.outputCoins[i] = new(privacy.OutputCoin)
		err := proof.outputCoins[i].SetBytes(proofbytes[offset : offset+lenOutputCoin])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenOutputCoin
	}
	//ComOutputValue   []*privacy.EllipticPoint
	lenComOutputValueArray := int(proofbytes[offset])
	offset += 1
	proof.commitmentOutputValue = make([]*privacy.EllipticPoint, lenComOutputValueArray)
	for i := 0; i < lenComOutputValueArray; i++ {
		lenComOutputValue := int(proofbytes[offset])
		offset += 1
		proof.commitmentOutputValue[i] = new(privacy.EllipticPoint)
		err := proof.commitmentOutputValue[i].Decompress(proofbytes[offset : offset+lenComOutputValue])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComOutputValue
	}
	//ComOutputSND     []*privacy.EllipticPoint
	lenComOutputSNDArray := int(proofbytes[offset])
	offset += 1
	proof.commitmentOutputSND = make([]*privacy.EllipticPoint, lenComOutputSNDArray)
	for i := 0; i < lenComOutputSNDArray; i++ {
		lenComOutputSND := int(proofbytes[offset])
		offset += 1
		proof.commitmentOutputSND[i] = new(privacy.EllipticPoint)
		err := proof.commitmentOutputSND[i].Decompress(proofbytes[offset : offset+lenComOutputSND])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComOutputSND
	}

	lenComOutputShardIdArray := int(proofbytes[offset])
	offset += 1
	proof.commitmentOutputShardID = make([]*privacy.EllipticPoint, lenComOutputShardIdArray)
	for i := 0; i < lenComOutputShardIdArray; i++ {
		lenComOutputShardId := int(proofbytes[offset])
		offset += 1
		proof.commitmentOutputShardID[i] = new(privacy.EllipticPoint)
		err := proof.commitmentOutputShardID[i].Decompress(proofbytes[offset : offset+lenComOutputShardId])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComOutputShardId
	}

	//ComInputSK 				*privacy.EllipticPoint
	lenComInputSK := int(proofbytes[offset])
	offset += 1
	if lenComInputSK > 0 {
		proof.commitmentInputSecretKey = new(privacy.EllipticPoint)
		err := proof.commitmentInputSecretKey.Decompress(proofbytes[offset : offset+lenComInputSK])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComInputSK
	}
	//ComInputValue 		[]*privacy.EllipticPoint
	lenComInputValueArr := int(proofbytes[offset])
	offset += 1
	proof.commitmentInputValue = make([]*privacy.EllipticPoint, lenComInputValueArr)
	for i := 0; i < lenComInputValueArr; i++ {
		lenComInputValue := int(proofbytes[offset])
		offset += 1
		proof.commitmentInputValue[i] = new(privacy.EllipticPoint)
		err := proof.commitmentInputValue[i].Decompress(proofbytes[offset : offset+lenComInputValue])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComInputValue
	}
	//ComInputSND 			[]*privacy.EllipticPoint
	lenComInputSNDArr := int(proofbytes[offset])
	offset += 1
	proof.commitmentInputSND = make([]*privacy.EllipticPoint, lenComInputSNDArr)
	for i := 0; i < lenComInputSNDArr; i++ {
		lenComInputSND := int(proofbytes[offset])
		offset += 1
		proof.commitmentInputSND[i] = new(privacy.EllipticPoint)
		err := proof.commitmentInputSND[i].Decompress(proofbytes[offset : offset+lenComInputSND])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComInputSND
	}
	//ComInputShardID 	*privacy.EllipticPoint
	lenComInputShardID := int(proofbytes[offset])
	offset += 1
	if lenComInputShardID > 0 {
		proof.commitmentInputShardID = new(privacy.EllipticPoint)
		err := proof.commitmentInputShardID.Decompress(proofbytes[offset : offset+lenComInputShardID])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComInputShardID
	}

	// get commitments list
	proof.commitmentIndices = make([]uint64, len(proof.oneOfManyProof)*privacy.CommitmentRingSize)
	for i := 0; i < len(proof.oneOfManyProof)*privacy.CommitmentRingSize; i++ {
		proof.commitmentIndices[i] = new(big.Int).SetBytes(proofbytes[offset : offset+common.Uint64Size]).Uint64()
		offset = offset + common.Uint64Size
	}

	//fmt.Printf("SETBYTES ------------------ %v\n", proof.Bytes())

	return nil
}

func (proof PaymentProof) verifyNoPrivacy(pubKey privacy.PublicKey, fee uint64, db database.DatabaseInterface, shardID byte, tokenID *common.Hash) (bool, error) {
	var sumInputValue, sumOutputValue uint64
	sumInputValue = 0
	sumOutputValue = 0

	pubKeyLastByteSender := pubKey[len(pubKey)-1]
	senderShardID := common.GetShardIDFromLastByte(pubKeyLastByteSender)
	cmShardIDSender := privacy.PedCom.G[privacy.PedersenShardIDIndex].ScalarMult(new(big.Int).SetBytes([]byte{senderShardID}))

	for i := 0; i < len(proof.inputCoins); i++ {
		// Check input coins' Serial number is created from input coins' input and sender's spending key
		valid, err := proof.serialNumberNoPrivacyProof[i].Verify(nil)
		if !valid {
			privacy.Logger.Log.Errorf("Verify serial number no privacy proof failed")
			return false, privacy.NewPrivacyErr(privacy.VerifySerialNumberNoPrivacyProofFailedErr, err)
		}

		// Check input coins' cm is calculated correctly
		cmTmp := proof.inputCoins[i].CoinDetails.GetPublicKey()
		cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.PedersenValueIndex].ScalarMult(big.NewInt(int64(proof.inputCoins[i].CoinDetails.GetValue()))))
		cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.PedersenSndIndex].ScalarMult(proof.inputCoins[i].CoinDetails.GetSNDerivator()))
		cmTmp = cmTmp.Add(cmShardIDSender)
		cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.PedersenRandomnessIndex].ScalarMult(proof.inputCoins[i].CoinDetails.GetRandomness()))
		if !cmTmp.IsEqual(proof.inputCoins[i].CoinDetails.GetCoinCommitment()) {
			privacy.Logger.Log.Errorf("Input coins %v commitment wrong!\n", i)
			return false, privacy.NewPrivacyErr(privacy.VerifyCoinCommitmentInputFailedErr, nil)
		}

		// Calculate sum of input values
		sumInputValue += proof.inputCoins[i].CoinDetails.GetValue()
	}

	for i := 0; i < len(proof.outputCoins); i++ {
		// Check output coins' cm is calculated correctly
		cmTmp := proof.outputCoins[i].CoinDetails.GetPublicKey()
		cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.PedersenValueIndex].ScalarMult(big.NewInt(int64(proof.outputCoins[i].CoinDetails.GetValue()))))
		cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.PedersenSndIndex].ScalarMult(proof.outputCoins[i].CoinDetails.GetSNDerivator()))
		shardID := common.GetShardIDFromLastByte(proof.outputCoins[i].CoinDetails.GetPubKeyLastByte())
		cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.PedersenShardIDIndex].ScalarMult(new(big.Int).SetBytes([]byte{shardID})))
		cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.PedersenRandomnessIndex].ScalarMult(proof.outputCoins[i].CoinDetails.GetRandomness()))
		if !cmTmp.IsEqual(proof.outputCoins[i].CoinDetails.GetCoinCommitment()) {
			privacy.Logger.Log.Errorf("Output coins %v commitment wrong!\n", i)
			return false, privacy.NewPrivacyErr(privacy.VerifyCoinCommitmentOutputFailedErr, nil)
		}

		// Calculate sum of output values
		sumOutputValue += proof.outputCoins[i].CoinDetails.GetValue()
	}

	// check if sum of input values equal sum of output values
	if sumInputValue != sumOutputValue+fee {
		privacy.Logger.Log.Debugf("sumInputValue: %v\n", sumInputValue)
		privacy.Logger.Log.Debugf("sumOutputValue: %v\n", sumOutputValue)
		privacy.Logger.Log.Debugf("fee: %v\n", fee)
		privacy.Logger.Log.Errorf("Sum of inputs is not equal sum of output!\n")
		return false, privacy.NewPrivacyErr(privacy.VerifyAmountNoPrivacyFailedErr, nil)
	}
	return true, nil
}

func (proof PaymentProof) verifyHasPrivacy(pubKey privacy.PublicKey, fee uint64, db database.DatabaseInterface, shardID byte, tokenID *common.Hash) (bool, error) {
	// verify for input coins
	cmInputSum := make([]*privacy.EllipticPoint, len(proof.oneOfManyProof))
	for i := 0; i < len(proof.oneOfManyProof); i++ {
		privacy.Logger.Log.Infof("[TEST] input coins %v\n ShardID %v fee %v", i, shardID, fee)
		privacy.Logger.Log.Infof("[TEST] commitments indices %v\n", proof.commitmentIndices[i*privacy.CommitmentRingSize:i*privacy.CommitmentRingSize+8])
		// Verify for the proof one-out-of-N commitments is a commitment to the coins being spent
		// Calculate cm input sum
		cmInputSum[i] = new(privacy.EllipticPoint)
		cmInputSum[i] = proof.commitmentInputSecretKey.Add(proof.commitmentInputValue[i])
		cmInputSum[i] = cmInputSum[i].Add(proof.commitmentInputSND[i])
		cmInputSum[i] = cmInputSum[i].Add(proof.commitmentInputShardID)

		// get commitments list from CommitmentIndices
		commitments := make([]*privacy.EllipticPoint, privacy.CommitmentRingSize)
		for j := 0; j < privacy.CommitmentRingSize; j++ {
			index := proof.commitmentIndices[i*privacy.CommitmentRingSize+j]
			commitmentBytes, err := db.GetCommitmentByIndex(*tokenID, index, shardID)
			privacy.Logger.Log.Infof("[TEST] commitment at index %v: %v\n", index, commitmentBytes)
			if err != nil {
				privacy.Logger.Log.Errorf("VERIFICATION PAYMENT PROOF 1: Error when get commitment by index from database", index, err)
				return false, privacy.NewPrivacyErr(privacy.VerifyOneOutOfManyProofFailedErr, err)
			}
			recheckIndex, err := db.GetCommitmentIndex(*tokenID, commitmentBytes, shardID)
			if err != nil || recheckIndex.Uint64() != index {
				privacy.Logger.Log.Errorf("VERIFICATION PAYMENT PROOF 2: Error when get commitment by index from database", index, err)
				return false, privacy.NewPrivacyErr(privacy.VerifyOneOutOfManyProofFailedErr, err)
			}

			commitments[j] = new(privacy.EllipticPoint)
			err = commitments[j].Decompress(commitmentBytes)
			if err != nil {
				privacy.Logger.Log.Errorf("VERIFICATION PAYMENT PROOF: Cannot decompress commitment from database", index, err)
				return false, privacy.NewPrivacyErr(privacy.VerifyOneOutOfManyProofFailedErr, err)
			}

			commitments[j], err = commitments[j].Sub(cmInputSum[i])
			if err != nil {
				privacy.Logger.Log.Errorf("VERIFICATION PAYMENT PROOF: Cannot sub commitment to sum of commitment inputs", index, err)
				return false, privacy.NewPrivacyErr(privacy.VerifyOneOutOfManyProofFailedErr, err)
			}
		}

		proof.oneOfManyProof[i].Statement.Commitments = commitments

		valid, err := proof.oneOfManyProof[i].Verify()
		if !valid {
			privacy.Logger.Log.Errorf("VERIFICATION PAYMENT PROOF: One out of many failed")
			return false, privacy.NewPrivacyErr(privacy.VerifyOneOutOfManyProofFailedErr, err)
		}
		// Verify for the Proof that input coins' serial number is derived from the committed derivator
		valid, err = proof.serialNumberProof[i].Verify(nil)
		if !valid {
			privacy.Logger.Log.Errorf("VERIFICATION PAYMENT PROOF: Serial number privacy failed")
			return false, privacy.NewPrivacyErr(privacy.VerifySerialNumberPrivacyProofFailedErr, err)
		}
	}

	// Check output coins' cm is calculated correctly
	for i := 0; i < len(proof.outputCoins); i++ {
		cmTmp := proof.outputCoins[i].CoinDetails.GetPublicKey().Add(proof.commitmentOutputValue[i])
		cmTmp = cmTmp.Add(proof.commitmentOutputSND[i])
		cmTmp = cmTmp.Add(proof.commitmentOutputShardID[i])

		if !cmTmp.IsEqual(proof.outputCoins[i].CoinDetails.GetCoinCommitment()) {
			privacy.Logger.Log.Errorf("VERIFICATION PAYMENT PROOF: Commitment for output coins are not computed correctly")
			return false, privacy.NewPrivacyErr(privacy.VerifyCoinCommitmentOutputFailedErr, nil)
		}
	}

	// Verify the proof that output values and sum of them do not exceed v_max
	valid, err := proof.aggregatedRangeProof.Verify()
	if !valid {
		privacy.Logger.Log.Errorf("VERIFICATION PAYMENT PROOF: Multi-range failed")
		return false, privacy.NewPrivacyErr(privacy.VerifyAggregatedProofFailedErr, err)
	}

	// Verify the proof that sum of all input values is equal to sum of all output values
	comInputValueSum := new(privacy.EllipticPoint)
	comInputValueSum.Zero()
	for i := 0; i < len(proof.commitmentInputValue); i++ {
		comInputValueSum = comInputValueSum.Add(proof.commitmentInputValue[i])
	}

	comOutputValueSum := new(privacy.EllipticPoint)
	comOutputValueSum.Zero()
	for i := 0; i < len(proof.commitmentOutputValue); i++ {
		comOutputValueSum = comOutputValueSum.Add(proof.commitmentOutputValue[i])
	}

	if fee > 0 {
		comOutputValueSum = comOutputValueSum.Add(privacy.PedCom.G[privacy.PedersenValueIndex].ScalarMult(big.NewInt(int64(fee))))
	}

	privacy.Logger.Log.Debugf("comInputValueSum: ", comInputValueSum)
	privacy.Logger.Log.Debugf("comOutputValueSum: ", comOutputValueSum)

	if !comInputValueSum.IsEqual(comOutputValueSum) {
		privacy.Logger.Log.Debugf("comInputValueSum: ", comInputValueSum)
		privacy.Logger.Log.Debugf("comOutputValueSum: ", comOutputValueSum)
		privacy.Logger.Log.Error("VERIFICATION PAYMENT PROOF: Sum of input coins' value is not equal to sum of output coins' value")
		return false, privacy.NewPrivacyErr(privacy.VerifyAmountPrivacyFailedErr, nil)
	}

	return true, nil
}

func (proof PaymentProof) Verify(hasPrivacy bool, pubKey privacy.PublicKey, fee uint64, db database.DatabaseInterface, shardID byte, tokenID *common.Hash) (bool, error) {
	// has no privacy
	if !hasPrivacy {
		return proof.verifyNoPrivacy(pubKey, fee, db, shardID, tokenID)
	}

	return proof.verifyHasPrivacy(pubKey, fee, db, shardID, tokenID)
}
