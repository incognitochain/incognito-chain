package zkp

import (
	"encoding/json"
	"errors"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
	"math/big"
)

// PaymentWitness contains all of witness for proving when spending coins
type PaymentWitness struct {
	spendingKey        *big.Int
	RandSK             *big.Int
	inputCoins         []*privacy.InputCoin
	outputCoins        []*privacy.OutputCoin
	commitmentIndexs   []uint64
	myCommitmentIndexs []uint64

	pkLastByteSender    byte
	pkLastByteReceivers []byte

	OneOfManyWitness              []*PKOneOfManyWitness
	SerialNumberWitness []*PKSNPrivacyWitness
	SNNoPrivacyWitness []*PKSNNoPrivacyWitness

	ComOutputMultiRangeWitness *PKComMultiRangeWitness
	ComZeroWitness *PKComZeroWitness

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
	OneOfManyProof              []*PKOneOfManyProof
	SerialNumberProof []*PKSNPrivacyProof
	// it is exits when tx has no privacy
	SNNoPrivacyProof []*PKSNNoPrivacyProof

	// for output coins
	// for proving each value and sum of them are less than a threshold value
	ComOutputMultiRangeProof *PKComMultiRangeProof
	// for proving that the last element of output array is really the sum of all other values
	SumOutRangeProof *PKComZeroProof
	// for input = output
	ComZeroProof *PKComZeroProof

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
		OneOfManyProof:              []*PKOneOfManyProof{},
		SerialNumberProof: []*PKSNPrivacyProof{},
		ComOutputMultiRangeProof:    new(PKComMultiRangeProof).Init(),
		SumOutRangeProof:            new(PKComZeroProof).Init(),
		ComZeroProof:                new(PKComZeroProof).Init(),
		InputCoins:                  []*privacy.InputCoin{},
		OutputCoins:                 []*privacy.OutputCoin{},
		ComOutputValue:              []*privacy.EllipticPoint{},
		ComOutputSND:                []*privacy.EllipticPoint{},
		ComOutputShardID:            []*privacy.EllipticPoint{},
		ComInputSK:                  new(privacy.EllipticPoint),
		ComInputValue:               []*privacy.EllipticPoint{},
		ComInputSND:                 []*privacy.EllipticPoint{},
		ComInputShardID:             new(privacy.EllipticPoint),
	}
	return proof
}

func (proof PaymentProof) MarshalJSON() ([]byte, error) {
	data := proof.Bytes()
	temp := base58.Base58Check{}.Encode(data, byte(0x00))
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
	if paymentProof.ComOutputMultiRangeProof != nil {
		comOutputMultiRangeProof := paymentProof.ComOutputMultiRangeProof.Bytes()

		proofbytes = append(proofbytes, privacy.IntToByteArr(len(comOutputMultiRangeProof))...)
		proofbytes = append(proofbytes, comOutputMultiRangeProof...)
	} else {
		proofbytes = append(proofbytes, byte(0))
	}

	// SumOutRangeProof
	if paymentProof.SumOutRangeProof != nil {
		sumOutRangeProof := paymentProof.SumOutRangeProof.Bytes()
		proofbytes = append(proofbytes, byte(privacy.ComZeroProofSize))
		proofbytes = append(proofbytes, sumOutRangeProof...)
	} else {
		proofbytes = append(proofbytes, byte(0))
	}

	// ComZeroProof
	if paymentProof.ComZeroProof != nil {
		comZeroProof := paymentProof.ComZeroProof.Bytes()
		proofbytes = append(proofbytes, byte(privacy.ComZeroProofSize))
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

	return proofbytes
}

func (proof *PaymentProof) SetBytes(proofbytes []byte) (*privacy.PrivacyError) {
	offset := 0

	// Set OneOfManyProofSize
	lenOneOfManyProofArray := int(proofbytes[offset])
	offset += 1
	proof.OneOfManyProof = make([]*PKOneOfManyProof, lenOneOfManyProofArray)
	for i := 0; i < lenOneOfManyProofArray; i++ {
		lenOneOfManyProof := privacy.ByteArrToInt(proofbytes[offset: offset+2])
		offset += 2
		proof.OneOfManyProof[i] = new(PKOneOfManyProof).Init()
		err := proof.OneOfManyProof[i].SetBytes(proofbytes[offset: offset+lenOneOfManyProof])
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
		lenSerialNumberProof := privacy.ByteArrToInt(proofbytes[offset: offset+2])
		offset += 2
		proof.SerialNumberProof[i] = new(PKSNPrivacyProof).Init()
		err := proof.SerialNumberProof[i].SetBytes(proofbytes[offset: offset+lenSerialNumberProof])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenSerialNumberProof
	}

	// Set SNNoPrivacyProofSize
	lenSNNoPrivacyProofArray := int(proofbytes[offset])
	offset += 1
	proof.SNNoPrivacyProof = make([]*PKSNNoPrivacyProof, lenSNNoPrivacyProofArray)
	for i := 0; i < lenSNNoPrivacyProofArray; i++ {
		lenSNNoPrivacyProof := int(proofbytes[offset])
		offset += 1
		proof.SNNoPrivacyProof[i] = new(PKSNNoPrivacyProof).Init()
		err := proof.SNNoPrivacyProof[i].SetBytes(proofbytes[offset: offset+lenSNNoPrivacyProof])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenSNNoPrivacyProof
	}

	//ComOutputMultiRangeProofSize *PKComMultiRangeProof
	lenComOutputMultiRangeProof := privacy.ByteArrToInt(proofbytes[offset: offset+2])
	offset += 2
	if lenComOutputMultiRangeProof > 0 {
		proof.ComOutputMultiRangeProof = new(PKComMultiRangeProof).Init()
		err := proof.ComOutputMultiRangeProof.SetBytes(proofbytes[offset: offset+lenComOutputMultiRangeProof])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComOutputMultiRangeProof
	}

	//SumOutRangeProof *PKComZeroProof
	lenSumOutRangeProof := int(proofbytes[offset])
	offset += 1
	if lenSumOutRangeProof > 0 {
		proof.SumOutRangeProof = new(PKComZeroProof).Init()
		err := proof.SumOutRangeProof.SetBytes(proofbytes[offset: offset+lenSumOutRangeProof])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenSumOutRangeProof
	}

	//ComZeroProof *PKComZeroProof
	lenComZeroProof := int(proofbytes[offset])
	offset += 1
	if lenComZeroProof > 0 {
		proof.ComZeroProof = new(PKComZeroProof).Init()
		err := proof.ComZeroProof.SetBytes(proofbytes[offset: offset+lenComZeroProof])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComZeroProof
	}

	if len(proof.OneOfManyProof) == 0 {
		offset -= 1
	}

	//InputCoins  []*privacy.InputCoin
	lenInputCoinsArray := int(proofbytes[offset])
	offset += 1
	proof.InputCoins = make([]*privacy.InputCoin, lenInputCoinsArray)
	for i := 0; i < lenInputCoinsArray; i++ {
		lenInputCoin := int(proofbytes[offset])
		offset += 1
		proof.InputCoins[i] = new(privacy.InputCoin)
		err := proof.InputCoins[i].SetBytes(proofbytes[offset: offset+lenInputCoin])
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
		err := proof.OutputCoins[i].SetBytes(proofbytes[offset: offset+lenOutputCoin])
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
		proof.ComOutputValue[i], err = privacy.DecompressKey(proofbytes[offset: offset+lenComOutputValue])
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
		proof.ComOutputSND[i], err = privacy.DecompressKey(proofbytes[offset: offset+lenComOutputSND])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComOutputSND
	}

	if len(proof.OneOfManyProof) == 0 {
		offset -= 1
	}
	lenComOutputShardIdArray := int(proofbytes[offset])
	offset += 1
	proof.ComOutputShardID = make([]*privacy.EllipticPoint, lenComOutputShardIdArray)
	for i := 0; i < lenComOutputShardIdArray; i++ {
		lenComOutputShardId := int(proofbytes[offset])
		offset += 1
		proof.ComOutputShardID[i] = new(privacy.EllipticPoint)
		proof.ComOutputShardID[i], err = privacy.DecompressKey(proofbytes[offset: offset+lenComOutputShardId])
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
		proof.ComInputSK, err = privacy.DecompressKey(proofbytes[offset: offset+lenComInputSK])
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
		proof.ComInputValue[i], err = privacy.DecompressKey(proofbytes[offset: offset+lenComInputValue])
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
		proof.ComInputSND[i], err = privacy.DecompressKey(proofbytes[offset: offset+lenComInputSND])
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
		proof.ComInputShardID, err = privacy.DecompressKey(proofbytes[offset: offset+lenComInputShardID])
		if err != nil {
			return privacy.NewPrivacyErr(privacy.SetBytesProofErr, err)
		}
		offset += lenComInputShardID
	}

	return nil
}

// END----------------------------------------------------------------------------------------------------------------------------------------------

//func (wit *Pprivacy-protocol/zeroknowledge/zkp_opening.goaymentWitness) Set(spendingKey *big.Int, inputCoins []*privacy.InputCoin, outputCoins []*privacy.OutputCoin) {
//	wit.spendingKey = spendingKey
//	wit.inputCoins = inputCoins
//	wit.outputCoins = outputCoins
//}

// Build prepares witnesses for all protocol need to be proved when create tx
// if hashPrivacy = false, witness includes spending key, input coins, output coins
// otherwise, witness includes all attributes in PaymentWitness struct
func (wit *PaymentWitness) Build(hasPrivacy bool,
	spendingKey *big.Int,
	inputCoins []*privacy.InputCoin, outputCoins []*privacy.OutputCoin,
	pkLastByteSender byte, pkLastByteReceivers []byte,
	commitments []*privacy.EllipticPoint, commitmentIndexs []uint64, myCommitmentIndexs []uint64,
	fee uint64) (*privacy.PrivacyError) {

	if !hasPrivacy {
		wit.spendingKey = spendingKey
		wit.inputCoins = inputCoins
		wit.outputCoins = outputCoins
		wit.commitmentIndexs = commitmentIndexs
		wit.myCommitmentIndexs = myCommitmentIndexs
		wit.pkLastByteSender = pkLastByteSender

		publicKey := inputCoins[0].CoinDetails.PublicKey

		for i := 0; i < len(inputCoins); i++ {
			/***** Build witness for proving that serial number is derived from the committed derivator *****/
			if wit.SNNoPrivacyWitness[i] == nil {
				wit.SNNoPrivacyWitness[i] = new(PKSNNoPrivacyWitness)
			}
			wit.SNNoPrivacyWitness[i].Set(inputCoins[i].CoinDetails.SerialNumber, publicKey, inputCoins[i].CoinDetails.SNDerivator, wit.spendingKey)
		}
		return nil
	}

	wit.spendingKey = spendingKey
	wit.inputCoins = inputCoins
	wit.outputCoins = outputCoins
	wit.commitmentIndexs = commitmentIndexs
	wit.myCommitmentIndexs = myCommitmentIndexs
	wit.pkLastByteSender = pkLastByteSender

	numInputCoin := len(wit.inputCoins)

	randInputSK := privacy.RandInt()
	// set rand SK for Schnorr signature
	wit.RandSK = new(big.Int).Set(randInputSK)

	cmInputSK := privacy.PedCom.CommitAtIndex(wit.spendingKey, randInputSK, privacy.SK)
	wit.ComInputSK = new(privacy.EllipticPoint). Zero()
	wit.ComInputSK.X.Set(cmInputSK.X)
	wit.ComInputSK.Y.Set(cmInputSK.Y)
	randInputShardID := privacy.RandInt()
	wit.ComInputShardID = privacy.PedCom.CommitAtIndex(big.NewInt(int64(wit.pkLastByteSender)), randInputShardID, privacy.SHARDID)
	wit.ComInputValue = make([]*privacy.EllipticPoint, numInputCoin)
	wit.ComInputSND = make([]*privacy.EllipticPoint, numInputCoin)
	// It is used for proving 2 commitments commit to the same value (SND)
	cmInputSNDIndexSK := make([]*privacy.EllipticPoint, numInputCoin)

	randInputValue := make([]*big.Int, numInputCoin)
	randInputSND := make([]*big.Int, numInputCoin)
	randInputSNDIndexSK := make([]*big.Int, numInputCoin)

	// cmInputValueAll is sum of all value coin commitments
	cmInputValueAll := privacy.EllipticPoint{big.NewInt(0), big.NewInt(0)}
	randInputValueAll := big.NewInt(0)

	// commit each component of coin commitment
	for i, inputCoin := range wit.inputCoins {
		randInputValue[i] = privacy.RandInt()
		randInputSND[i] = privacy.RandInt()
		randInputSNDIndexSK[i] = privacy.RandInt()

		wit.ComInputValue[i] = privacy.PedCom.CommitAtIndex(new(big.Int).SetUint64(inputCoin.CoinDetails.Value), randInputValue[i], privacy.VALUE)
		wit.ComInputSND[i] = privacy.PedCom.CommitAtIndex(inputCoin.CoinDetails.SNDerivator, randInputSND[i], privacy.SND)
		cmInputSNDIndexSK[i] = privacy.PedCom.CommitAtIndex(inputCoin.CoinDetails.SNDerivator, randInputSNDIndexSK[i], privacy.SK)

		cmInputValueAll = *cmInputValueAll.Add(wit.ComInputValue[i])
		randInputValueAll.Add(randInputValueAll, randInputValue[i])
		randInputValueAll.Mod(randInputValueAll, privacy.Curve.Params().N)
	}

	// Summing all commitments of each input coin into one commitment and proving the knowledge of its Openings
	cmInputSum := make([]*privacy.EllipticPoint, numInputCoin)
	cmInputSumInverse := make([]*privacy.EllipticPoint, numInputCoin)
	//cmInputSumInverse := make([]*privacy.EllipticPoint, numInputCoin)
	randInputSum := make([]*big.Int, numInputCoin)
	// randInputSumAll is sum of all randomess of coin commitments
	randInputSumAll := big.NewInt(0)

	wit.OneOfManyWitness = make([]*PKOneOfManyWitness, numInputCoin)
	wit.SerialNumberWitness = make([]*PKSNPrivacyWitness, numInputCoin)

	commitmentTemps := make([][]*privacy.EllipticPoint, numInputCoin)
	rndInputIsZero := make([]*big.Int, numInputCoin)
	//commitmentIndexTemps := make([][])

	preIndex := 0
	var err error
	for i := 0; i < numInputCoin; i++ {
		/***** Build witness for proving one-out-of-N commitments is a commitment to the coins being spent *****/
		cmInputSum[i] = new(privacy.EllipticPoint)
		cmInputSum[i].X, cmInputSum[i].Y = big.NewInt(0), big.NewInt(0)
		cmInputSum[i].X.Set(cmInputSK.X)
		cmInputSum[i].Y.Set(cmInputSK.Y)

		cmInputSum[i].X, cmInputSum[i].Y = privacy.Curve.Add(cmInputSum[i].X, cmInputSum[i].Y, wit.ComInputValue[i].X, wit.ComInputValue[i].Y)
		cmInputSum[i].X, cmInputSum[i].Y = privacy.Curve.Add(cmInputSum[i].X, cmInputSum[i].Y, wit.ComInputSND[i].X, wit.ComInputSND[i].Y)
		cmInputSum[i].X, cmInputSum[i].Y = privacy.Curve.Add(cmInputSum[i].X, cmInputSum[i].Y, wit.ComInputShardID.X, wit.ComInputShardID.Y)

		randInputSum[i] = new(big.Int).Set(randInputSK)
		randInputSum[i].Add(randInputSum[i], randInputValue[i])
		randInputSum[i].Add(randInputSum[i], randInputSND[i])
		randInputSum[i].Add(randInputSum[i], randInputShardID)
		randInputSum[i].Mod(randInputSum[i], privacy.Curve.Params().N)

		randInputSumAll.Add(randInputSumAll, randInputSum[i])
		randInputSumAll.Mod(randInputSumAll, privacy.Curve.Params().N)

		// commitmentTemps is a list of commitments for protocol one-out-of-N
		commitmentTemps[i] = make([]*privacy.EllipticPoint, privacy.CMRingSize)

		rndInputIsZero[i] = big.NewInt(0)
		rndInputIsZero[i].Set(inputCoins[i].CoinDetails.Randomness)
		rndInputIsZero[i].Sub(rndInputIsZero[i], randInputSum[i])
		rndInputIsZero[i].Mod(rndInputIsZero[i], privacy.Curve.Params().N)

		cmInputSumInverse[i], err = cmInputSum[i].Inverse()
		if err != nil {
			return privacy.NewPrivacyErr(privacy.UnexpectedErr, err)
		}

		for j := 0; j < privacy.CMRingSize; j++ {
			commitmentTemps[i][j] = new(privacy.EllipticPoint).Zero()
			commitmentTemps[i][j].X, commitmentTemps[i][j].Y = privacy.Curve.Add(commitments[preIndex+j].X, commitments[preIndex+j].Y, cmInputSumInverse[i].X, cmInputSumInverse[i].Y)
		}

		if wit.OneOfManyWitness[i] == nil {
			wit.OneOfManyWitness[i] = new(PKOneOfManyWitness)
		}
		indexIsZero := myCommitmentIndexs[i] % privacy.CMRingSize

		wit.OneOfManyWitness[i].Set(commitmentTemps[i], commitmentIndexs[preIndex:preIndex+privacy.CMRingSize], rndInputIsZero[i], indexIsZero, privacy.SK)
		preIndex = privacy.CMRingSize * (i + 1)

		/***** Build witness for proving that serial number is derived from the committed derivator *****/
		if wit.SerialNumberWitness[i] == nil {
			wit.SerialNumberWitness[i] = new(PKSNPrivacyWitness)
		}
		wit.SerialNumberWitness[i].Set(inputCoins[i].CoinDetails.SerialNumber, cmInputSK, wit.ComInputSND[i], cmInputSNDIndexSK[i],
			spendingKey, randInputSK, inputCoins[i].CoinDetails.SNDerivator, randInputSND[i], randInputSNDIndexSK[i])
	}

	numOutputCoin := len(wit.outputCoins)

	randOutputValue := make([]*big.Int, numOutputCoin)
	randOutputSND := make([]*big.Int, numOutputCoin)
	cmOutputValue := make([]*privacy.EllipticPoint, numOutputCoin)
	cmOutputSND := make([]*privacy.EllipticPoint, numOutputCoin)

	cmOutputSum := make([]*privacy.EllipticPoint, numOutputCoin)
	randOutputSum := make([]*big.Int, numOutputCoin)

	cmOutputSumAll := new(privacy.EllipticPoint)
	cmOutputSumAll.X = big.NewInt(0)
	cmOutputSumAll.Y = big.NewInt(0)

	// cmOutputValueAll is sum of all value coin commitments
	cmOutputValueAll := privacy.EllipticPoint{big.NewInt(0), big.NewInt(0)}
	randOutputValueAll := big.NewInt(0)

	randOutputShardID := make([]*big.Int, numOutputCoin)
	cmOutputShardID := make([]*privacy.EllipticPoint, numOutputCoin)

	for i, outputCoin := range wit.outputCoins {
		randOutputValue[i] = privacy.RandInt()
		randOutputSND[i] = privacy.RandInt()
		cmOutputValue[i] = privacy.PedCom.CommitAtIndex(new(big.Int).SetUint64(outputCoin.CoinDetails.Value), randOutputValue[i], privacy.VALUE)
		cmOutputSND[i] = privacy.PedCom.CommitAtIndex(outputCoin.CoinDetails.SNDerivator, randOutputSND[i], privacy.SND)
		randOutputShardID[i] = privacy.RandInt()
		cmOutputShardID[i] = privacy.PedCom.CommitAtIndex(big.NewInt(int64(outputCoins[i].CoinDetails.GetPubKeyLastByte())), randOutputShardID[i], privacy.SHARDID)

		randOutputSum[i] = big.NewInt(0)
		randOutputSum[i].Set(randOutputValue[i])
		randOutputSum[i].Add(randOutputSum[i], randOutputSND[i])
		randOutputSum[i].Add(randOutputSum[i], randOutputShardID[i])
		randOutputSum[i].Mod(randOutputSum[i], privacy.Curve.Params().N)

		cmOutputSum[i] = new(privacy.EllipticPoint).Zero()
		cmOutputSum[i].Set(cmOutputValue[i].X, cmOutputValue[i].Y)
		cmOutputSum[i].X, cmOutputSum[i].Y = privacy.Curve.Add(cmOutputSum[i].X, cmOutputSum[i].Y, cmOutputSND[i].X, cmOutputSND[i].Y)
		cmOutputSum[i] = cmOutputSum[i].Add(outputCoins[i].CoinDetails.PublicKey)
		cmOutputSum[i] = cmOutputSum[i].Add(cmOutputShardID[i])

		cmOutputValueAll = *(cmOutputValueAll.Add(cmOutputValue[i]))
		randOutputValueAll.Add(randOutputValueAll, randOutputValue[i])

		// calculate final commitment for output coins

		outputCoins[i].CoinDetails.CoinCommitment = cmOutputSum[i]
		outputCoins[i].CoinDetails.Randomness = randOutputSum[i]

		cmOutputSumAll.X, cmOutputSumAll.Y = privacy.Curve.Add(cmOutputSumAll.X, cmOutputSumAll.Y, cmOutputSum[i].X, cmOutputSum[i].Y)
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
		wit.ComOutputMultiRangeWitness = new(PKComMultiRangeWitness)
	}
	wit.ComOutputMultiRangeWitness.Set(outputValue, 64)
	// ------------------------

	// Build witness for proving Sum(Input's value) == Sum(Output's Value)
	if fee > 0 {
		cmOutputValueAll.Add(privacy.PedCom.G[privacy.VALUE].ScalarMult(big.NewInt(int64(fee))))
	}

	cmOutputValueAllInverse, err := cmOutputValueAll.Inverse()
	if err != nil {
		return privacy.NewPrivacyErr(privacy.UnexpectedErr, err)
	}

	cmEqualCoinValue := new(privacy.EllipticPoint)
	cmEqualCoinValue = cmInputValueAll.Add(cmOutputValueAllInverse)

	randEqualCoinValue := big.NewInt(0)
	randEqualCoinValue.Sub(randInputValueAll, randOutputValueAll)
	randEqualCoinValue.Mod(randEqualCoinValue, privacy.Curve.Params().N)

	wit.ComZeroWitness = new(PKComZeroWitness)
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
		if err != nil{
			return nil, privacy.NewPrivacyErr(privacy.ProvingErr, err)
		}
		proof.SerialNumberProof = append(proof.SerialNumberProof, serialNumberProof)
	}

	// Proving that each output values and sum of them does not exceed v_max
	proof.ComOutputMultiRangeProof, err = wit.ComOutputMultiRangeWitness.Prove()
	if err != nil {
		return nil, privacy.NewPrivacyErr(privacy.ProvingErr, err)
	}
	proof.SumOutRangeProof, err = wit.ComOutputMultiRangeWitness.ProveSum()
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

func (pro PaymentProof) Verify(hasPrivacy bool, pubKey privacy.PublicKey, db database.DatabaseInterface, chainId byte, tokenID *common.Hash) bool {
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
		if sumInputValue != sumOutputValue {
			return false
		}
		return true
	}

	// if hasPrivacy == true
	// verify for input coins
	var err error
	cmInputSum := make([]*privacy.EllipticPoint, len(pro.OneOfManyProof))
	cmInputSumInverse := make([]*privacy.EllipticPoint, len(pro.OneOfManyProof))
	for i := 0; i < len(pro.OneOfManyProof); i++ {
		// Verify the proof for knowledge of input coins' Openings
		/*if !pro.ComInputOpeningsProof[i].Verify() {
			return false
		}*/
		// Verify for the proof one-out-of-N commitments is a commitment to the coins being spent
		// Calculate cm input inverse
		cmInputSum[i] = new(privacy.EllipticPoint)
		cmInputSum[i].X = big.NewInt(0)
		cmInputSum[i].Y = big.NewInt(0)
		cmInputSum[i].X.Set(pro.ComInputSK.X)
		cmInputSum[i].Y.Set(pro.ComInputSK.Y)

		cmInputSum[i] = cmInputSum[i].Add(pro.ComInputValue[i])
		cmInputSum[i] = cmInputSum[i].Add(pro.ComInputSND[i])
		cmInputSum[i] = cmInputSum[i].Add(pro.ComInputShardID)

		cmInputSumInverse[i], err = cmInputSum[i].Inverse()
		if err != nil {
			return false
		}

		// get commitments list from CommitmentIndices
		commitments := make([]*privacy.EllipticPoint, privacy.CMRingSize)
		for j := 0; j < privacy.CMRingSize; j++ {
			commitmentBytes, err := db.GetCommitmentByIndex(tokenID, pro.OneOfManyProof[i].CommitmentIndices[j], chainId)
			if err != nil {
				privacy.NewPrivacyErr(privacy.VerificationErr, errors.New("Zero knowledge verification error"))
				return false
			}
			commitments[j], err = privacy.DecompressKey(commitmentBytes)
			if err != nil {
				privacy.NewPrivacyErr(privacy.VerificationErr, errors.New("Zero knowledge verification error"))
				return false
			}

			commitments[j] = commitments[j].Add(cmInputSumInverse[i])
		}

		pro.OneOfManyProof[i].Commitments = commitments

		if !pro.OneOfManyProof[i].Verify() {
			return false
		}
		// Verify for the Proof that input coins' serial number is derived from the committed derivator
		if !pro.SerialNumberProof[i].Verify() {
			return false
		}
	}

	// Check output coins' cm is calculated correctly
	for i := 0; i < len(pro.OutputCoins); i++ {
		cmTmp := pro.OutputCoins[i].CoinDetails.PublicKey
		cmTmp = cmTmp.Add(pro.ComOutputValue[i])
		cmTmp = cmTmp.Add(pro.ComOutputSND[i])
		cmTmp = cmTmp.Add(pro.ComOutputShardID[i])

		if !cmTmp.IsEqual(pro.OutputCoins[i].CoinDetails.CoinCommitment) {
			return false
		}
	}

	// Verify the proof that output values and sum of them do not exceed v_max
	if !pro.ComOutputMultiRangeProof.Verify() {
		return false
	}
	// Verify the last values of array is really the sum of all output value
	if !pro.ComOutputMultiRangeProof.VerifySum(pro.SumOutRangeProof) {
		return false
	}
	// Verify the proof that sum of all input values is equal to sum of all output values
	if !pro.ComZeroProof.Verify() {
		return false
	}

	return true
}
