package zkp

import (
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge/aggregaterange"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge/oneoutofmany"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge/serialnumbernoprivacy"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge/serialnumberprivacy"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge/utils"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/privacy"
)

// PaymentWitness contains all of witness for proving when spending coins
type PaymentWitness struct {
	privateKey         *big.Int
	inputCoins         []*privacy.InputCoin
	outputCoins        []*privacy.OutputCoin
	commitmentIndexs   []uint64
	myCommitmentIndexs []uint64

	oneOfManyWitness             []*oneoutofmany.OneOutOfManyWitness
	serialNumberWitness          []*serialnumberprivacy.SNPrivacyWitness
	serialNumberNoPrivacyWitness []*serialnumbernoprivacy.SNNoPrivacyWitness

	aggregatedRangeWitness *aggregaterange.AggregatedRangeWitness

	comOutputValue                 []*privacy.EllipticPoint
	comOutputSerialNumberDerivator []*privacy.EllipticPoint
	comOutputShardID               []*privacy.EllipticPoint

	comInputSecretKey             *privacy.EllipticPoint
	comInputValue                 []*privacy.EllipticPoint
	comInputSerialNumberDerivator []*privacy.EllipticPoint
	comInputShardID               *privacy.EllipticPoint

	randSecretKey *big.Int
}

func (paymentWitness PaymentWitness) GetRandSecretKey() *big.Int {
	return paymentWitness.randSecretKey
}

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

func (paymentProof PaymentProof) GetOutputCoins() []*privacy.OutputCoin {
	return paymentProof.outputCoins
}

func (paymentProof *PaymentProof) SetOutputCoins(v []*privacy.OutputCoin) {
	paymentProof.outputCoins = v
}

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

func (proof PaymentProof) MarshalJSON() ([]byte, error) {
	data := proof.Bytes()
	temp := base58.Base58Check{}.Encode(data, common.ZeroByte)
	return json.Marshal(temp)
}

func (proof *PaymentProof) UnmarshalJSON(data []byte) error {
	dataStr := ""
	_ = json.Unmarshal(data, &dataStr)
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
		bytes = append(bytes, privacy.IntToByteArr(utils.OneOfManyProofSize)...)
		bytes = append(bytes, oneOfManyProof...)
	}

	// SerialNumberProofSize
	bytes = append(bytes, byte(len(proof.serialNumberProof)))
	for i := 0; i < len(proof.serialNumberProof); i++ {
		serialNumberProof := proof.serialNumberProof[i].Bytes()
		bytes = append(bytes, privacy.IntToByteArr(utils.SnPrivacyProofSize)...)
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
		bytes = append(bytes, privacy.IntToByteArr(len(comOutputMultiRangeProof))...)
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
		bytes = append(bytes, privacy.AddPaddingBigInt(big.NewInt(int64(proof.commitmentIndices[i])), common.Uint64Size)...)
	}
	//fmt.Printf("BYTES ------------------ %v\n", bytes)
	//fmt.Printf("LEN BYTES ------------------ %v\n", len(bytes))

	return bytes
}

func (proof *PaymentProof) SetBytes(proofbytes []byte) *privacy.PrivacyError {
	offset := 0

	// Set OneOfManyProofSize
	lenOneOfManyProofArray := int(proofbytes[offset])
	offset += 1
	proof.oneOfManyProof = make([]*oneoutofmany.OneOutOfManyProof, lenOneOfManyProofArray)
	for i := 0; i < lenOneOfManyProofArray; i++ {
		lenOneOfManyProof := privacy.ByteArrToInt(proofbytes[offset : offset+2])
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
		lenSerialNumberProof := privacy.ByteArrToInt(proofbytes[offset : offset+2])
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
	lenComOutputMultiRangeProof := privacy.ByteArrToInt(proofbytes[offset : offset+2])
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

// Build prepares witnesses for all protocol need to be proved when create tx
// if hashPrivacy = false, witness includes spending key, input coins, output coins
// otherwise, witness includes all attributes in PaymentWitness struct
func (wit *PaymentWitness) Init(hasPrivacy bool,
	privateKey *big.Int,
	inputCoins []*privacy.InputCoin, outputCoins []*privacy.OutputCoin,
	pkLastByteSender byte,
	commitments []*privacy.EllipticPoint, commitmentIndices []uint64, myCommitmentIndices []uint64,
	fee uint64) *privacy.PrivacyError {

	if !hasPrivacy {
		for _, outCoin := range outputCoins {
			outCoin.CoinDetails.Randomness = privacy.RandScalar()
			outCoin.CoinDetails.CommitAll()
		}

		wit.privateKey = privateKey
		wit.inputCoins = inputCoins
		wit.outputCoins = outputCoins

		publicKey := inputCoins[0].CoinDetails.PublicKey

		wit.serialNumberNoPrivacyWitness = make([]*serialnumbernoprivacy.SNNoPrivacyWitness, len(inputCoins))
		for i := 0; i < len(inputCoins); i++ {
			/***** Build witness for proving that serial number is derived from the committed derivator *****/
			if wit.serialNumberNoPrivacyWitness[i] == nil {
				wit.serialNumberNoPrivacyWitness[i] = new(serialnumbernoprivacy.SNNoPrivacyWitness)
			}
			wit.serialNumberNoPrivacyWitness[i].Set(inputCoins[i].CoinDetails.SerialNumber, publicKey, inputCoins[i].CoinDetails.SNDerivator, wit.privateKey)
		}
		return nil
	}

	wit.privateKey = privateKey
	wit.inputCoins = inputCoins
	wit.outputCoins = outputCoins
	wit.commitmentIndexs = commitmentIndices
	wit.myCommitmentIndexs = myCommitmentIndices

	numInputCoin := len(wit.inputCoins)

	randInputSK := privacy.RandScalar()
	// set rand sk for Schnorr signature
	wit.randSecretKey = new(big.Int).Set(randInputSK)

	cmInputSK := privacy.PedCom.CommitAtIndex(wit.privateKey, randInputSK, privacy.SK)
	wit.comInputSecretKey = new(privacy.EllipticPoint)
	wit.comInputSecretKey.Set(cmInputSK.X, cmInputSK.Y)

	randInputShardID := privacy.RandScalar()
	senderShardID := common.GetShardIDFromLastByte(pkLastByteSender)
	wit.comInputShardID = privacy.PedCom.CommitAtIndex(big.NewInt(int64(senderShardID)), randInputShardID, privacy.SHARDID)

	wit.comInputValue = make([]*privacy.EllipticPoint, numInputCoin)
	wit.comInputSerialNumberDerivator = make([]*privacy.EllipticPoint, numInputCoin)
	// It is used for proving 2 commitments commit to the same value (input)
	//cmInputSNDIndexSK := make([]*privacy.EllipticPoint, numInputCoin)

	randInputValue := make([]*big.Int, numInputCoin)
	randInputSND := make([]*big.Int, numInputCoin)
	//randInputSNDIndexSK := make([]*big.Int, numInputCoin)

	// cmInputValueAll is sum of all input coins' value commitments
	cmInputValueAll := new(privacy.EllipticPoint).Zero()
	randInputValueAll := big.NewInt(0)

	// Summing all commitments of each input coin into one commitment and proving the knowledge of its Openings
	cmInputSum := make([]*privacy.EllipticPoint, numInputCoin)
	randInputSum := make([]*big.Int, numInputCoin)
	// randInputSumAll is sum of all randomess of coin commitments
	randInputSumAll := big.NewInt(0)

	wit.oneOfManyWitness = make([]*oneoutofmany.OneOutOfManyWitness, numInputCoin)
	wit.serialNumberWitness = make([]*serialnumberprivacy.SNPrivacyWitness, numInputCoin)

	commitmentTemps := make([][]*privacy.EllipticPoint, numInputCoin)
	randInputIsZero := make([]*big.Int, numInputCoin)

	preIndex := 0

	for i, inputCoin := range wit.inputCoins {
		// commit each component of coin commitment
		randInputValue[i] = privacy.RandScalar()
		randInputSND[i] = privacy.RandScalar()

		wit.comInputValue[i] = privacy.PedCom.CommitAtIndex(new(big.Int).SetUint64(inputCoin.CoinDetails.Value), randInputValue[i], privacy.VALUE)
		wit.comInputSerialNumberDerivator[i] = privacy.PedCom.CommitAtIndex(inputCoin.CoinDetails.SNDerivator, randInputSND[i], privacy.SND)

		cmInputValueAll = cmInputValueAll.Add(wit.comInputValue[i])

		randInputValueAll.Add(randInputValueAll, randInputValue[i])
		randInputValueAll.Mod(randInputValueAll, privacy.Curve.Params().N)

		/***** Build witness for proving one-out-of-N commitments is a commitment to the coins being spent *****/
		cmInputSum[i] = cmInputSK.Add(wit.comInputValue[i])
		cmInputSum[i] = cmInputSum[i].Add(wit.comInputSerialNumberDerivator[i])
		cmInputSum[i] = cmInputSum[i].Add(wit.comInputShardID)

		randInputSum[i] = new(big.Int).Set(randInputSK)
		randInputSum[i].Add(randInputSum[i], randInputValue[i])
		randInputSum[i].Add(randInputSum[i], randInputSND[i])
		randInputSum[i].Add(randInputSum[i], randInputShardID)
		randInputSum[i].Mod(randInputSum[i], privacy.Curve.Params().N)

		randInputSumAll.Add(randInputSumAll, randInputSum[i])
		randInputSumAll.Mod(randInputSumAll, privacy.Curve.Params().N)

		// commitmentTemps is a list of commitments for protocol one-out-of-N
		commitmentTemps[i] = make([]*privacy.EllipticPoint, privacy.CommitmentRingSize)

		randInputIsZero[i] = big.NewInt(0)
		randInputIsZero[i].Sub(inputCoin.CoinDetails.Randomness, randInputSum[i])
		randInputIsZero[i].Mod(randInputIsZero[i], privacy.Curve.Params().N)
		var err error
		for j := 0; j < privacy.CommitmentRingSize; j++ {
			commitmentTemps[i][j], err = commitments[preIndex+j].Sub(cmInputSum[i])
			if err != nil {
				return privacy.NewPrivacyErr(privacy.UnexpectedErr, err)

			}
		}

		if wit.oneOfManyWitness[i] == nil {
			wit.oneOfManyWitness[i] = new(oneoutofmany.OneOutOfManyWitness)
		}
		indexIsZero := myCommitmentIndices[i] % privacy.CommitmentRingSize

		wit.oneOfManyWitness[i].Set(commitmentTemps[i], randInputIsZero[i], indexIsZero)
		preIndex = privacy.CommitmentRingSize * (i + 1)
		// ---------------------------------------------------

		/***** Build witness for proving that serial number is derived from the committed derivator *****/
		if wit.serialNumberWitness[i] == nil {
			wit.serialNumberWitness[i] = new(serialnumberprivacy.SNPrivacyWitness)
		}
		stmt := new(serialnumberprivacy.SerialNumberPrivacyStatement)
		stmt.Set(inputCoin.CoinDetails.SerialNumber, cmInputSK, wit.comInputSerialNumberDerivator[i])
		wit.serialNumberWitness[i].Set(stmt, privateKey, randInputSK, inputCoin.CoinDetails.SNDerivator, randInputSND[i])
		// ---------------------------------------------------
	}

	numOutputCoin := len(wit.outputCoins)

	randOutputValue := make([]*big.Int, numOutputCoin)
	randOutputSND := make([]*big.Int, numOutputCoin)
	cmOutputValue := make([]*privacy.EllipticPoint, numOutputCoin)
	cmOutputSND := make([]*privacy.EllipticPoint, numOutputCoin)

	cmOutputSum := make([]*privacy.EllipticPoint, numOutputCoin)
	randOutputSum := make([]*big.Int, numOutputCoin)

	cmOutputSumAll := new(privacy.EllipticPoint).Zero()

	// cmOutputValueAll is sum of all value coin commitments
	cmOutputValueAll := new(privacy.EllipticPoint).Zero()
	randOutputValueAll := big.NewInt(0)

	randOutputShardID := make([]*big.Int, numOutputCoin)
	cmOutputShardID := make([]*privacy.EllipticPoint, numOutputCoin)

	for i, outputCoin := range wit.outputCoins {
		if i == len(outputCoins)-1 {
			randOutputValue[i] = new(big.Int).Sub(randInputValueAll, randOutputValueAll)
			randOutputValue[i].Mod(randOutputValue[i], privacy.Curve.Params().N)
		} else {
			randOutputValue[i] = privacy.RandScalar()
		}

		randOutputSND[i] = privacy.RandScalar()
		randOutputShardID[i] = privacy.RandScalar()

		cmOutputValue[i] = privacy.PedCom.CommitAtIndex(new(big.Int).SetUint64(outputCoin.CoinDetails.Value), randOutputValue[i], privacy.VALUE)
		cmOutputSND[i] = privacy.PedCom.CommitAtIndex(outputCoin.CoinDetails.SNDerivator, randOutputSND[i], privacy.SND)

		receiverShardID := common.GetShardIDFromLastByte(outputCoins[i].CoinDetails.GetPubKeyLastByte())
		cmOutputShardID[i] = privacy.PedCom.CommitAtIndex(big.NewInt(int64(receiverShardID)), randOutputShardID[i], privacy.SHARDID)

		randOutputSum[i] = big.NewInt(0)
		randOutputSum[i].Add(randOutputValue[i], randOutputSND[i])
		randOutputSum[i].Add(randOutputSum[i], randOutputShardID[i])
		randOutputSum[i].Mod(randOutputSum[i], privacy.Curve.Params().N)

		cmOutputSum[i] = new(privacy.EllipticPoint).Zero()
		cmOutputSum[i] = cmOutputValue[i].Add(cmOutputSND[i])
		cmOutputSum[i] = cmOutputSum[i].Add(outputCoins[i].CoinDetails.PublicKey)
		cmOutputSum[i] = cmOutputSum[i].Add(cmOutputShardID[i])

		cmOutputValueAll = cmOutputValueAll.Add(cmOutputValue[i])
		randOutputValueAll.Add(randOutputValueAll, randOutputValue[i])
		randOutputValueAll.Mod(randOutputValueAll, privacy.Curve.Params().N)

		// calculate final commitment for output coins
		outputCoins[i].CoinDetails.CoinCommitment = cmOutputSum[i]
		outputCoins[i].CoinDetails.Randomness = randOutputSum[i]

		cmOutputSumAll = cmOutputSumAll.Add(cmOutputSum[i])
	}

	// For Multi Range Protocol
	// proving each output value is less than vmax
	// proving sum of output values is less than vmax
	outputValue := make([]*big.Int, numOutputCoin)
	for i := 0; i < numOutputCoin; i++ {
		if outputCoins[i].CoinDetails.Value > 0 {
			outputValue[i] = big.NewInt(int64(outputCoins[i].CoinDetails.Value))
		} else {
			return privacy.NewPrivacyErr(privacy.UnexpectedErr, errors.New("output coin's value is less than 0"))
		}
	}
	if wit.aggregatedRangeWitness == nil {
		wit.aggregatedRangeWitness = new(aggregaterange.AggregatedRangeWitness)
	}
	wit.aggregatedRangeWitness.Set(outputValue, randOutputValue)
	// ---------------------------------------------------

	// save partial commitments (value, input, shardID)
	wit.comOutputValue = cmOutputValue
	wit.comOutputSerialNumberDerivator = cmOutputSND
	wit.comOutputShardID = cmOutputShardID
	return nil
}

// Prove creates big proof
func (wit *PaymentWitness) Prove(hasPrivacy bool) (*PaymentProof, *privacy.PrivacyError) {
	proof := new(PaymentProof)
	proof.Init()

	proof.inputCoins = wit.inputCoins
	proof.outputCoins = wit.outputCoins
	proof.commitmentOutputValue = wit.comOutputValue
	proof.commitmentOutputSND = wit.comOutputSerialNumberDerivator
	proof.commitmentOutputShardID = wit.comOutputShardID

	proof.commitmentInputSecretKey = wit.comInputSecretKey
	proof.commitmentInputValue = wit.comInputValue
	proof.commitmentInputSND = wit.comInputSerialNumberDerivator
	proof.commitmentInputShardID = wit.comInputShardID
	proof.commitmentIndices = wit.commitmentIndexs

	// if hasPrivacy == false, don't need to create the zero knowledge proof
	// proving user has spending key corresponding with public key in input coins
	// is proved by signing with spending key
	if !hasPrivacy {
		// Proving that serial number is derived from the committed derivator
		for i := 0; i < len(wit.inputCoins); i++ {
			snNoPrivacyProof, err := wit.serialNumberNoPrivacyWitness[i].Prove(nil)
			if err != nil {
				return nil, privacy.NewPrivacyErr(privacy.ProvingErr, err)
			}
			proof.serialNumberNoPrivacyProof = append(proof.serialNumberNoPrivacyProof, snNoPrivacyProof)
		}
		return proof, nil
	}

	// if hasPrivacy == true
	numInputCoins := len(wit.oneOfManyWitness)

	for i := 0; i < numInputCoins; i++ {
		// Proving one-out-of-N commitments is a commitment to the coins being spent
		oneOfManyProof, err := wit.oneOfManyWitness[i].Prove()
		if err != nil {
			return nil, privacy.NewPrivacyErr(privacy.ProvingErr, err)
		}
		proof.oneOfManyProof = append(proof.oneOfManyProof, oneOfManyProof)

		// Proving that serial number is derived from the committed derivator
		serialNumberProof, err := wit.serialNumberWitness[i].Prove(nil)
		if err != nil {
			return nil, privacy.NewPrivacyErr(privacy.ProvingErr, err)
		}
		proof.serialNumberProof = append(proof.serialNumberProof, serialNumberProof)
	}
	var err error

	// Proving that each output values and sum of them does not exceed v_max
	proof.aggregatedRangeProof, err = wit.aggregatedRangeWitness.Prove()
	if err != nil {
		return nil, privacy.NewPrivacyErr(privacy.ProvingErr, err)
	}

	privacy.Logger.Log.Info("Privacy log: PROVING DONE!!!")
	return proof, nil
}

func (proof PaymentProof) verifyNoPrivacy(pubKey privacy.PublicKey, fee uint64, db database.DatabaseInterface, shardID byte, tokenID *common.Hash) bool {
	var sumInputValue, sumOutputValue uint64
	sumInputValue = 0
	sumOutputValue = 0

	pubKeyLastByteSender := pubKey[len(pubKey)-1]
	senderShardID := common.GetShardIDFromLastByte(pubKeyLastByteSender)
	cmShardIDSender := privacy.PedCom.G[privacy.SHARDID].ScalarMult(new(big.Int).SetBytes([]byte{senderShardID}))

	for i := 0; i < len(proof.inputCoins); i++ {
		// Check input coins' Serial number is created from input coins' input and sender's spending key
		if !proof.serialNumberNoPrivacyProof[i].Verify(nil) {
			privacy.Logger.Log.Errorf("Failed verify serial number no privacy\n")
			return false
		}

		// Check input coins' cm is calculated correctly
		cmTmp := proof.inputCoins[i].CoinDetails.PublicKey
		cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.VALUE].ScalarMult(big.NewInt(int64(proof.inputCoins[i].CoinDetails.Value))))
		cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SND].ScalarMult(proof.inputCoins[i].CoinDetails.SNDerivator))
		cmTmp = cmTmp.Add(cmShardIDSender)
		cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.RAND].ScalarMult(proof.inputCoins[i].CoinDetails.Randomness))
		if !cmTmp.IsEqual(proof.inputCoins[i].CoinDetails.CoinCommitment) {
			privacy.Logger.Log.Errorf("Input coins %v commitment wrong!\n", i)
			return false
		}

		// Calculate sum of input values
		sumInputValue += proof.inputCoins[i].CoinDetails.Value
	}

	for i := 0; i < len(proof.outputCoins); i++ {
		// Check output coins' cm is calculated correctly
		cmTmp := proof.outputCoins[i].CoinDetails.PublicKey
		cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.VALUE].ScalarMult(big.NewInt(int64(proof.outputCoins[i].CoinDetails.Value))))
		cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SND].ScalarMult(proof.outputCoins[i].CoinDetails.SNDerivator))
		shardID := common.GetShardIDFromLastByte(proof.outputCoins[i].CoinDetails.GetPubKeyLastByte())
		cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SHARDID].ScalarMult(new(big.Int).SetBytes([]byte{shardID})))
		cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.RAND].ScalarMult(proof.outputCoins[i].CoinDetails.Randomness))
		if !cmTmp.IsEqual(proof.outputCoins[i].CoinDetails.CoinCommitment) {
			privacy.Logger.Log.Errorf("Output coins %v commitment wrong!\n", i)
			return false
		}

		// Calculate sum of output values
		sumOutputValue += proof.outputCoins[i].CoinDetails.Value
	}

	// check if sum of input values equal sum of output values
	if sumInputValue != sumOutputValue+fee {
		privacy.Logger.Log.Infof("sumInputValue: %v\n", sumInputValue)
		privacy.Logger.Log.Infof("sumOutputValue: %v\n", sumOutputValue)
		privacy.Logger.Log.Infof("fee: %v\n", fee)
		privacy.Logger.Log.Errorf("Sum of inputs is not equal sum of output!\n")
		return false
	}
	return true
}

func (proof PaymentProof) verifyHasPrivacy(pubKey privacy.PublicKey, fee uint64, db database.DatabaseInterface, shardID byte, tokenID *common.Hash) bool {
	// verify for input coins
	cmInputSum := make([]*privacy.EllipticPoint, len(proof.oneOfManyProof))
	for i := 0; i < len(proof.oneOfManyProof); i++ {
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

			if err != nil {
				privacy.Logger.Log.Error("VERIFICATION PAYMENT PROOF: Error when get commitment by index from database", index, err)
				privacy.NewPrivacyErr(privacy.VerificationErr, errors.New("zero knowledge verification error"))
				return false
			}
			commitments[j] = new(privacy.EllipticPoint)
			err = commitments[j].Decompress(commitmentBytes)
			if err != nil {
				privacy.Logger.Log.Error("VERIFICATION PAYMENT PROOF: Cannot decompress commitment from database", index, err)
				privacy.NewPrivacyErr(privacy.VerificationErr, errors.New("zero knowledge verification error"))
				return false
			}

			commitments[j], err = commitments[j].Sub(cmInputSum[i])
			if err != nil {
				privacy.Logger.Log.Error("VERIFICATION PAYMENT PROOF: Cannot sub commitment to sum of commitment inputs", index, err)
				privacy.NewPrivacyErr(privacy.VerificationErr, errors.New("zero knowledge verification error"))
				return false
			}
		}

		proof.oneOfManyProof[i].Statement.Commitments = commitments

		if !proof.oneOfManyProof[i].Verify() {
			privacy.Logger.Log.Error("VERIFICATION PAYMENT PROOF: One out of many failed")
			return false
		}
		// Verify for the Proof that input coins' serial number is derived from the committed derivator
		if !proof.serialNumberProof[i].Verify(nil) {
			privacy.Logger.Log.Error("VERIFICATION PAYMENT PROOF: Serial number privacy failed")
			return false
		}
	}

	// Check output coins' cm is calculated correctly
	for i := 0; i < len(proof.outputCoins); i++ {
		cmTmp := proof.outputCoins[i].CoinDetails.PublicKey.Add(proof.commitmentOutputValue[i])
		cmTmp = cmTmp.Add(proof.commitmentOutputSND[i])
		cmTmp = cmTmp.Add(proof.commitmentOutputShardID[i])

		if !cmTmp.IsEqual(proof.outputCoins[i].CoinDetails.CoinCommitment) {
			privacy.Logger.Log.Error("VERIFICATION PAYMENT PROOF: Commitment for output coins are not computed correctly")
			return false
		}
	}

	// Verify the proof that output values and sum of them do not exceed v_max
	if !proof.aggregatedRangeProof.Verify() {
		privacy.Logger.Log.Error("VERIFICATION PAYMENT PROOF: Multi-range failed")
		return false
	}

	// Verify the proof that sum of all input values is equal to sum of all output values
	comInputValueSum := new(privacy.EllipticPoint).Zero()
	for i := 0; i < len(proof.commitmentInputValue); i++ {
		comInputValueSum = comInputValueSum.Add(proof.commitmentInputValue[i])
	}

	comOutputValueSum := new(privacy.EllipticPoint).Zero()
	for i := 0; i < len(proof.commitmentOutputValue); i++ {
		comOutputValueSum = comOutputValueSum.Add(proof.commitmentOutputValue[i])
	}

	if fee > 0 {
		comOutputValueSum = comOutputValueSum.Add(privacy.PedCom.G[privacy.VALUE].ScalarMult(big.NewInt(int64(fee))))
	}

	privacy.Logger.Log.Debugf("comInputValueSum: ", comInputValueSum)
	privacy.Logger.Log.Debugf("comOutputValueSum: ", comOutputValueSum)

	if !comInputValueSum.IsEqual(comOutputValueSum) {
		privacy.Logger.Log.Debugf("comInputValueSum: ", comInputValueSum)
		privacy.Logger.Log.Debugf("comOutputValueSum: ", comOutputValueSum)
		privacy.Logger.Log.Error("VERIFICATION PAYMENT PROOF: Sum of input coins' value is not equal to sum of output coins' value")
		return false
	}

	return true
}

func (proof PaymentProof) Verify(hasPrivacy bool, pubKey privacy.PublicKey, fee uint64, db database.DatabaseInterface, shardID byte, tokenID *common.Hash) bool {
	// has no privacy
	if !hasPrivacy {
		return proof.verifyNoPrivacy(pubKey, fee, db, shardID, tokenID)
	}

	return proof.verifyHasPrivacy(pubKey, fee, db, shardID, tokenID)
}
