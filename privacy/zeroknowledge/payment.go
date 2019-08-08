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
	OneOfManyProof    []*oneoutofmany.OneOutOfManyProof
	SerialNumberProof []*serialnumberprivacy.SNPrivacyProof
	// it is exits when tx has no privacy
	SerialNumberNoPrivacyProof []*serialnumbernoprivacy.SNNoPrivacyProof

	// for output coins
	// for proving each value and sum of them are less than a threshold value
	AggregatedRangeProof *aggregaterange.AggregatedRangeProof

	InputCoins  []*privacy.InputCoin
	OutputCoins []*privacy.OutputCoin

	CommitmentOutputValue   []*privacy.EllipticPoint
	CommitmentOutputSND     []*privacy.EllipticPoint
	CommitmentOutputShardID []*privacy.EllipticPoint

	CommitmentInputSecretKey *privacy.EllipticPoint
	CommitmentInputValue     []*privacy.EllipticPoint
	CommitmentInputSND       []*privacy.EllipticPoint
	CommitmentInputShardID   *privacy.EllipticPoint

	CommitmentIndices []uint64
}

// GET/SET function
func (paymentProof PaymentProof) GetOneOfManyProof() []*oneoutofmany.OneOutOfManyProof {
	return paymentProof.OneOfManyProof
}

func (paymentProof PaymentProof) GetSerialNumberProof() []*serialnumberprivacy.SNPrivacyProof {
	return paymentProof.SerialNumberProof
}

func (paymentProof PaymentProof) GetSerialNumberNoPrivacyProof() []*serialnumbernoprivacy.SNNoPrivacyProof {
	return paymentProof.SerialNumberNoPrivacyProof
}

func (paymentProof PaymentProof) GetAggregatedRangeProof() *aggregaterange.AggregatedRangeProof {
	return paymentProof.AggregatedRangeProof
}

func (paymentProof PaymentProof) GetCommitmentOutputValue() []*privacy.EllipticPoint {
	return paymentProof.CommitmentOutputValue
}

func (paymentProof PaymentProof) GetCommitmentOutputSND() []*privacy.EllipticPoint {
	return paymentProof.CommitmentOutputSND
}

func (paymentProof PaymentProof) GetCommitmentOutputShardID() []*privacy.EllipticPoint {
	return paymentProof.CommitmentOutputShardID
}

func (paymentProof PaymentProof) GetCommitmentInputSecretKey() *privacy.EllipticPoint {
	return paymentProof.CommitmentInputSecretKey
}

func (paymentProof PaymentProof) GetCommitmentInputValue() []*privacy.EllipticPoint {
	return paymentProof.CommitmentInputValue
}

func (paymentProof PaymentProof) GetCommitmentInputSND() []*privacy.EllipticPoint {
	return paymentProof.CommitmentInputSND
}

func (paymentProof PaymentProof) GetCommitmentInputShardID() *privacy.EllipticPoint {
	return paymentProof.CommitmentInputShardID
}

func (paymentProof PaymentProof) GetCommitmentIndices() []uint64 {
	return paymentProof.CommitmentIndices
}

func (paymentProof PaymentProof) GetInputCoins() []*privacy.InputCoin {
	return paymentProof.InputCoins
}

func (paymentProof *PaymentProof) SetInputCoins(v []*privacy.InputCoin) {
	paymentProof.InputCoins = v
}

func (paymentProof PaymentProof) GetOutputCoins() []*privacy.OutputCoin {
	return paymentProof.OutputCoins
}

func (paymentProof *PaymentProof) SetOutputCoins(v []*privacy.OutputCoin) {
	paymentProof.OutputCoins = v
}

// End GET/SET function

// Init
func (proof *PaymentProof) Init() {
	aggregatedRangeProof := &aggregaterange.AggregatedRangeProof{}
	aggregatedRangeProof.Init()
	proof.OneOfManyProof = []*oneoutofmany.OneOutOfManyProof{}
	proof.SerialNumberProof = []*serialnumberprivacy.SNPrivacyProof{}
	proof.AggregatedRangeProof = aggregatedRangeProof
	proof.InputCoins = []*privacy.InputCoin{}
	proof.OutputCoins = []*privacy.OutputCoin{}

	proof.CommitmentOutputValue = []*privacy.EllipticPoint{}
	proof.CommitmentOutputSND = []*privacy.EllipticPoint{}
	proof.CommitmentOutputShardID = []*privacy.EllipticPoint{}

	proof.CommitmentInputSecretKey = new(privacy.EllipticPoint)
	proof.CommitmentInputValue = []*privacy.EllipticPoint{}
	proof.CommitmentInputSND = []*privacy.EllipticPoint{}
	proof.CommitmentInputShardID = new(privacy.EllipticPoint)

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
	hasPrivacy := len(proof.OneOfManyProof) > 0

	// OneOfManyProofSize
	bytes = append(bytes, byte(len(proof.OneOfManyProof)))
	for i := 0; i < len(proof.OneOfManyProof); i++ {
		oneOfManyProof := proof.OneOfManyProof[i].Bytes()
		bytes = append(bytes, common.IntToBytes(utils.OneOfManyProofSize)...)
		bytes = append(bytes, oneOfManyProof...)
	}

	// SerialNumberProofSize
	bytes = append(bytes, byte(len(proof.SerialNumberProof)))
	for i := 0; i < len(proof.SerialNumberProof); i++ {
		serialNumberProof := proof.SerialNumberProof[i].Bytes()
		bytes = append(bytes, common.IntToBytes(utils.SnPrivacyProofSize)...)
		bytes = append(bytes, serialNumberProof...)
	}

	// SNNoPrivacyProofSize
	bytes = append(bytes, byte(len(proof.SerialNumberNoPrivacyProof)))
	for i := 0; i < len(proof.SerialNumberNoPrivacyProof); i++ {
		snNoPrivacyProof := proof.SerialNumberNoPrivacyProof[i].Bytes()
		bytes = append(bytes, byte(utils.SnNoPrivacyProofSize))
		bytes = append(bytes, snNoPrivacyProof...)
	}

	//ComOutputMultiRangeProofSize
	if hasPrivacy {
		comOutputMultiRangeProof := proof.AggregatedRangeProof.Bytes()
		bytes = append(bytes, common.IntToBytes(len(comOutputMultiRangeProof))...)
		bytes = append(bytes, comOutputMultiRangeProof...)
	} else {
		bytes = append(bytes, []byte{0, 0}...)
	}

	// InputCoins
	bytes = append(bytes, byte(len(proof.InputCoins)))
	for i := 0; i < len(proof.InputCoins); i++ {
		inputCoins := proof.InputCoins[i].Bytes()
		bytes = append(bytes, byte(len(inputCoins)))
		bytes = append(bytes, inputCoins...)
	}

	// OutputCoins
	bytes = append(bytes, byte(len(proof.OutputCoins)))
	for i := 0; i < len(proof.OutputCoins); i++ {
		outputCoins := proof.OutputCoins[i].Bytes()
		lenOutputCoins := len(outputCoins)
		bytes = append(bytes, byte(lenOutputCoins))
		bytes = append(bytes, outputCoins...)
	}

	// ComOutputValue
	bytes = append(bytes, byte(len(proof.CommitmentOutputValue)))
	for i := 0; i < len(proof.CommitmentOutputValue); i++ {
		comOutputValue := proof.CommitmentOutputValue[i].Compress()
		bytes = append(bytes, byte(privacy.CompressedEllipticPointSize))
		bytes = append(bytes, comOutputValue...)
	}

	// ComOutputSND
	bytes = append(bytes, byte(len(proof.CommitmentOutputSND)))
	for i := 0; i < len(proof.CommitmentOutputSND); i++ {
		comOutputSND := proof.CommitmentOutputSND[i].Compress()
		bytes = append(bytes, byte(privacy.CompressedEllipticPointSize))
		bytes = append(bytes, comOutputSND...)
	}

	// ComOutputShardID
	bytes = append(bytes, byte(len(proof.CommitmentOutputShardID)))
	for i := 0; i < len(proof.CommitmentOutputShardID); i++ {
		comOutputShardID := proof.CommitmentOutputShardID[i].Compress()
		bytes = append(bytes, byte(privacy.CompressedEllipticPointSize))
		bytes = append(bytes, comOutputShardID...)
	}

	//ComInputSK 				*privacy.EllipticPoint
	if proof.CommitmentInputSecretKey != nil {
		comInputSK := proof.CommitmentInputSecretKey.Compress()
		bytes = append(bytes, byte(privacy.CompressedEllipticPointSize))
		bytes = append(bytes, comInputSK...)
	} else {
		bytes = append(bytes, byte(0))
	}

	//ComInputValue 		[]*privacy.EllipticPoint
	bytes = append(bytes, byte(len(proof.CommitmentInputValue)))
	for i := 0; i < len(proof.CommitmentInputValue); i++ {
		comInputValue := proof.CommitmentInputValue[i].Compress()
		bytes = append(bytes, byte(privacy.CompressedEllipticPointSize))
		bytes = append(bytes, comInputValue...)
	}

	//ComInputSND 			[]*privacy.EllipticPoint
	bytes = append(bytes, byte(len(proof.CommitmentInputSND)))
	for i := 0; i < len(proof.CommitmentInputSND); i++ {
		comInputSND := proof.CommitmentInputSND[i].Compress()
		bytes = append(bytes, byte(privacy.CompressedEllipticPointSize))
		bytes = append(bytes, comInputSND...)
	}

	//ComInputShardID 	*privacy.EllipticPoint
	if proof.CommitmentInputShardID != nil {
		comInputShardID := proof.CommitmentInputShardID.Compress()
		bytes = append(bytes, byte(privacy.CompressedEllipticPointSize))
		bytes = append(bytes, comInputShardID...)
	} else {
		bytes = append(bytes, byte(0))
	}

	// convert commitment index to bytes array
	for i := 0; i < len(proof.CommitmentIndices); i++ {
		bytes = append(bytes, common.AddPaddingBigInt(big.NewInt(int64(proof.CommitmentIndices[i])), common.Uint64Size)...)
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
	proof.OneOfManyProof = make([]*oneoutofmany.OneOutOfManyProof, lenOneOfManyProofArray)
	for i := 0; i < lenOneOfManyProofArray; i++ {
		lenOneOfManyProof := common.BytesToInt(proofbytes[offset : offset+2])
		offset += 2
		proof.OneOfManyProof[i] = new(oneoutofmany.OneOutOfManyProof).Init()
		err := proof.OneOfManyProof[i].SetBytes(proofbytes[offset : offset+lenOneOfManyProof])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenOneOfManyProof
	}

	// Set serialNumberProofSize
	lenSerialNumberProofArray := int(proofbytes[offset])
	offset += 1
	proof.SerialNumberProof = make([]*serialnumberprivacy.SNPrivacyProof, lenSerialNumberProofArray)
	for i := 0; i < lenSerialNumberProofArray; i++ {
		lenSerialNumberProof := common.BytesToInt(proofbytes[offset : offset+2])
		offset += 2
		proof.SerialNumberProof[i] = new(serialnumberprivacy.SNPrivacyProof).Init()
		err := proof.SerialNumberProof[i].SetBytes(proofbytes[offset : offset+lenSerialNumberProof])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenSerialNumberProof
	}

	// Set SNNoPrivacyProofSize
	lenSNNoPrivacyProofArray := int(proofbytes[offset])
	offset += 1
	proof.SerialNumberNoPrivacyProof = make([]*serialnumbernoprivacy.SNNoPrivacyProof, lenSNNoPrivacyProofArray)
	for i := 0; i < lenSNNoPrivacyProofArray; i++ {
		lenSNNoPrivacyProof := int(proofbytes[offset])
		offset += 1
		proof.SerialNumberNoPrivacyProof[i] = new(serialnumbernoprivacy.SNNoPrivacyProof).Init()
		err := proof.SerialNumberNoPrivacyProof[i].SetBytes(proofbytes[offset : offset+lenSNNoPrivacyProof])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenSNNoPrivacyProof
	}

	//ComOutputMultiRangeProofSize *AggregatedRangeProof
	lenComOutputMultiRangeProof := common.BytesToInt(proofbytes[offset : offset+2])
	offset += 2
	if lenComOutputMultiRangeProof > 0 {
		aggregatedRangeProof := &aggregaterange.AggregatedRangeProof{}
		aggregatedRangeProof.Init()
		proof.AggregatedRangeProof = aggregatedRangeProof
		err := proof.AggregatedRangeProof.SetBytes(proofbytes[offset : offset+lenComOutputMultiRangeProof])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComOutputMultiRangeProof
	}

	//InputCoins  []*privacy.InputCoin
	lenInputCoinsArray := int(proofbytes[offset])
	offset += 1
	proof.InputCoins = make([]*privacy.InputCoin, lenInputCoinsArray)
	for i := 0; i < lenInputCoinsArray; i++ {
		lenInputCoin := int(proofbytes[offset])
		offset += 1
		proof.InputCoins[i] = new(privacy.InputCoin)
		err := proof.InputCoins[i].SetBytes(proofbytes[offset : offset+lenInputCoin])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenInputCoin
	}

	//OutputCoins []*privacy.OutputCoin
	lenOutputCoinsArray := int(proofbytes[offset])
	offset += 1
	proof.OutputCoins = make([]*privacy.OutputCoin, lenOutputCoinsArray)
	for i := 0; i < lenOutputCoinsArray; i++ {
		lenOutputCoin := int(proofbytes[offset])
		offset += 1
		proof.OutputCoins[i] = new(privacy.OutputCoin)
		err := proof.OutputCoins[i].SetBytes(proofbytes[offset : offset+lenOutputCoin])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenOutputCoin
	}
	//ComOutputValue   []*privacy.EllipticPoint
	lenComOutputValueArray := int(proofbytes[offset])
	offset += 1
	proof.CommitmentOutputValue = make([]*privacy.EllipticPoint, lenComOutputValueArray)
	for i := 0; i < lenComOutputValueArray; i++ {
		lenComOutputValue := int(proofbytes[offset])
		offset += 1
		proof.CommitmentOutputValue[i] = new(privacy.EllipticPoint)
		err := proof.CommitmentOutputValue[i].Decompress(proofbytes[offset : offset+lenComOutputValue])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComOutputValue
	}
	//ComOutputSND     []*privacy.EllipticPoint
	lenComOutputSNDArray := int(proofbytes[offset])
	offset += 1
	proof.CommitmentOutputSND = make([]*privacy.EllipticPoint, lenComOutputSNDArray)
	for i := 0; i < lenComOutputSNDArray; i++ {
		lenComOutputSND := int(proofbytes[offset])
		offset += 1
		proof.CommitmentOutputSND[i] = new(privacy.EllipticPoint)
		err := proof.CommitmentOutputSND[i].Decompress(proofbytes[offset : offset+lenComOutputSND])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComOutputSND
	}

	lenComOutputShardIdArray := int(proofbytes[offset])
	offset += 1
	proof.CommitmentOutputShardID = make([]*privacy.EllipticPoint, lenComOutputShardIdArray)
	for i := 0; i < lenComOutputShardIdArray; i++ {
		lenComOutputShardId := int(proofbytes[offset])
		offset += 1
		proof.CommitmentOutputShardID[i] = new(privacy.EllipticPoint)
		err := proof.CommitmentOutputShardID[i].Decompress(proofbytes[offset : offset+lenComOutputShardId])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComOutputShardId
	}

	//ComInputSK 				*privacy.EllipticPoint
	lenComInputSK := int(proofbytes[offset])
	offset += 1
	if lenComInputSK > 0 {
		proof.CommitmentInputSecretKey = new(privacy.EllipticPoint)
		err := proof.CommitmentInputSecretKey.Decompress(proofbytes[offset : offset+lenComInputSK])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComInputSK
	}
	//ComInputValue 		[]*privacy.EllipticPoint
	lenComInputValueArr := int(proofbytes[offset])
	offset += 1
	proof.CommitmentInputValue = make([]*privacy.EllipticPoint, lenComInputValueArr)
	for i := 0; i < lenComInputValueArr; i++ {
		lenComInputValue := int(proofbytes[offset])
		offset += 1
		proof.CommitmentInputValue[i] = new(privacy.EllipticPoint)
		err := proof.CommitmentInputValue[i].Decompress(proofbytes[offset : offset+lenComInputValue])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComInputValue
	}
	//ComInputSND 			[]*privacy.EllipticPoint
	lenComInputSNDArr := int(proofbytes[offset])
	offset += 1
	proof.CommitmentInputSND = make([]*privacy.EllipticPoint, lenComInputSNDArr)
	for i := 0; i < lenComInputSNDArr; i++ {
		lenComInputSND := int(proofbytes[offset])
		offset += 1
		proof.CommitmentInputSND[i] = new(privacy.EllipticPoint)
		err := proof.CommitmentInputSND[i].Decompress(proofbytes[offset : offset+lenComInputSND])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComInputSND
	}
	//ComInputShardID 	*privacy.EllipticPoint
	lenComInputShardID := int(proofbytes[offset])
	offset += 1
	if lenComInputShardID > 0 {
		proof.CommitmentInputShardID = new(privacy.EllipticPoint)
		err := proof.CommitmentInputShardID.Decompress(proofbytes[offset : offset+lenComInputShardID])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComInputShardID
	}

	// get commitments list
	proof.CommitmentIndices = make([]uint64, len(proof.OneOfManyProof)*privacy.CommitmentRingSize)
	for i := 0; i < len(proof.OneOfManyProof)*privacy.CommitmentRingSize; i++ {
		proof.CommitmentIndices[i] = new(big.Int).SetBytes(proofbytes[offset : offset+common.Uint64Size]).Uint64()
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

	for i := 0; i < len(proof.InputCoins); i++ {
		// Check input coins' Serial number is created from input coins' input and sender's spending key
		valid, err := proof.SerialNumberNoPrivacyProof[i].Verify(nil)
		if !valid {
			privacy.Logger.Log.Errorf("Verify serial number no privacy proof failed")
			return false, privacy.NewPrivacyErr(privacy.VerifySerialNumberNoPrivacyProofFailedErr, err)
		}

		// Check input coins' cm is calculated correctly
		cmTmp := proof.InputCoins[i].CoinDetails.GetPublicKey()
		cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.PedersenValueIndex].ScalarMult(big.NewInt(int64(proof.InputCoins[i].CoinDetails.GetValue()))))
		cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.PedersenSndIndex].ScalarMult(proof.InputCoins[i].CoinDetails.GetSNDerivator()))
		cmTmp = cmTmp.Add(cmShardIDSender)
		cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.PedersenRandomnessIndex].ScalarMult(proof.InputCoins[i].CoinDetails.GetRandomness()))
		if !cmTmp.IsEqual(proof.InputCoins[i].CoinDetails.GetCoinCommitment()) {
			privacy.Logger.Log.Errorf("Input coins %v commitment wrong!\n", i)
			return false, privacy.NewPrivacyErr(privacy.VerifyCoinCommitmentInputFailedErr, nil)
		}

		// Calculate sum of input values
		sumInputValue += proof.InputCoins[i].CoinDetails.GetValue()
	}

	for i := 0; i < len(proof.OutputCoins); i++ {
		// Check output coins' cm is calculated correctly
		cmTmp := proof.OutputCoins[i].CoinDetails.GetPublicKey()
		cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.PedersenValueIndex].ScalarMult(big.NewInt(int64(proof.OutputCoins[i].CoinDetails.GetValue()))))
		cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.PedersenSndIndex].ScalarMult(proof.OutputCoins[i].CoinDetails.GetSNDerivator()))
		shardID := common.GetShardIDFromLastByte(proof.OutputCoins[i].CoinDetails.GetPubKeyLastByte())
		cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.PedersenShardIDIndex].ScalarMult(new(big.Int).SetBytes([]byte{shardID})))
		cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.PedersenRandomnessIndex].ScalarMult(proof.OutputCoins[i].CoinDetails.GetRandomness()))
		if !cmTmp.IsEqual(proof.OutputCoins[i].CoinDetails.GetCoinCommitment()) {
			privacy.Logger.Log.Errorf("Output coins %v commitment wrong!\n", i)
			return false, privacy.NewPrivacyErr(privacy.VerifyCoinCommitmentOutputFailedErr, nil)
		}

		// Calculate sum of output values
		sumOutputValue += proof.OutputCoins[i].CoinDetails.GetValue()
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
	cmInputSum := make([]*privacy.EllipticPoint, len(proof.OneOfManyProof))
	for i := 0; i < len(proof.OneOfManyProof); i++ {
		// Verify for the proof one-out-of-N commitments is a commitment to the coins being spent
		// Calculate cm input sum
		cmInputSum[i] = new(privacy.EllipticPoint)
		cmInputSum[i] = proof.CommitmentInputSecretKey.Add(proof.CommitmentInputValue[i])
		cmInputSum[i] = cmInputSum[i].Add(proof.CommitmentInputSND[i])
		cmInputSum[i] = cmInputSum[i].Add(proof.CommitmentInputShardID)

		// get commitments list from CommitmentIndices
		commitments := make([]*privacy.EllipticPoint, privacy.CommitmentRingSize)
		for j := 0; j < privacy.CommitmentRingSize; j++ {
			index := proof.CommitmentIndices[i*privacy.CommitmentRingSize+j]
			commitmentBytes, err := db.GetCommitmentByIndex(*tokenID, index, shardID)

			if err != nil {
				privacy.Logger.Log.Errorf("VERIFICATION PAYMENT PROOF: Error when get commitment by index from database", index, err)
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

		proof.OneOfManyProof[i].Statement.Commitments = commitments

		valid, err := proof.OneOfManyProof[i].Verify()
		if !valid {
			privacy.Logger.Log.Errorf("VERIFICATION PAYMENT PROOF: One out of many failed")
			return false, privacy.NewPrivacyErr(privacy.VerifyOneOutOfManyProofFailedErr, err)
		}
		// Verify for the Proof that input coins' serial number is derived from the committed derivator
		valid, err = proof.SerialNumberProof[i].Verify(nil)
		if !valid {
			privacy.Logger.Log.Errorf("VERIFICATION PAYMENT PROOF: Serial number privacy failed")
			return false, privacy.NewPrivacyErr(privacy.VerifySerialNumberPrivacyProofFailedErr, err)
		}
	}

	// Check output coins' cm is calculated correctly
	for i := 0; i < len(proof.OutputCoins); i++ {
		cmTmp := proof.OutputCoins[i].CoinDetails.GetPublicKey().Add(proof.CommitmentOutputValue[i])
		cmTmp = cmTmp.Add(proof.CommitmentOutputSND[i])
		cmTmp = cmTmp.Add(proof.CommitmentOutputShardID[i])

		if !cmTmp.IsEqual(proof.OutputCoins[i].CoinDetails.GetCoinCommitment()) {
			privacy.Logger.Log.Errorf("VERIFICATION PAYMENT PROOF: Commitment for output coins are not computed correctly")
			return false, privacy.NewPrivacyErr(privacy.VerifyCoinCommitmentOutputFailedErr, nil)
		}
	}

	// Verify the proof that output values and sum of them do not exceed v_max
	valid, err := proof.AggregatedRangeProof.Verify()
	if !valid {
		privacy.Logger.Log.Errorf("VERIFICATION PAYMENT PROOF: Multi-range failed")
		return false, privacy.NewPrivacyErr(privacy.VerifyAggregatedProofFailedErr, err)
	}

	// Verify the proof that sum of all input values is equal to sum of all output values
	comInputValueSum := new(privacy.EllipticPoint)
	comInputValueSum.Zero()
	for i := 0; i < len(proof.CommitmentInputValue); i++ {
		comInputValueSum = comInputValueSum.Add(proof.CommitmentInputValue[i])
	}

	comOutputValueSum := new(privacy.EllipticPoint)
	comOutputValueSum.Zero()
	for i := 0; i < len(proof.CommitmentOutputValue); i++ {
		comOutputValueSum = comOutputValueSum.Add(proof.CommitmentOutputValue[i])
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
