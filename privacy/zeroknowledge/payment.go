package zkp

import (
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
	SerialNumberWitness []*PKSNPrivacyWitness
	SNNoPrivacyWitness  []*SNNoPrivacyWitness

	ComOutputMultiRangeWitness *MultiRangeWitness
	ComZeroWitness             *ComZeroWitness

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
	SerialNumberProof []*PKSNPrivacyProof
	// it is exits when tx has no privacy
	SNNoPrivacyProof []*SNNoPrivacyProof

	// for output coins
	// for proving each value and sum of them are less than a threshold value
	ComOutputMultiRangeProof *MultiRangeProof
	// for input = output
	ComZeroProof *ComZeroProof

	InputCoins  []*privacy.InputCoin
	OutputCoins []*privacy.OutputCoin

	ComOutputValue   []*privacy.EllipticPoint
	ComOutputSND     []*privacy.EllipticPoint
	ComOutputShardID []*privacy.EllipticPoint

	ComInputSK      *privacy.EllipticPoint
	ComInputValue   []*privacy.EllipticPoint
	ComInputSND     []*privacy.EllipticPoint
	ComInputShardID *privacy.EllipticPoint
}

func (proof *PaymentProof) Init() *PaymentProof {
	proof = &PaymentProof{
		OneOfManyProof:           []*OneOutOfManyProof{},
		SerialNumberProof:        []*PKSNPrivacyProof{},
		ComOutputMultiRangeProof: new(MultiRangeProof).Init(),
		ComZeroProof:             new(ComZeroProof).Init(),
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

func (paymentProof *PaymentProof) Bytes() []byte {
	var proofbytes []byte
	hasPrivacy := len(paymentProof.OneOfManyProof) > 0
	// OneOfManyProofSize
	proofbytes = append(proofbytes, byte(len(paymentProof.OneOfManyProof)))
	for i := 0; i < len(paymentProof.OneOfManyProof); i++ {
		oneOfManyProof := paymentProof.OneOfManyProof[i].Bytes()
		proofbytes = append(proofbytes, privacy.IntToByteArr(privacy.OneOfManyProofSize)...)
		proofbytes = append(proofbytes, oneOfManyProof...)
	}

	// SerialNumberProofSize
	proofbytes = append(proofbytes, byte(len(paymentProof.SerialNumberProof)))
	for i := 0; i < len(paymentProof.SerialNumberProof); i++ {
		serialNumberProof := paymentProof.SerialNumberProof[i].Bytes()
		proofbytes = append(proofbytes, privacy.IntToByteArr(privacy.SNPrivacyProofSize)...)
		proofbytes = append(proofbytes, serialNumberProof...)
	}

	// SNNoPrivacyProofSize
	proofbytes = append(proofbytes, byte(len(paymentProof.SNNoPrivacyProof)))
	for i := 0; i < len(paymentProof.SNNoPrivacyProof); i++ {
		snNoPrivacyProof := paymentProof.SNNoPrivacyProof[i].Bytes()
		proofbytes = append(proofbytes, byte(privacy.SNNoPrivacyProofSize))
		proofbytes = append(proofbytes, snNoPrivacyProof...)
	}

	// ComOutputMultiRangeProofSize
	if hasPrivacy {
		comOutputMultiRangeProof := paymentProof.ComOutputMultiRangeProof.Bytes()
		proofbytes = append(proofbytes, privacy.IntToByteArr(len(comOutputMultiRangeProof))...)
		proofbytes = append(proofbytes, comOutputMultiRangeProof...)
	} else {
		proofbytes = append(proofbytes, []byte{0, 0}...)
	}

	if hasPrivacy {
		comZeroProof := paymentProof.ComZeroProof.Bytes()
		proofbytes = append(proofbytes, byte(len(comZeroProof)))
		proofbytes = append(proofbytes, comZeroProof...)
	} else {
		proofbytes = append(proofbytes, byte(0))
	}

	// InputCoins
	proofbytes = append(proofbytes, byte(len(paymentProof.InputCoins)))

	for i := 0; i < len(paymentProof.InputCoins); i++ {
		inputCoins := paymentProof.InputCoins[i].Bytes()
		proofbytes = append(proofbytes, byte(len(inputCoins)))
		proofbytes = append(proofbytes, inputCoins...)
	}
	// OutputCoins
	proofbytes = append(proofbytes, byte(len(paymentProof.OutputCoins)))
	for i := 0; i < len(paymentProof.OutputCoins); i++ {
		outputCoins := paymentProof.OutputCoins[i].Bytes()
		lenOutputCoins := len(outputCoins)
		proofbytes = append(proofbytes, byte(lenOutputCoins))
		proofbytes = append(proofbytes, outputCoins...)
	}
	// ComOutputValue
	proofbytes = append(proofbytes, byte(len(paymentProof.ComOutputValue)))
	for i := 0; i < len(paymentProof.ComOutputValue); i++ {
		comOutputValue := paymentProof.ComOutputValue[i].Compress()
		proofbytes = append(proofbytes, byte(privacy.CompressedPointSize))
		proofbytes = append(proofbytes, comOutputValue...)
	}
	// ComOutputSND
	proofbytes = append(proofbytes, byte(len(paymentProof.ComOutputSND)))
	for i := 0; i < len(paymentProof.ComOutputSND); i++ {
		comOutputSND := paymentProof.ComOutputSND[i].Compress()
		proofbytes = append(proofbytes, byte(privacy.CompressedPointSize))
		proofbytes = append(proofbytes, comOutputSND...)
	}
	// ComOutputShardID
	proofbytes = append(proofbytes, byte(len(paymentProof.ComOutputShardID)))
	for i := 0; i < len(paymentProof.ComOutputShardID); i++ {
		comOutputShardID := paymentProof.ComOutputShardID[i].Compress()
		proofbytes = append(proofbytes, byte(privacy.CompressedPointSize))
		proofbytes = append(proofbytes, comOutputShardID...)
	}

	//ComInputSK 				*privacy.EllipticPoint
	if paymentProof.ComInputSK != nil {
		comInputSK := paymentProof.ComInputSK.Compress()
		proofbytes = append(proofbytes, byte(privacy.CompressedPointSize))
		proofbytes = append(proofbytes, comInputSK...)
	} else {
		proofbytes = append(proofbytes, byte(0))
	}
	//ComInputValue 		[]*privacy.EllipticPoint
	proofbytes = append(proofbytes, byte(len(paymentProof.ComInputValue)))
	for i := 0; i < len(paymentProof.ComInputValue); i++ {
		comInputValue := paymentProof.ComInputValue[i].Compress()
		proofbytes = append(proofbytes, byte(privacy.CompressedPointSize))
		proofbytes = append(proofbytes, comInputValue...)
	}
	//ComInputSND 			[]*privacy.EllipticPoint
	proofbytes = append(proofbytes, byte(len(paymentProof.ComInputSND)))
	for i := 0; i < len(paymentProof.ComInputSND); i++ {
		comInputSND := paymentProof.ComInputSND[i].Compress()
		proofbytes = append(proofbytes, byte(privacy.CompressedPointSize))
		proofbytes = append(proofbytes, comInputSND...)
	}
	//ComInputShardID 	*privacy.EllipticPoint
	if paymentProof.ComInputShardID != nil {
		comInputShardID := paymentProof.ComInputShardID.Compress()
		proofbytes = append(proofbytes, byte(privacy.CompressedPointSize))
		proofbytes = append(proofbytes, comInputShardID...)
	} else {
		proofbytes = append(proofbytes, byte(0))
	}

	fmt.Printf("BYTES ------------------ %v\n", proofbytes)

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
	proof.SerialNumberProof = make([]*PKSNPrivacyProof, lenSerialNumberProofArray)
	for i := 0; i < lenSerialNumberProofArray; i++ {
		lenSerialNumberProof := privacy.ByteArrToInt(proofbytes[offset : offset+2])
		offset += 2
		proof.SerialNumberProof[i] = new(PKSNPrivacyProof).Init()
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

	//ComOutputMultiRangeProofSize *MultiRangeProof
	lenComOutputMultiRangeProof := privacy.ByteArrToInt(proofbytes[offset : offset+2])
	offset += 2
	if lenComOutputMultiRangeProof > 0 {
		proof.ComOutputMultiRangeProof = new(MultiRangeProof).Init()
		err := proof.ComOutputMultiRangeProof.SetBytes(proofbytes[offset : offset+lenComOutputMultiRangeProof])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComOutputMultiRangeProof
	}

	//ComZeroProof *ComZeroProof
	lenComZeroProof := int(proofbytes[offset])
	offset += 1
	if lenComZeroProof > 0 {
		proof.ComZeroProof = new(ComZeroProof).Init()
		err := proof.ComZeroProof.SetBytes(proofbytes[offset : offset+lenComZeroProof])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComZeroProof
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
	var err error
	lenComOutputValueArray := int(proofbytes[offset])
	offset += 1
	proof.ComOutputValue = make([]*privacy.EllipticPoint, lenComOutputValueArray)
	for i := 0; i < lenComOutputValueArray; i++ {
		lenComOutputValue := int(proofbytes[offset])
		offset += 1
		proof.ComOutputValue[i] = new(privacy.EllipticPoint)
		proof.ComOutputValue[i], err = privacy.DecompressKey(proofbytes[offset : offset+lenComOutputValue])
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
		proof.ComOutputSND[i], err = privacy.DecompressKey(proofbytes[offset : offset+lenComOutputSND])
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
		proof.ComOutputShardID[i], err = privacy.DecompressKey(proofbytes[offset : offset+lenComOutputShardId])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComOutputShardId
	}

	//ComInputSK 				*privacy.EllipticPoint
	lenComInputSK := int(proofbytes[offset])
	offset += 1
	if lenComZeroProof > 0 {
		proof.ComInputSK = new(privacy.EllipticPoint)
		proof.ComInputSK, err = privacy.DecompressKey(proofbytes[offset : offset+lenComInputSK])
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
		proof.ComInputValue[i], err = privacy.DecompressKey(proofbytes[offset : offset+lenComInputValue])
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
		proof.ComInputSND[i], err = privacy.DecompressKey(proofbytes[offset : offset+lenComInputSND])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComInputSND
	}
	//ComInputShardID 	*privacy.EllipticPoint
	lenComInputShardID := int(proofbytes[offset])
	offset += 1
	if lenComZeroProof > 0 {
		proof.ComInputShardID = new(privacy.EllipticPoint)
		proof.ComInputShardID, err = privacy.DecompressKey(proofbytes[offset : offset+lenComInputShardID])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComInputShardID
	}

	fmt.Printf("SETBYTES ------------------ %v\n", proof.Bytes())

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
		//wit.commitmentIndexs = commitmentIndices
		//wit.myCommitmentIndexs = myCommitmentIndices

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
	// set rand SK for Schnorr signature
	wit.RandSK = new(big.Int).Set(randInputSK)

	cmInputSK := privacy.PedCom.CommitAtIndex(wit.spendingKey, randInputSK, privacy.SK)
	wit.ComInputSK = new(privacy.EllipticPoint)
	wit.ComInputSK.Set(cmInputSK.X, cmInputSK.Y)

	randInputShardID := privacy.RandInt()
	wit.ComInputShardID = privacy.PedCom.CommitAtIndex(big.NewInt(int64(pkLastByteSender)), randInputShardID, privacy.SHARDID)

	wit.ComInputValue = make([]*privacy.EllipticPoint, numInputCoin)
	wit.ComInputSND = make([]*privacy.EllipticPoint, numInputCoin)
	// It is used for proving 2 commitments commit to the same value (SND)
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
	wit.SerialNumberWitness = make([]*PKSNPrivacyWitness, numInputCoin)

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
			commitmentTemps[i][j] = commitments[preIndex+j].Sub(cmInputSum[i])
			//commitmentTemps[i][j].X, commitmentTemps[i][j].Y = privacy.Curve.Add(commitments[preIndex+j].X, commitments[preIndex+j].Y, cmInputSumInverse[i].X, cmInputSumInverse[i].Y)
		}

		if wit.OneOfManyWitness[i] == nil {
			wit.OneOfManyWitness[i] = new(OneOutOfManyWitness)
		}
		indexIsZero := myCommitmentIndices[i] % privacy.CMRingSize

		wit.OneOfManyWitness[i].Set(commitmentTemps[i], commitmentIndices[preIndex:preIndex+privacy.CMRingSize], randInputIsZero[i], indexIsZero, privacy.SK)
		preIndex = privacy.CMRingSize * (i + 1)
		// ---------------------------------------------------

		/***** Build witness for proving that serial number is derived from the committed derivator *****/
		if wit.SerialNumberWitness[i] == nil {
			wit.SerialNumberWitness[i] = new(PKSNPrivacyWitness)
		}
		wit.SerialNumberWitness[i].Set(inputCoin.CoinDetails.SerialNumber, cmInputSK, wit.ComInputSND[i], cmInputSNDIndexSK[i],
			spendingKey, randInputSK, inputCoin.CoinDetails.SNDerivator, randInputSND[i], randInputSNDIndexSK[i])
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
		randOutputValue[i] = privacy.RandInt()
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
		wit.ComOutputMultiRangeWitness = new(MultiRangeWitness)
	}
	wit.ComOutputMultiRangeWitness.Set(outputValue, 64)
	// ---------------------------------------------------

	// Build witness for proving Sum(Input's value) == Sum(Output's Value)
	if fee > 0 {
		cmOutputValueAll = cmOutputValueAll.Add(privacy.PedCom.G[privacy.VALUE].ScalarMult(big.NewInt(int64(fee))))
	}

	//cmEqualCoinValue := new(privacy.EllipticPoint)
	cmEqualCoinValue := cmInputValueAll.Sub(cmOutputValueAll)

	randEqualCoinValue := big.NewInt(0)
	randEqualCoinValue.Sub(randInputValueAll, randOutputValueAll)
	randEqualCoinValue.Mod(randEqualCoinValue, privacy.Curve.Params().N)

	wit.ComZeroWitness = new(ComZeroWitness)
	index := new(byte)
	*index = privacy.VALUE
	wit.ComZeroWitness.Set(cmEqualCoinValue, index, randEqualCoinValue)
	// ---------------------------------------------------

	// save partial commitments (value, snd, shardID)
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

	// if hasPrivacy == false, don't need to create the zero knowledge proof
	// proving user has spending key corresponding with public key in input coins
	// is proved by signing with spending key
	if !hasPrivacy {
		// Proving that serial number is derived from the committed derivator
		for i := 0; i < len(wit.inputCoins); i++ {
			snNoPrivacyProof, err := wit.SNNoPrivacyWitness[i].Prove()
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
		serialNumberProof, err := wit.SerialNumberWitness[i].Prove()
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

	// Proving that sum of all input values is equal to sum of all output values
	proof.ComZeroProof, err = wit.ComZeroWitness.Prove()
	if err != nil {
		return nil, privacy.NewPrivacyErr(privacy.ProvingErr, err)
	}

	privacy.Logger.Log.Info("Privacy log: PROVING DONE!!!")
	return proof, nil
}

func (pro PaymentProof) Verify(hasPrivacy bool, pubKey privacy.PublicKey, fee uint64, db database.DatabaseInterface, chainId byte, tokenID *common.Hash) bool {
	// has no privacy
	if !hasPrivacy {
		var sumInputValue, sumOutputValue uint64
		sumInputValue = 0
		sumOutputValue = 0

		for i := 0; i < len(pro.InputCoins); i++ {
			// Check input coins' Serial number is created from input coins' SND and sender's spending key
			if !pro.SNNoPrivacyProof[i].Verify() {
				return false
			}

			pubKeyLastByteSender := pubKey[len(pubKey)-1]

			// Check input coins' cm is calculated correctly
			cmTmp := pro.InputCoins[i].CoinDetails.PublicKey
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.VALUE].ScalarMult(big.NewInt(int64(pro.InputCoins[i].CoinDetails.Value))))
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SND].ScalarMult(pro.InputCoins[i].CoinDetails.SNDerivator))
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SHARDID].ScalarMult(new(big.Int).SetBytes([]byte{pubKeyLastByteSender})))
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.RAND].ScalarMult(pro.InputCoins[i].CoinDetails.Randomness))
			if !cmTmp.IsEqual(pro.InputCoins[i].CoinDetails.CoinCommitment) {
				return false
			}

			// Calculate sum of input values
			sumInputValue += pro.InputCoins[i].CoinDetails.Value
		}

		for i := 0; i < len(pro.OutputCoins); i++ {
			// Check output coins' cm is calculated correctly
			cmTmp := pro.OutputCoins[i].CoinDetails.PublicKey
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.VALUE].ScalarMult(big.NewInt(int64(pro.OutputCoins[i].CoinDetails.Value))))
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SND].ScalarMult(pro.OutputCoins[i].CoinDetails.SNDerivator))
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SHARDID].ScalarMult(new(big.Int).SetBytes([]byte{pro.OutputCoins[i].CoinDetails.GetPubKeyLastByte()})))
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.RAND].ScalarMult(pro.OutputCoins[i].CoinDetails.Randomness))
			if !cmTmp.IsEqual(pro.OutputCoins[i].CoinDetails.CoinCommitment) {
				return false
			}

			// Calculate sum of output values
			sumOutputValue += pro.OutputCoins[i].CoinDetails.Value
		}

		// check if sum of input values equal sum of output values
		if sumInputValue != sumOutputValue+fee {
			return false
		}
		return true
	}

	// if hasPrivacy == true
	// verify for input coins
	cmInputSum := make([]*privacy.EllipticPoint, len(pro.OneOfManyProof))
	for i := 0; i < len(pro.OneOfManyProof); i++ {
		// Verify for the proof one-out-of-N commitments is a commitment to the coins being spent
		// Calculate cm input sum
		cmInputSum[i] = new(privacy.EllipticPoint)
		cmInputSum[i] = pro.ComInputSK.Add(pro.ComInputValue[i])
		cmInputSum[i] = cmInputSum[i].Add(pro.ComInputSND[i])
		cmInputSum[i] = cmInputSum[i].Add(pro.ComInputShardID)

		// get commitments list from CommitmentIndices
		commitments := make([]*privacy.EllipticPoint, privacy.CMRingSize)
		for j := 0; j < privacy.CMRingSize; j++ {
			commitmentBytes, err := db.GetCommitmentByIndex(tokenID, pro.OneOfManyProof[i].CommitmentIndices[j], chainId)
			if err != nil {
				fmt.Printf("err 1\n")
				privacy.NewPrivacyErr(privacy.VerificationErr, errors.New("Zero knowledge verification error"))
				return false
			}
			commitments[j], err = privacy.DecompressKey(commitmentBytes)
			if err != nil {
				fmt.Printf("err 2\n")
				privacy.NewPrivacyErr(privacy.VerificationErr, errors.New("Zero knowledge verification error"))
				return false
			}

			commitments[j] = commitments[j].Sub(cmInputSum[i])
		}

		pro.OneOfManyProof[i].Commitments = commitments

		if !pro.OneOfManyProof[i].Verify() {
			fmt.Printf("err 3\n")
			return false
		}
		// Verify for the Proof that input coins' serial number is derived from the committed derivator
		if !pro.SerialNumberProof[i].Verify() {
			fmt.Printf("err 4\n")
			return false
		}
	}

	// Check output coins' cm is calculated correctly
	for i := 0; i < len(pro.OutputCoins); i++ {
		cmTmp := pro.OutputCoins[i].CoinDetails.PublicKey.Add(pro.ComOutputValue[i])
		cmTmp = cmTmp.Add(pro.ComOutputSND[i])
		cmTmp = cmTmp.Add(pro.ComOutputShardID[i])

		if !cmTmp.IsEqual(pro.OutputCoins[i].CoinDetails.CoinCommitment) {
			fmt.Printf("err 5\n")
			return false
		}
	}

	// Verify the proof that output values and sum of them do not exceed v_max
	if !pro.ComOutputMultiRangeProof.Verify() {
		fmt.Printf("err 6\n")
		return false
	}
	// Verify the proof that sum of all input values is equal to sum of all output values
	comInputValueSum := new(privacy.EllipticPoint).Zero()
	for i := 0; i < len(pro.ComInputValue); i++ {
		comInputValueSum = comInputValueSum.Add(pro.ComInputValue[i])
	}

	comOutputValueSum := new(privacy.EllipticPoint).Zero()
	for i := 0; i < len(pro.ComOutputValue); i++ {
		comOutputValueSum = comOutputValueSum.Add(pro.ComOutputValue[i])
	}

	if fee > 0 {
		comOutputValueSum = comOutputValueSum.Add(privacy.PedCom.G[privacy.VALUE].ScalarMult(big.NewInt(int64(fee))))
	}

	comZero := comInputValueSum.Sub(comOutputValueSum)
	pro.ComZeroProof.commitmentValue = comZero

	if !pro.ComZeroProof.Verify() {
		fmt.Printf("err 7\n")
		return false
	}

	return true
}
