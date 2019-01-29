package zkp

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
)

// PaymentWitness contains all of witness for proving when spending coins
type PaymentWitness struct {
	spendingKey        *big.Int
	RandSK             *big.Int
	inputCoins         []*privacy.InputCoin
	outputCoins        []*privacy.OutputCoin
	commitmentIndexs   []uint64
	myCommitmentIndexs []uint64

	OneOfManyWitness    []*OneOutOfManyWitness
	SerialNumberWitness []*SNPrivacyWitness
	SNNoPrivacyWitness  []*SNNoPrivacyWitness

	ComOutputMultiRangeWitness *AggregatedRangeWitness

	ComOutputValue   []*privacy.EllipticPoint
	ComOutputSND     []*privacy.EllipticPoint
	ComOutputShardID []*privacy.EllipticPoint

	ComInputSK      *privacy.EllipticPoint
	ComInputValue   []*privacy.EllipticPoint
	ComInputSND     []*privacy.EllipticPoint
	ComInputShardID *privacy.EllipticPoint
}

// PaymentProof contains all of PoK for spending coin
type PaymentProof struct {
	// for input coins
	OneOfManyProof    []*OneOutOfManyProof
	SerialNumberProof []*SNPrivacyProof
	// it is exits when tx has no privacy
	SNNoPrivacyProof []*SNNoPrivacyProof

	// for output coins
	// for proving each value and sum of them are less than a threshold value
	ComOutputMultiRangeProof *AggregatedRangeProof

	InputCoins  []*privacy.InputCoin
	OutputCoins []*privacy.OutputCoin

	ComOutputValue   []*privacy.EllipticPoint
	ComOutputSND     []*privacy.EllipticPoint
	ComOutputShardID []*privacy.EllipticPoint

	ComInputSK      *privacy.EllipticPoint
	ComInputValue   []*privacy.EllipticPoint
	ComInputSND     []*privacy.EllipticPoint
	ComInputShardID *privacy.EllipticPoint


	CommitmentIndices []uint64
}

func (proof *PaymentProof) Init() *PaymentProof {
	proof = &PaymentProof{
		OneOfManyProof:           []*OneOutOfManyProof{},
		SerialNumberProof:        []*SNPrivacyProof{},
		ComOutputMultiRangeProof: new(AggregatedRangeProof).Init(),
		InputCoins:               []*privacy.InputCoin{},
		OutputCoins:              []*privacy.OutputCoin{},
		ComOutputValue:           []*privacy.EllipticPoint{},
		ComOutputSND:             []*privacy.EllipticPoint{},
		ComOutputShardID:         []*privacy.EllipticPoint{},
		ComInputSK:               new(privacy.EllipticPoint),
		ComInputValue:            []*privacy.EllipticPoint{},
		ComInputSND:              []*privacy.EllipticPoint{},
		ComInputShardID:          new(privacy.EllipticPoint),
	}
	return proof
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

	proof.SetBytes(temp)
	return nil
}

func (proof *PaymentProof) Bytes() []byte {
	var proofbytes []byte
	hasPrivacy := len(proof.OneOfManyProof) > 0
	// OneOfManyProofSize
	proofbytes = append(proofbytes, byte(len(proof.OneOfManyProof)))
	for i := 0; i < len(proof.OneOfManyProof); i++ {
		oneOfManyProof := proof.OneOfManyProof[i].Bytes()
		proofbytes = append(proofbytes, privacy.IntToByteArr(privacy.OneOfManyProofSize)...)
		proofbytes = append(proofbytes, oneOfManyProof...)
	}

	// SerialNumberProofSize
	proofbytes = append(proofbytes, byte(len(proof.SerialNumberProof)))
	for i := 0; i < len(proof.SerialNumberProof); i++ {
		serialNumberProof := proof.SerialNumberProof[i].Bytes()
		proofbytes = append(proofbytes, privacy.IntToByteArr(privacy.SNPrivacyProofSize)...)
		proofbytes = append(proofbytes, serialNumberProof...)
	}

	// SNNoPrivacyProofSize
	proofbytes = append(proofbytes, byte(len(proof.SNNoPrivacyProof)))
	for i := 0; i < len(proof.SNNoPrivacyProof); i++ {
		snNoPrivacyProof := proof.SNNoPrivacyProof[i].Bytes()
		proofbytes = append(proofbytes, byte(privacy.SNNoPrivacyProofSize))
		proofbytes = append(proofbytes, snNoPrivacyProof...)
	}

	//ComOutputMultiRangeProofSize
	if hasPrivacy {
		comOutputMultiRangeProof := proof.ComOutputMultiRangeProof.Bytes()
		proofbytes = append(proofbytes, privacy.IntToByteArr(len(comOutputMultiRangeProof))...)
		proofbytes = append(proofbytes, comOutputMultiRangeProof...)
	} else {
		proofbytes = append(proofbytes, []byte{0, 0}...)
	}

	// InputCoins
	proofbytes = append(proofbytes, byte(len(proof.InputCoins)))

	for i := 0; i < len(proof.InputCoins); i++ {
		inputCoins := proof.InputCoins[i].Bytes()
		proofbytes = append(proofbytes, byte(len(inputCoins)))
		proofbytes = append(proofbytes, inputCoins...)
	}
	// OutputCoins
	proofbytes = append(proofbytes, byte(len(proof.OutputCoins)))
	for i := 0; i < len(proof.OutputCoins); i++ {
		outputCoins := proof.OutputCoins[i].Bytes()
		lenOutputCoins := len(outputCoins)
		proofbytes = append(proofbytes, byte(lenOutputCoins))
		proofbytes = append(proofbytes, outputCoins...)
	}
	// ComOutputValue
	proofbytes = append(proofbytes, byte(len(proof.ComOutputValue)))
	for i := 0; i < len(proof.ComOutputValue); i++ {
		comOutputValue := proof.ComOutputValue[i].Compress()
		proofbytes = append(proofbytes, byte(privacy.CompressedPointSize))
		proofbytes = append(proofbytes, comOutputValue...)
	}
	// ComOutputSND
	proofbytes = append(proofbytes, byte(len(proof.ComOutputSND)))
	for i := 0; i < len(proof.ComOutputSND); i++ {
		comOutputSND := proof.ComOutputSND[i].Compress()
		proofbytes = append(proofbytes, byte(privacy.CompressedPointSize))
		proofbytes = append(proofbytes, comOutputSND...)
	}
	// ComOutputShardID
	proofbytes = append(proofbytes, byte(len(proof.ComOutputShardID)))
	for i := 0; i < len(proof.ComOutputShardID); i++ {
		comOutputShardID := proof.ComOutputShardID[i].Compress()
		proofbytes = append(proofbytes, byte(privacy.CompressedPointSize))
		proofbytes = append(proofbytes, comOutputShardID...)
	}

	//ComInputSK 				*privacy.EllipticPoint
	if proof.ComInputSK != nil {
		comInputSK := proof.ComInputSK.Compress()
		proofbytes = append(proofbytes, byte(privacy.CompressedPointSize))
		proofbytes = append(proofbytes, comInputSK...)
	} else {
		proofbytes = append(proofbytes, byte(0))
	}
	//ComInputValue 		[]*privacy.EllipticPoint
	proofbytes = append(proofbytes, byte(len(proof.ComInputValue)))
	for i := 0; i < len(proof.ComInputValue); i++ {
		fmt.Printf("comInputValue: %v\n", proof.ComInputValue[i])
		comInputValue := proof.ComInputValue[i].Compress()
		proofbytes = append(proofbytes, byte(privacy.CompressedPointSize))
		proofbytes = append(proofbytes, comInputValue...)
	}
	//ComInputSND 			[]*privacy.EllipticPoint
	proofbytes = append(proofbytes, byte(len(proof.ComInputSND)))
	for i := 0; i < len(proof.ComInputSND); i++ {
		comInputSND := proof.ComInputSND[i].Compress()
		proofbytes = append(proofbytes, byte(privacy.CompressedPointSize))
		proofbytes = append(proofbytes, comInputSND...)
	}
	//ComInputShardID 	*privacy.EllipticPoint
	if proof.ComInputShardID != nil {
		comInputShardID := proof.ComInputShardID.Compress()
		proofbytes = append(proofbytes, byte(privacy.CompressedPointSize))
		proofbytes = append(proofbytes, comInputShardID...)
	} else {
		proofbytes = append(proofbytes, byte(0))
	}

	// convert commitment index to bytes array
	for i := 0; i < len(proof.CommitmentIndices); i++ {
		commitmentIndexBytes := make([]byte, privacy.Uint64Size)
		binary.LittleEndian.PutUint64(commitmentIndexBytes, proof.CommitmentIndices[i])
		proofbytes = append(proofbytes, commitmentIndexBytes...)
	}

	//fmt.Printf("BYTES ------------------ %v\n", proofbytes)

	return proofbytes
}

func (proof *PaymentProof) SetBytes(proofbytes []byte) *privacy.PrivacyError {
	offset := 0

	// Set OneOfManyProofSize
	lenOneOfManyProofArray := int(proofbytes[offset])
	offset += 1
	proof.OneOfManyProof = make([]*OneOutOfManyProof, lenOneOfManyProofArray)
	for i := 0; i < lenOneOfManyProofArray; i++ {
		lenOneOfManyProof := privacy.ByteArrToInt(proofbytes[offset : offset+2])
		offset += 2
		proof.OneOfManyProof[i] = new(OneOutOfManyProof).Init()
		err := proof.OneOfManyProof[i].SetBytes(proofbytes[offset : offset+lenOneOfManyProof])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenOneOfManyProof
	}

	// Set serialNumberProofSize
	lenSerialNumberProofArray := int(proofbytes[offset])
	offset += 1
	proof.SerialNumberProof = make([]*SNPrivacyProof, lenSerialNumberProofArray)
	for i := 0; i < lenSerialNumberProofArray; i++ {
		lenSerialNumberProof := privacy.ByteArrToInt(proofbytes[offset : offset+2])
		offset += 2
		proof.SerialNumberProof[i] = new(SNPrivacyProof).Init()
		err := proof.SerialNumberProof[i].SetBytes(proofbytes[offset : offset+lenSerialNumberProof])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenSerialNumberProof
	}

	// Set SNNoPrivacyProofSize
	lenSNNoPrivacyProofArray := int(proofbytes[offset])
	offset += 1
	proof.SNNoPrivacyProof = make([]*SNNoPrivacyProof, lenSNNoPrivacyProofArray)
	for i := 0; i < lenSNNoPrivacyProofArray; i++ {
		lenSNNoPrivacyProof := int(proofbytes[offset])
		offset += 1
		proof.SNNoPrivacyProof[i] = new(SNNoPrivacyProof).Init()
		err := proof.SNNoPrivacyProof[i].SetBytes(proofbytes[offset : offset+lenSNNoPrivacyProof])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenSNNoPrivacyProof
	}

	//ComOutputMultiRangeProofSize *AggregatedRangeProof
	lenComOutputMultiRangeProof := privacy.ByteArrToInt(proofbytes[offset : offset+2])
	offset += 2
	if lenComOutputMultiRangeProof > 0 {
		proof.ComOutputMultiRangeProof = new(AggregatedRangeProof).Init()
		err := proof.ComOutputMultiRangeProof.SetBytes(proofbytes[offset : offset+lenComOutputMultiRangeProof])
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
	proof.ComOutputValue = make([]*privacy.EllipticPoint, lenComOutputValueArray)
	for i := 0; i < lenComOutputValueArray; i++ {
		lenComOutputValue := int(proofbytes[offset])
		offset += 1
		proof.ComOutputValue[i] = new(privacy.EllipticPoint)
		err := proof.ComOutputValue[i].Decompress(proofbytes[offset : offset+lenComOutputValue])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComOutputValue
	}
	//ComOutputSND     []*privacy.EllipticPoint
	lenComOutputSNDArray := int(proofbytes[offset])
	offset += 1
	proof.ComOutputSND = make([]*privacy.EllipticPoint, lenComOutputSNDArray)
	for i := 0; i < lenComOutputSNDArray; i++ {
		lenComOutputSND := int(proofbytes[offset])
		offset += 1
		proof.ComOutputSND[i] = new(privacy.EllipticPoint)
		err := proof.ComOutputSND[i].Decompress(proofbytes[offset : offset+lenComOutputSND])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComOutputSND
	}

	lenComOutputShardIdArray := int(proofbytes[offset])
	offset += 1
	proof.ComOutputShardID = make([]*privacy.EllipticPoint, lenComOutputShardIdArray)
	for i := 0; i < lenComOutputShardIdArray; i++ {
		lenComOutputShardId := int(proofbytes[offset])
		offset += 1
		proof.ComOutputShardID[i] = new(privacy.EllipticPoint)
		err := proof.ComOutputShardID[i].Decompress(proofbytes[offset : offset+lenComOutputShardId])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComOutputShardId
	}

	//ComInputSK 				*privacy.EllipticPoint
	lenComInputSK := int(proofbytes[offset])
	offset += 1
	if lenComInputSK > 0 {
		proof.ComInputSK = new(privacy.EllipticPoint)
		err := proof.ComInputSK.Decompress(proofbytes[offset : offset+lenComInputSK])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComInputSK
	}
	//ComInputValue 		[]*privacy.EllipticPoint
	lenComInputValueArr := int(proofbytes[offset])
	offset += 1
	proof.ComInputValue = make([]*privacy.EllipticPoint, lenComInputValueArr)
	for i := 0; i < lenComInputValueArr; i++ {
		lenComInputValue := int(proofbytes[offset])
		offset += 1
		proof.ComInputValue[i] = new(privacy.EllipticPoint)
		err := proof.ComInputValue[i].Decompress(proofbytes[offset : offset+lenComInputValue])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComInputValue
	}
	//ComInputSND 			[]*privacy.EllipticPoint
	lenComInputSNDArr := int(proofbytes[offset])
	offset += 1
	proof.ComInputSND = make([]*privacy.EllipticPoint, lenComInputSNDArr)
	for i := 0; i < lenComInputSNDArr; i++ {
		lenComInputSND := int(proofbytes[offset])
		offset += 1
		proof.ComInputSND[i] = new(privacy.EllipticPoint)
		err := proof.ComInputSND[i].Decompress(proofbytes[offset : offset+lenComInputSND])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComInputSND
	}
	//ComInputShardID 	*privacy.EllipticPoint
	lenComInputShardID := int(proofbytes[offset])
	offset += 1
	if lenComInputShardID > 0 {
		proof.ComInputShardID = new(privacy.EllipticPoint)
		err := proof.ComInputShardID.Decompress(proofbytes[offset : offset+lenComInputShardID])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComInputShardID
	}

	// get commitments list
	proof.CommitmentIndices = make([]uint64, len(proof.OneOfManyProof)* privacy.CMRingSize)
	for i := 0; i < len(proof.OneOfManyProof)* privacy.CMRingSize; i++ {
		proof.CommitmentIndices[i] = binary.LittleEndian.Uint64(proofbytes[offset : offset+privacy.Uint64Size])
		offset = offset + privacy.Uint64Size
	}

	//fmt.Printf("SETBYTES ------------------ %v\n", proof.Bytes())

	return nil
}

// Build prepares witnesses for all protocol need to be proved when create tx
// if hashPrivacy = false, witness includes spending key, input coins, output coins
// otherwise, witness includes all attributes in PaymentWitness struct
func (wit *PaymentWitness) Init(hasPrivacy bool,
	spendingKey *big.Int,
	inputCoins []*privacy.InputCoin, outputCoins []*privacy.OutputCoin,
	pkLastByteSender byte,
	commitments []*privacy.EllipticPoint, commitmentIndices []uint64, myCommitmentIndices []uint64,
	fee uint64) *privacy.PrivacyError {

	if !hasPrivacy {
		for _, outCoin := range outputCoins {
			outCoin.CoinDetails.Randomness = privacy.RandInt()
			outCoin.CoinDetails.CommitAll()
		}

		wit.spendingKey = spendingKey
		wit.inputCoins = inputCoins
		wit.outputCoins = outputCoins

		publicKey := inputCoins[0].CoinDetails.PublicKey

		wit.SNNoPrivacyWitness = make([]*SNNoPrivacyWitness, len(inputCoins))
		for i := 0; i < len(inputCoins); i++ {
			/***** Build witness for proving that serial number is derived from the committed derivator *****/
			if wit.SNNoPrivacyWitness[i] == nil {
				wit.SNNoPrivacyWitness[i] = new(SNNoPrivacyWitness)
			}
			wit.SNNoPrivacyWitness[i].Set(inputCoins[i].CoinDetails.SerialNumber, publicKey, inputCoins[i].CoinDetails.SNDerivator, wit.spendingKey)
		}
		return nil
	}

	wit.spendingKey = spendingKey
	wit.inputCoins = inputCoins
	wit.outputCoins = outputCoins
	wit.commitmentIndexs = commitmentIndices
	wit.myCommitmentIndexs = myCommitmentIndices

	numInputCoin := len(wit.inputCoins)

	randInputSK := privacy.RandInt()
	// set rand sk for Schnorr signature
	wit.RandSK = new(big.Int).Set(randInputSK)

	cmInputSK := privacy.PedCom.CommitAtIndex(wit.spendingKey, randInputSK, privacy.SK)
	wit.ComInputSK = new(privacy.EllipticPoint)
	wit.ComInputSK.Set(cmInputSK.X, cmInputSK.Y)

	randInputShardID := privacy.RandInt()
	wit.ComInputShardID = privacy.PedCom.CommitAtIndex(big.NewInt(int64(pkLastByteSender)), randInputShardID, privacy.SHARDID)

	wit.ComInputValue = make([]*privacy.EllipticPoint, numInputCoin)
	wit.ComInputSND = make([]*privacy.EllipticPoint, numInputCoin)
	// It is used for proving 2 commitments commit to the same value (input)
	cmInputSNDIndexSK := make([]*privacy.EllipticPoint, numInputCoin)

	randInputValue := make([]*big.Int, numInputCoin)
	randInputSND := make([]*big.Int, numInputCoin)
	randInputSNDIndexSK := make([]*big.Int, numInputCoin)

	// cmInputValueAll is sum of all input coins' value commitments
	cmInputValueAll := new(privacy.EllipticPoint).Zero()
	randInputValueAll := big.NewInt(0)

	// Summing all commitments of each input coin into one commitment and proving the knowledge of its Openings
	cmInputSum := make([]*privacy.EllipticPoint, numInputCoin)
	randInputSum := make([]*big.Int, numInputCoin)
	// randInputSumAll is sum of all randomess of coin commitments
	randInputSumAll := big.NewInt(0)

	wit.OneOfManyWitness = make([]*OneOutOfManyWitness, numInputCoin)
	wit.SerialNumberWitness = make([]*SNPrivacyWitness, numInputCoin)

	commitmentTemps := make([][]*privacy.EllipticPoint, numInputCoin)
	randInputIsZero := make([]*big.Int, numInputCoin)

	preIndex := 0

	for i, inputCoin := range wit.inputCoins {
		// commit each component of coin commitment
		randInputValue[i] = privacy.RandInt()
		randInputSND[i] = privacy.RandInt()
		randInputSNDIndexSK[i] = privacy.RandInt()

		wit.ComInputValue[i] = privacy.PedCom.CommitAtIndex(new(big.Int).SetUint64(inputCoin.CoinDetails.Value), randInputValue[i], privacy.VALUE)
		wit.ComInputSND[i] = privacy.PedCom.CommitAtIndex(inputCoin.CoinDetails.SNDerivator, randInputSND[i], privacy.SND)
		cmInputSNDIndexSK[i] = privacy.PedCom.CommitAtIndex(inputCoin.CoinDetails.SNDerivator, randInputSNDIndexSK[i], privacy.SK)

		cmInputValueAll = cmInputValueAll.Add(wit.ComInputValue[i])

		randInputValueAll.Add(randInputValueAll, randInputValue[i])
		randInputValueAll.Mod(randInputValueAll, privacy.Curve.Params().N)

		/***** Build witness for proving one-out-of-N commitments is a commitment to the coins being spent *****/
		cmInputSum[i] = cmInputSK.Add(wit.ComInputValue[i])
		cmInputSum[i] = cmInputSum[i].Add(wit.ComInputSND[i])
		cmInputSum[i] = cmInputSum[i].Add(wit.ComInputShardID)

		randInputSum[i] = new(big.Int).Set(randInputSK)
		randInputSum[i].Add(randInputSum[i], randInputValue[i])
		randInputSum[i].Add(randInputSum[i], randInputSND[i])
		randInputSum[i].Add(randInputSum[i], randInputShardID)
		randInputSum[i].Mod(randInputSum[i], privacy.Curve.Params().N)

		randInputSumAll.Add(randInputSumAll, randInputSum[i])
		randInputSumAll.Mod(randInputSumAll, privacy.Curve.Params().N)

		// commitmentTemps is a list of commitments for protocol one-out-of-N
		commitmentTemps[i] = make([]*privacy.EllipticPoint, privacy.CMRingSize)

		randInputIsZero[i] = big.NewInt(0)
		randInputIsZero[i].Sub(inputCoin.CoinDetails.Randomness, randInputSum[i])
		randInputIsZero[i].Mod(randInputIsZero[i], privacy.Curve.Params().N)

		for j := 0; j < privacy.CMRingSize; j++ {
			commitmentTemps[i][j] = new(privacy.EllipticPoint).Zero()
			commitmentTemps[i][j], _ = commitments[preIndex+j].Sub(cmInputSum[i])
		}

		if wit.OneOfManyWitness[i] == nil {
			wit.OneOfManyWitness[i] = new(OneOutOfManyWitness)
		}
		indexIsZero := myCommitmentIndices[i] % privacy.CMRingSize

		wit.OneOfManyWitness[i].Set(commitmentTemps[i], randInputIsZero[i], indexIsZero)
		preIndex = privacy.CMRingSize * (i + 1)
		// ---------------------------------------------------

		/***** Build witness for proving that serial number is derived from the committed derivator *****/
		if wit.SerialNumberWitness[i] == nil {
			wit.SerialNumberWitness[i] = new(SNPrivacyWitness)
		}
		stmt := new(SNPrivacyStatement)
		stmt.Set(inputCoin.CoinDetails.SerialNumber, cmInputSK, wit.ComInputSND[i])
		wit.SerialNumberWitness[i].Set(stmt, spendingKey, randInputSK, inputCoin.CoinDetails.SNDerivator, randInputSND[i])
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
			randOutputValue[i] = privacy.RandInt()
		}

		randOutputSND[i] = privacy.RandInt()
		randOutputShardID[i] = privacy.RandInt()

		cmOutputValue[i] = privacy.PedCom.CommitAtIndex(new(big.Int).SetUint64(outputCoin.CoinDetails.Value), randOutputValue[i], privacy.VALUE)
		cmOutputSND[i] = privacy.PedCom.CommitAtIndex(outputCoin.CoinDetails.SNDerivator, randOutputSND[i], privacy.SND)
		cmOutputShardID[i] = privacy.PedCom.CommitAtIndex(big.NewInt(int64(outputCoins[i].CoinDetails.GetPubKeyLastByte())), randOutputShardID[i], privacy.SHARDID)

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
	if wit.ComOutputMultiRangeWitness == nil {
		wit.ComOutputMultiRangeWitness = new(AggregatedRangeWitness)
	}
	wit.ComOutputMultiRangeWitness.Set(outputValue, randOutputValue)
	// ---------------------------------------------------

	// save partial commitments (value, input, shardID)
	wit.ComOutputValue = cmOutputValue
	wit.ComOutputSND = cmOutputSND
	wit.ComOutputShardID = cmOutputShardID
	return nil
}

// Prove creates big proof
func (wit *PaymentWitness) Prove(hasPrivacy bool) (*PaymentProof, *privacy.PrivacyError) {
	proof := new(PaymentProof).Init()
	var err error

	proof.InputCoins = wit.inputCoins
	proof.OutputCoins = wit.outputCoins
	proof.ComOutputValue = wit.ComOutputValue
	proof.ComOutputSND = wit.ComOutputSND
	proof.ComOutputShardID = wit.ComOutputShardID

	proof.ComInputSK = wit.ComInputSK
	proof.ComInputValue = wit.ComInputValue
	proof.ComInputSND = wit.ComInputSND
	proof.ComInputShardID = wit.ComInputShardID
	proof.CommitmentIndices = wit.commitmentIndexs

	// if hasPrivacy == false, don't need to create the zero knowledge proof
	// proving user has spending key corresponding with public key in input coins
	// is proved by signing with spending key
	if !hasPrivacy {
		// Proving that serial number is derived from the committed derivator
		for i := 0; i < len(wit.inputCoins); i++ {
			snNoPrivacyProof, err := wit.SNNoPrivacyWitness[i].Prove(nil)
			if err != nil {
				return nil, privacy.NewPrivacyErr(privacy.ProvingErr, err)
			}
			proof.SNNoPrivacyProof = append(proof.SNNoPrivacyProof, snNoPrivacyProof)
		}
		return proof, nil
	}

	// if hasPrivacy == true
	numInputCoins := len(wit.OneOfManyWitness)

	for i := 0; i < numInputCoins; i++ {
		// Proving one-out-of-N commitments is a commitment to the coins being spent
		oneOfManyProof, err := wit.OneOfManyWitness[i].Prove()
		if err != nil {
			return nil, privacy.NewPrivacyErr(privacy.ProvingErr, err)
		}
		proof.OneOfManyProof = append(proof.OneOfManyProof, oneOfManyProof)

		// Proving that serial number is derived from the committed derivator
		serialNumberProof, err := wit.SerialNumberWitness[i].Prove(nil)
		if err != nil {
			return nil, privacy.NewPrivacyErr(privacy.ProvingErr, err)
		}
		proof.SerialNumberProof = append(proof.SerialNumberProof, serialNumberProof)
	}

	// Proving that each output values and sum of them does not exceed v_max
	proof.ComOutputMultiRangeProof, err = wit.ComOutputMultiRangeWitness.Prove()
	if err != nil {
		return nil, privacy.NewPrivacyErr(privacy.ProvingErr, err)
	}

	privacy.Logger.Log.Info("Privacy log: PROVING DONE!!!")
	return proof, nil
}

func (proof PaymentProof) Verify(hasPrivacy bool, pubKey privacy.PublicKey, fee uint64, db database.DatabaseInterface, chainId byte, tokenID *common.Hash) bool {
	// has no privacy
	if !hasPrivacy {
		var sumInputValue, sumOutputValue uint64
		sumInputValue = 0
		sumOutputValue = 0

		for i := 0; i < len(proof.InputCoins); i++ {
			// Check input coins' Serial number is created from input coins' input and sender's spending key
			if !proof.SNNoPrivacyProof[i].Verify(nil) {
				return false
			}

			fmt.Printf("******************************************** Public key: %v\n", pubKey)
			pubKeyLastByteSender := pubKey[len(pubKey)-1]

			// Check input coins' cm is calculated correctly
			cmTmp := proof.InputCoins[i].CoinDetails.PublicKey
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.VALUE].ScalarMult(big.NewInt(int64(proof.InputCoins[i].CoinDetails.Value))))
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SND].ScalarMult(proof.InputCoins[i].CoinDetails.SNDerivator))
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SHARDID].ScalarMult(new(big.Int).SetBytes([]byte{pubKeyLastByteSender})))
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.RAND].ScalarMult(proof.InputCoins[i].CoinDetails.Randomness))
			if !cmTmp.IsEqual(proof.InputCoins[i].CoinDetails.CoinCommitment) {
				return false
			}

			// Calculate sum of input values
			sumInputValue += proof.InputCoins[i].CoinDetails.Value
		}

		for i := 0; i < len(proof.OutputCoins); i++ {
			// Check output coins' cm is calculated correctly
			cmTmp := proof.OutputCoins[i].CoinDetails.PublicKey
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.VALUE].ScalarMult(big.NewInt(int64(proof.OutputCoins[i].CoinDetails.Value))))
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SND].ScalarMult(proof.OutputCoins[i].CoinDetails.SNDerivator))
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SHARDID].ScalarMult(new(big.Int).SetBytes([]byte{proof.OutputCoins[i].CoinDetails.GetPubKeyLastByte()})))
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.RAND].ScalarMult(proof.OutputCoins[i].CoinDetails.Randomness))
			if !cmTmp.IsEqual(proof.OutputCoins[i].CoinDetails.CoinCommitment) {
				return false
			}

			// Calculate sum of output values
			sumOutputValue += proof.OutputCoins[i].CoinDetails.Value
		}

		// check if sum of input values equal sum of output values
		return sumInputValue == sumOutputValue+fee
	}

	// if hasPrivacy == true
	// verify for input coins
	cmInputSum := make([]*privacy.EllipticPoint, len(proof.OneOfManyProof))
	for i := 0; i < len(proof.OneOfManyProof); i++ {
		// Verify for the proof one-out-of-N commitments is a commitment to the coins being spent
		// Calculate cm input sum
		cmInputSum[i] = new(privacy.EllipticPoint)
		cmInputSum[i] = proof.ComInputSK.Add(proof.ComInputValue[i])
		cmInputSum[i] = cmInputSum[i].Add(proof.ComInputSND[i])
		cmInputSum[i] = cmInputSum[i].Add(proof.ComInputShardID)

		// get commitments list from CommitmentIndices
		commitments := make([]*privacy.EllipticPoint, privacy.CMRingSize)
		for j := 0; j < privacy.CMRingSize; j++ {
			commitmentBytes, err := db.GetCommitmentByIndex(tokenID, proof.CommitmentIndices[i*privacy.CMRingSize + j], chainId)
			if err != nil {
				fmt.Printf("err 1\n")
				privacy.NewPrivacyErr(privacy.VerificationErr, errors.New("zero knowledge verification error"))
				return false
			}
			commitments[j] = new(privacy.EllipticPoint)
			err = commitments[j].Decompress(commitmentBytes)
			if err != nil {
				fmt.Printf("err 2\n")
				privacy.NewPrivacyErr(privacy.VerificationErr, errors.New("zero knowledge verification error"))
				return false
			}

			commitments[j], err = commitments[j].Sub(cmInputSum[i])
			if err != nil {
				fmt.Printf("err 2\n")
				privacy.NewPrivacyErr(privacy.VerificationErr, errors.New("zero knowledge verification error"))
				return false
			}
		}

		proof.OneOfManyProof[i].stmt.commitments = commitments

		if !proof.OneOfManyProof[i].Verify() {
			fmt.Printf("err 3\n")
			return false
		}
		// Verify for the Proof that input coins' serial number is derived from the committed derivator
		if !proof.SerialNumberProof[i].Verify(nil) {
			fmt.Printf("err 4\n")
			return false
		}
	}

	// Check output coins' cm is calculated correctly
	for i := 0; i < len(proof.OutputCoins); i++ {
		cmTmp := proof.OutputCoins[i].CoinDetails.PublicKey.Add(proof.ComOutputValue[i])
		cmTmp = cmTmp.Add(proof.ComOutputSND[i])
		cmTmp = cmTmp.Add(proof.ComOutputShardID[i])

		if !cmTmp.IsEqual(proof.OutputCoins[i].CoinDetails.CoinCommitment) {
			fmt.Printf("err 5\n")
			return false
		}
	}

	// Verify the proof that output values and sum of them do not exceed v_max
	if !proof.ComOutputMultiRangeProof.Verify() {
		fmt.Printf("err 6\n")
		return false
	}

	// Verify the proof that sum of all input values is equal to sum of all output values
	comInputValueSum := new(privacy.EllipticPoint).Zero()
	for i := 0; i < len(proof.ComInputValue); i++ {
		comInputValueSum = comInputValueSum.Add(proof.ComInputValue[i])
	}

	comOutputValueSum := new(privacy.EllipticPoint).Zero()
	for i := 0; i < len(proof.ComOutputValue); i++ {
		comOutputValueSum = comOutputValueSum.Add(proof.ComOutputValue[i])
	}

	if fee > 0 {
		comOutputValueSum = comOutputValueSum.Add(privacy.PedCom.G[privacy.VALUE].ScalarMult(big.NewInt(int64(fee))))
	}

	return comInputValueSum.IsEqual(comOutputValueSum)
}
