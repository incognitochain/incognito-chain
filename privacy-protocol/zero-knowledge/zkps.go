package zkp

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/database"
	privacy "github.com/ninjadotorg/constant/privacy-protocol"
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

	ComInputOpeningsWitness       []*PKComOpeningsWitness
	OneOfManyWitness              []*PKOneOfManyWitness
	EqualityOfCommittedValWitness []*PKEqualityOfCommittedValWitness
	ProductCommitmentWitness      []*PKComProductWitness

	ComOutputOpeningsWitness []*PKComOpeningsWitness

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
	ComInputOpeningsProof       []*PKComOpeningsProof
	OneOfManyProof              []*PKOneOfManyProof
	EqualityOfCommittedValProof []*PKEqualityOfCommittedValProof
	ProductCommitmentProof      []*PKComProductProof
	// for output coins
	ComOutputOpeningsProof []*PKComOpeningsProof
	// for proving each value and sum of them are less than a threshold value
	ComOutputMultiRangeProof *PKComMultiRangeProof
	// for proving that the last element of output array is really the sum of all other values
	SumOutRangeProof *PKComZeroProof
	// for input = output
	ComZeroProof *PKComZeroProof
	// add list input coins' SN to proof for serial number
	// these following attributes just exist when tx doesn't have privacy
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
		ComInputOpeningsProof:       []*PKComOpeningsProof{},
		OneOfManyProof:              []*PKOneOfManyProof{},
		EqualityOfCommittedValProof: []*PKEqualityOfCommittedValProof{},
		ProductCommitmentProof:      []*PKComProductProof{},
		ComOutputOpeningsProof:      []*PKComOpeningsProof{},
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
	// ComInputOpeningsProof
	lenComInputOpeningsProofArray := len(paymentProof.ComInputOpeningsProof)
	proofbytes = append(proofbytes, byte(lenComInputOpeningsProofArray))

	for i := 0; i < len(paymentProof.ComInputOpeningsProof); i++ {
		comInputOpeningProof := paymentProof.ComInputOpeningsProof[i].Bytes()
		proofbytes = append(proofbytes, byte(len(comInputOpeningProof)))
		proofbytes = append(proofbytes, comInputOpeningProof...)
	}
	// OneOfManyProofSize
	lenOneOfManyProofArray := len(paymentProof.OneOfManyProof)
	proofbytes = append(proofbytes, byte(lenOneOfManyProofArray))

	for i := 0; i < len(paymentProof.OneOfManyProof); i++ {
		oneOfManyProof := paymentProof.OneOfManyProof[i].Bytes()
		proofbytes = append(proofbytes, privacy.IntToByteArr(len(oneOfManyProof))...)
		proofbytes = append(proofbytes, oneOfManyProof...)
	}
	// EqualityOfCommittedValProofSize
	lenEqualityOfCommittedValProofArray := len(paymentProof.EqualityOfCommittedValProof)
	proofbytes = append(proofbytes, byte(lenEqualityOfCommittedValProofArray))

	for i := 0; i < len(paymentProof.EqualityOfCommittedValProof); i++ {
		equalityOfCommittedValProof := paymentProof.EqualityOfCommittedValProof[i].Bytes()
		proofbytes = append(proofbytes, byte(len(equalityOfCommittedValProof)))
		proofbytes = append(proofbytes, equalityOfCommittedValProof...)
	}
	// ProductCommitmentProofSize
	proofbytes = append(proofbytes, byte(len(paymentProof.ProductCommitmentProof)))
	for i := 0; i < len(paymentProof.ProductCommitmentProof); i++ {
		productCommitmentProof := paymentProof.ProductCommitmentProof[i].Bytes()
		proofbytes = append(proofbytes, byte(len(productCommitmentProof)))
		proofbytes = append(proofbytes, paymentProof.ProductCommitmentProof[i].Bytes()...)

	}
	//ComOutputOpeningsProofSize
	proofbytes = append(proofbytes, byte(len(paymentProof.ComOutputOpeningsProof)))
	for i := 0; i < len(paymentProof.ComOutputOpeningsProof); i++ {
		comOutputOpeningsProof := paymentProof.ComOutputOpeningsProof[i].Bytes()
		proofbytes = append(proofbytes, byte(len(comOutputOpeningsProof)))
		proofbytes = append(proofbytes, comOutputOpeningsProof...)
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
	//paymentProof.SumOutRangeProof = nil
	if paymentProof.SumOutRangeProof != nil {
		sumOutRangeProof := paymentProof.SumOutRangeProof.Bytes()
		proofbytes = append(proofbytes, byte(len(sumOutRangeProof)))
		proofbytes = append(proofbytes, sumOutRangeProof...)
	} else {
		proofbytes = append(proofbytes, byte(0))
	}

	// ComZeroProof
	//paymentProof.ComZeroProof = nil
	if paymentProof.ComZeroProof != nil {
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
		proofbytes = append(proofbytes, byte(len(comOutputValue)))
		proofbytes = append(proofbytes, comOutputValue...)
	}
	// ComOutputSND
	proofbytes = append(proofbytes, byte(len(paymentProof.ComOutputSND)))
	for i := 0; i < len(paymentProof.ComOutputSND); i++ {
		comOutputSND := paymentProof.ComOutputSND[i].Compress()
		proofbytes = append(proofbytes, byte(len(comOutputSND)))
		proofbytes = append(proofbytes, comOutputSND...)
	}
	// ComOutputShardID
	proofbytes = append(proofbytes, byte(len(paymentProof.ComOutputShardID)))
	for i := 0; i < len(paymentProof.ComOutputShardID); i++ {
		comOutputShardID := paymentProof.ComOutputShardID[i].Compress()
		proofbytes = append(proofbytes, byte(len(comOutputShardID)))
		proofbytes = append(proofbytes, comOutputShardID...)
	}

	//ComInputSK 				*privacy.EllipticPoint
	if paymentProof.ComInputSK != nil {
		comInputSK := paymentProof.ComInputSK.Compress()
		proofbytes = append(proofbytes, byte(len(comInputSK)))
		proofbytes = append(proofbytes, comInputSK...)
	} else {
		proofbytes = append(proofbytes, byte(0))
	}
	//ComInputValue 		[]*privacy.EllipticPoint
	proofbytes = append(proofbytes, byte(len(paymentProof.ComInputValue)))
	for i := 0; i < len(paymentProof.ComInputValue); i++ {
		comInputValue := paymentProof.ComInputValue[i].Compress()
		proofbytes = append(proofbytes, byte(len(comInputValue)))
		proofbytes = append(proofbytes, comInputValue...)
	}
	//ComInputSND 			[]*privacy.EllipticPoint
	proofbytes = append(proofbytes, byte(len(paymentProof.ComInputSND)))
	for i := 0; i < len(paymentProof.ComInputSND); i++ {
		comInputSND := paymentProof.ComInputSND[i].Compress()
		proofbytes = append(proofbytes, byte(len(comInputSND)))
		proofbytes = append(proofbytes, comInputSND...)
	}
	//ComInputShardID 	*privacy.EllipticPoint
	if paymentProof.ComInputShardID != nil {
		comInputShardID := paymentProof.ComInputShardID.Compress()
		proofbytes = append(proofbytes, byte(len(comInputShardID)))
		proofbytes = append(proofbytes, comInputShardID...)
	} else {
		proofbytes = append(proofbytes, byte(0))
	}

	return proofbytes
}

func (proof *PaymentProof) SetBytes(proofbytes []byte) (err error) {

	//fmt.Printf("***************BEFORE SETBYTE - PROOF %v\n", proofbytes)
	offset := 0
	// Set ComInputOpeningsProof
	lenComInputOpeningsProofArray := int(proofbytes[offset])
	offset += 1
	proof.ComInputOpeningsProof = make([]*PKComOpeningsProof, lenComInputOpeningsProofArray)
	for i := 0; i < lenComInputOpeningsProofArray; i++ {
		lenComInputOpeningsProof := int(proofbytes[offset])
		offset += 1

		proof.ComInputOpeningsProof[i] = new(PKComOpeningsProof).Init()
		proof.ComInputOpeningsProof[i].SetBytes(proofbytes[offset : offset+lenComInputOpeningsProof])
		offset += lenComInputOpeningsProof
	}
	// Set OneOfManyProofSize
	lenOneOfManyProofArray := int(proofbytes[offset])
	offset += 1
	proof.OneOfManyProof = make([]*PKOneOfManyProof, lenOneOfManyProofArray)
	for i := 0; i < lenOneOfManyProofArray; i++ {
		lenOneOfManyProof := privacy.ByteArrToInt(proofbytes[offset : offset+2])
		offset += 2

		proof.OneOfManyProof[i] = new(PKOneOfManyProof).Init()
		proof.OneOfManyProof[i].SetBytes(proofbytes[offset : offset+lenOneOfManyProof])
		offset += lenOneOfManyProof
	}
	// Set EqualityOfCommittedValProofSize
	lenEqualityOfCommittedValProofArray := int(proofbytes[offset])
	offset += 1
	proof.EqualityOfCommittedValProof = make([]*PKEqualityOfCommittedValProof, lenEqualityOfCommittedValProofArray)
	for i := 0; i < lenEqualityOfCommittedValProofArray; i++ {
		lenEqualityOfCommittedValProof := int(proofbytes[offset])
		offset += 1

		proof.EqualityOfCommittedValProof[i] = new(PKEqualityOfCommittedValProof).Init()
		err := proof.EqualityOfCommittedValProof[i].SetBytes(proofbytes[offset : offset+lenEqualityOfCommittedValProof])
		if err != nil {
			return err
		}
		offset += lenEqualityOfCommittedValProof
	}
	// Set ProductCommitmentProofSize
	lenProductCommitmentProofArray := int(proofbytes[offset])
	offset += 1
	proof.ProductCommitmentProof = make([]*PKComProductProof, lenProductCommitmentProofArray)
	for i := 0; i < lenProductCommitmentProofArray; i++ {
		lenProductCommitmentProof := int(proofbytes[offset])
		offset += 1

		proof.ProductCommitmentProof[i] = new(PKComProductProof).Init()
		proof.ProductCommitmentProof[i].SetBytes(proofbytes[offset : offset+lenProductCommitmentProof])
		offset += lenProductCommitmentProof
	}
	//Set ComOutputOpeningsProofSize
	lenComOutputOpeningsProofArray := int(proofbytes[offset])
	offset += 1
	proof.ComOutputOpeningsProof = make([]*PKComOpeningsProof, lenComOutputOpeningsProofArray)
	for i := 0; i < lenComOutputOpeningsProofArray; i++ {
		lenComOutputOpeningsProof := int(proofbytes[offset])
		offset += 1

		proof.ComOutputOpeningsProof[i] = new(PKComOpeningsProof).Init()
		proof.ComOutputOpeningsProof[i].SetBytes(proofbytes[offset : offset+lenComOutputOpeningsProof])
		offset += lenComOutputOpeningsProof
	}

	//ComOutputMultiRangeProofSize *PKComMultiRangeProof
	lenComOutputMultiRangeProof := privacy.ByteArrToInt(proofbytes[offset : offset+2])
	offset += 2
	if lenComOutputMultiRangeProof > 0 {
		proof.ComOutputMultiRangeProof = new(PKComMultiRangeProof).Init()
		proof.ComOutputMultiRangeProof.SetBytes(proofbytes[offset : offset+lenComOutputMultiRangeProof])
		offset += lenComOutputMultiRangeProof
	}

	//SumOutRangeProof *PKComZeroProof
	lenSumOutRangeProof := int(proofbytes[offset])
	offset += 1
	if lenSumOutRangeProof > 0 {
		proof.SumOutRangeProof = new(PKComZeroProof).Init()
		proof.SumOutRangeProof.SetBytes(proofbytes[offset : offset+lenSumOutRangeProof])
		offset += lenSumOutRangeProof
	}

	//ComZeroProof *PKComZeroProof
	lenComZeroProof := int(proofbytes[offset])
	offset += 1
	if lenComZeroProof > 0 {
		proof.ComZeroProof = new(PKComZeroProof).Init()
		proof.ComZeroProof.SetBytes(proofbytes[offset : offset+lenComZeroProof])
		offset += lenComZeroProof
	}

	if len(proof.ComInputOpeningsProof) == 0 {
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
		proof.InputCoins[i].SetBytes(proofbytes[offset : offset+lenInputCoin])
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
			return err
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
		proof.ComOutputValue[i], err = privacy.DecompressKey(proofbytes[offset : offset+lenComOutputValue])
		if err != nil {
			return err
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
			return err
		}
		offset += lenComOutputSND
	}
	//ComOutputShardID []*privacy.EllipticPoint
	if len(proof.ComInputOpeningsProof) == 0 {
		offset -= 1
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
			return err
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
			return err
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
			return err
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
			return err
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
			return err
		}
		offset += lenComInputShardID
	}
	//fmt.Printf("***************AFTER SETBYTE - PROOF %v\n", proof.Bytes())
	return nil
}

// END----------------------------------------------------------------------------------------------------------------------------------------------

//func (wit *Pprivacy-protocol/zero-knowledge/zkp_opening.goaymentWitness) Set(spendingKey *big.Int, inputCoins []*privacy.InputCoin, outputCoins []*privacy.OutputCoin) {
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
	fee uint64) (err error) {

	if !hasPrivacy {
		wit.spendingKey = spendingKey
		wit.inputCoins = inputCoins
		wit.outputCoins = outputCoins
		wit.commitmentIndexs = commitmentIndexs
		wit.myCommitmentIndexs = myCommitmentIndexs

		for i := 0; i < len(inputCoins); i++ {
			/***** Build witness for proving that serial number is derived from the committed derivator *****/

			/****Build witness for proving that the commitment of serial number is equivalent to Mul(com(sk), com(snd))****/
			witnesssA := new(big.Int)
			witnesssA.Add(wit.spendingKey, inputCoins[i].CoinDetails.SNDerivator)

			randA := big.NewInt(0)
			witIndex := new(byte)
			*witIndex = privacy.SK
			if wit.ProductCommitmentWitness[i] == nil {
				wit.ProductCommitmentWitness[i] = new(PKComProductWitness)
			}
			wit.ProductCommitmentWitness[i].Set(witnesssA, randA, inputCoins[i].CoinDetails.SerialNumber, witIndex)
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
	wit.ComInputSK = new(privacy.EllipticPoint)
	wit.ComInputSK.X = big.NewInt(0)
	wit.ComInputSK.Y = big.NewInt(0)
	wit.ComInputSK.X.Set(cmInputSK.X)
	wit.ComInputSK.Y.Set(cmInputSK.Y)
	randInputShardID := privacy.RandInt()
	wit.ComInputShardID = privacy.PedCom.CommitAtIndex(big.NewInt(int64(wit.pkLastByteSender)), randInputShardID, privacy.SHARDID)
	//wit.ComInputShardID = privacy.PedCom.CommitAtIndex(big.NewInt(int64(inputCoins[0].CoinDetails.GetPubKeyLastByte())), randInputShardID, privacy.SHARDID)
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
	randInputSumwithoutShardID := make([]*big.Int, numInputCoin)
	cmInputSumwithoutShardID := make([]*privacy.EllipticPoint, numInputCoin)
	// randInputSumAll is sum of all randomess of coin commitments
	randInputSumAll := big.NewInt(0)

	wit.ComInputOpeningsWitness = make([]*PKComOpeningsWitness, numInputCoin)
	wit.OneOfManyWitness = make([]*PKOneOfManyWitness, numInputCoin)
	wit.ProductCommitmentWitness = make([]*PKComProductWitness, numInputCoin)
	wit.EqualityOfCommittedValWitness = make([]*PKEqualityOfCommittedValWitness, numInputCoin)
	indexZKPEqual := make([]byte, 2)
	indexZKPEqual[0] = privacy.SK
	indexZKPEqual[1] = privacy.SND

	commitmentTemps := make([][]*privacy.EllipticPoint, numInputCoin)
	rndInputIsZero := make([]*big.Int, numInputCoin)
	//commitmentIndexTemps := make([][])

	preIndex := 0
	for i := 0; i < numInputCoin; i++ {
		/***** Build witness for proving the knowledge of input coins' Openings  *****/
		cmInputSum[i] = new(privacy.EllipticPoint)
		cmInputSum[i].X, cmInputSum[i].Y = big.NewInt(0), big.NewInt(0)
		cmInputSum[i].X.Set(cmInputSK.X)
		cmInputSum[i].Y.Set(cmInputSK.Y)
		//cmInputSum[i] = cmInputSK
		cmInputSum[i].X, cmInputSum[i].Y = privacy.Curve.Add(cmInputSum[i].X, cmInputSum[i].Y, wit.ComInputValue[i].X, wit.ComInputValue[i].Y)
		cmInputSum[i].X, cmInputSum[i].Y = privacy.Curve.Add(cmInputSum[i].X, cmInputSum[i].Y, wit.ComInputSND[i].X, wit.ComInputSND[i].Y)
		cmInputSumwithoutShardID[i] = new(privacy.EllipticPoint)
		cmInputSumwithoutShardID[i].X, cmInputSumwithoutShardID[i].Y = big.NewInt(0), big.NewInt(0)
		cmInputSumwithoutShardID[i].X.Set(cmInputSum[i].X)
		cmInputSumwithoutShardID[i].Y.Set(cmInputSum[i].Y)

		randInputSum[i] = new(big.Int).Set(randInputSK)
		randInputSum[i].Add(randInputSum[i], randInputValue[i])
		randInputSum[i].Mod(randInputSum[i], privacy.Curve.Params().N)
		randInputSum[i].Add(randInputSum[i], randInputSND[i])
		randInputSum[i].Mod(randInputSum[i], privacy.Curve.Params().N)
		randInputSumwithoutShardID[i] = big.NewInt(0)
		randInputSumwithoutShardID[i].Set(randInputSum[i])

		if wit.ComInputOpeningsWitness[i] == nil {
			wit.ComInputOpeningsWitness[i] = new(PKComOpeningsWitness)
		}
		wit.ComInputOpeningsWitness[i].Set(cmInputSumwithoutShardID[i],
			[]*big.Int{wit.spendingKey, big.NewInt(int64(inputCoins[i].CoinDetails.Value)), inputCoins[i].CoinDetails.SNDerivator, randInputSumwithoutShardID[i]},
			[]byte{privacy.SK, privacy.VALUE, privacy.SND, privacy.RAND})

		/***** Build witness for proving one-out-of-N commitments is a commitment to the coins being spent *****/
		// commitmentTemps is a list of commitments for protocol one-out-of-N
		commitmentTemps[i] = make([]*privacy.EllipticPoint, privacy.CMRingSize)
		randInputSum[i].Add(randInputSum[i], randInputShardID)
		randInputSum[i].Mod(randInputSum[i], privacy.Curve.Params().N)

		rndInputIsZero[i] = big.NewInt(0)
		rndInputIsZero[i].Set(inputCoins[i].CoinDetails.Randomness)
		rndInputIsZero[i].Sub(rndInputIsZero[i], randInputSum[i])
		rndInputIsZero[i].Mod(rndInputIsZero[i], privacy.Curve.Params().N)

		cmInputSum[i].X, cmInputSum[i].Y = privacy.Curve.Add(cmInputSum[i].X, cmInputSum[i].Y, wit.ComInputShardID.X, wit.ComInputShardID.Y)
		cmInputSumInverse[i], _ = cmInputSum[i].Inverse()

		randInputSumAll.Add(randInputSumAll, randInputSum[i])
		randInputSumAll.Mod(randInputSumAll, privacy.Curve.Params().N)

		openingWitnessHien := new(PKComOpeningsWitness)
		openingWitnessHien.Set(cmInputSum[i],
			[]*big.Int{wit.spendingKey, new(big.Int).SetUint64(inputCoins[i].CoinDetails.Value), inputCoins[i].CoinDetails.SNDerivator, big.NewInt(int64(wit.pkLastByteSender)), randInputSum[i]},
			[]byte{privacy.SK, privacy.VALUE, privacy.SND, privacy.SHARDID, privacy.RAND})

		openingProofHien, _ := openingWitnessHien.Prove()
		fmt.Println(openingProofHien.Verify())

		for j := 0; j < privacy.CMRingSize; j++ {
			commitmentTemps[i][j] = new(privacy.EllipticPoint)
			commitmentTemps[i][j].X = big.NewInt(0)
			commitmentTemps[i][j].Y = big.NewInt(0)
			commitmentTemps[i][j].X, commitmentTemps[i][j].Y = privacy.Curve.Add(commitments[preIndex+j].X, commitments[preIndex+j].Y, cmInputSumInverse[i].X, cmInputSumInverse[i].Y)
		}

		if wit.OneOfManyWitness[i] == nil {
			wit.OneOfManyWitness[i] = new(PKOneOfManyWitness)
		}
		indexIsZero := myCommitmentIndexs[i] % privacy.CMRingSize

		wit.OneOfManyWitness[i].Set(commitmentTemps[i], commitmentIndexs[preIndex:preIndex+privacy.CMRingSize], rndInputIsZero[i], indexIsZero, privacy.SK)
		preIndex = privacy.CMRingSize * (i + 1)

		/***** Build witness for proving that serial number is derived from the committed derivator *****/
		if wit.EqualityOfCommittedValWitness[i] == nil {
			wit.EqualityOfCommittedValWitness[i] = new(PKEqualityOfCommittedValWitness)
		}
		wit.EqualityOfCommittedValWitness[i].Set([]*privacy.EllipticPoint{cmInputSNDIndexSK[i], wit.ComInputSND[i]}, indexZKPEqual, []*big.Int{inputCoins[i].CoinDetails.SNDerivator, randInputSNDIndexSK[i], randInputSND[i]})

		/****Build witness for proving that the commitment of serial number is equivalent to Mul(com(sk), com(snd))****/
		witnesssA := new(big.Int)
		witnesssA.Add(wit.spendingKey, inputCoins[i].CoinDetails.SNDerivator)
		randA := new(big.Int)
		randA.Add(randInputSK, randInputSNDIndexSK[i])
		witnessAInverse := new(big.Int)
		witnessAInverse.ModInverse(witnesssA, privacy.Curve.Params().N)

		cmInputInverseSum := privacy.PedCom.CommitAtIndex(witnessAInverse, new(big.Int).SetInt64(0), privacy.SK)
		witIndex := new(byte)
		*witIndex = privacy.SK
		if wit.ProductCommitmentWitness[i] == nil {
			wit.ProductCommitmentWitness[i] = new(PKComProductWitness)
		}
		wit.ProductCommitmentWitness[i].Set(witnesssA, randA, cmInputInverseSum, witIndex)
	}

	numOutputCoin := len(wit.outputCoins)

	randOutputValue := make([]*big.Int, numOutputCoin)
	randOutputSND := make([]*big.Int, numOutputCoin)
	cmOutputValue := make([]*privacy.EllipticPoint, numOutputCoin)
	cmOutputSND := make([]*privacy.EllipticPoint, numOutputCoin)

	cmOutputSum := make([]*privacy.EllipticPoint, numOutputCoin)
	cmOutputSumwithoutShardID := make([]*privacy.EllipticPoint, numOutputCoin)
	randOutputSum := make([]*big.Int, numOutputCoin)
	randOutputSumwithoutShardID := make([]*big.Int, numOutputCoin)

	cmOutputSumAll := new(privacy.EllipticPoint)
	cmOutputSumAll.X = big.NewInt(0)
	cmOutputSumAll.Y = big.NewInt(0)

	// cmOutputValueAll is sum of all value coin commitments
	cmOutputValueAll := privacy.EllipticPoint{big.NewInt(0), big.NewInt(0)}
	randOutputValueAll := big.NewInt(0)

	wit.ComOutputOpeningsWitness = make([]*PKComOpeningsWitness, numOutputCoin)

	for i, outputCoin := range wit.outputCoins {
		randOutputValue[i] = privacy.RandInt()
		randOutputSND[i] = privacy.RandInt()
		cmOutputValue[i] = privacy.PedCom.CommitAtIndex(new(big.Int).SetUint64(outputCoin.CoinDetails.Value), randOutputValue[i], privacy.VALUE)
		cmOutputSND[i] = privacy.PedCom.CommitAtIndex(outputCoin.CoinDetails.SNDerivator, randOutputSND[i], privacy.SND)

		randOutputSum[i] = big.NewInt(0)
		randOutputSum[i].Set(randOutputValue[i])
		randOutputSum[i].Add(randOutputSum[i], randOutputSND[i])
		randOutputSum[i].Mod(randOutputSum[i], privacy.Curve.Params().N)
		randOutputSumwithoutShardID[i] = big.NewInt(0)
		randOutputSumwithoutShardID[i].Set(randOutputSum[i])

		cmOutputSum[i] = new(privacy.EllipticPoint)
		cmOutputSum[i].X = big.NewInt(0)
		cmOutputSum[i].Y = big.NewInt(0)
		cmOutputSum[i].X.Set(cmOutputValue[i].X)
		cmOutputSum[i].Y.Set(cmOutputValue[i].Y)
		cmOutputSum[i].X, cmOutputSum[i].Y = privacy.Curve.Add(cmOutputSum[i].X, cmOutputSum[i].Y, cmOutputSND[i].X, cmOutputSND[i].Y)
		cmOutputSumwithoutShardID[i] = new(privacy.EllipticPoint)
		cmOutputSumwithoutShardID[i].X, cmOutputSumwithoutShardID[i].Y = big.NewInt(0), big.NewInt(0)
		cmOutputSumwithoutShardID[i].X.Set(cmOutputSum[i].X)
		cmOutputSumwithoutShardID[i].Y.Set(cmOutputSum[i].Y)
		cmOutputValueAll = *(cmOutputValueAll.Add(cmOutputValue[i]))
		randOutputValueAll.Add(randOutputValueAll, randOutputValue[i])

		/***** Build witness for proving the knowledge of output coins' Openings (value, snd, randomness) *****/
		if wit.ComOutputOpeningsWitness[i] == nil {
			wit.ComOutputOpeningsWitness[i] = new(PKComOpeningsWitness)
		}
		wit.ComOutputOpeningsWitness[i].Set(cmOutputSumwithoutShardID[i],
			[]*big.Int{big.NewInt(int64(outputCoins[i].CoinDetails.Value)), outputCoins[i].CoinDetails.SNDerivator, randOutputSumwithoutShardID[i]},
			[]byte{privacy.VALUE, privacy.SND, privacy.RAND})
	}

	randOutputShardID := make([]*big.Int, numOutputCoin)
	cmOutputShardID := make([]*privacy.EllipticPoint, numOutputCoin)

	for i := 0; i < numOutputCoin; i++ {
		// calculate final commitment for output coins
		randOutputShardID[i] = privacy.RandInt()
		cmOutputShardID[i] = privacy.PedCom.CommitAtIndex(big.NewInt(int64(outputCoins[i].CoinDetails.GetPubKeyLastByte())), randOutputShardID[i], privacy.SHARDID)

		cmOutputSum[i] = cmOutputSum[i].Add(outputCoins[i].CoinDetails.PublicKey)
		cmOutputSum[i] = cmOutputSum[i].Add(cmOutputShardID[i])

		randOutputSum[i].Add(randOutputSum[i], randOutputShardID[i])
		randOutputSum[i].Mod(randOutputSum[i], privacy.Curve.Params().N)

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
			return errors.New("output coin's value is less than 0")
		}
	}
	if wit.ComOutputMultiRangeWitness == nil {
		wit.ComOutputMultiRangeWitness = new(PKComMultiRangeWitness)
	}
	wit.ComOutputMultiRangeWitness.Set(outputValue, 64)
	// ------------------------

	// Build witness for proving Sum(Input's value) == Sum(Output's Value)
	if fee > 0 {
		cmOutputValueAll.Add(privacy.PedCom.G[privacy.VALUE].ScalarMul(big.NewInt(int64(fee))))
	}

	cmOutputValueAllInverse, err := cmOutputValueAll.Inverse()
	if err != nil {
		return err
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
func (wit *PaymentWitness) Prove(hasPrivacy bool) (*PaymentProof, error) {
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
			productCommitmentProof, err := wit.ProductCommitmentWitness[i].Prove()
			if err != nil {
				return nil, err
			}
			proof.ProductCommitmentProof = append(proof.ProductCommitmentProof, productCommitmentProof)
		}
		return proof, nil
	}

	// if hasPrivacy == true
	numInputCoins := len(wit.ComInputOpeningsWitness)
	numOutputCoins := len(wit.ComOutputOpeningsWitness)

	for i := 0; i < numInputCoins; i++ {
		// Proving the knowledge of input coins' Openings
		comInputOpeningsProof, err := wit.ComInputOpeningsWitness[i].Prove()
		if err != nil {
			return nil, err
		}
		proof.ComInputOpeningsProof = append(proof.ComInputOpeningsProof, comInputOpeningsProof)

		// Proving one-out-of-N commitments is a commitment to the coins being spent
		oneOfManyProof, err := wit.OneOfManyWitness[i].Prove()
		if err != nil {
			return nil, err
		}
		proof.OneOfManyProof = append(proof.OneOfManyProof, oneOfManyProof)

		// Proving that serial number is derived from the committed derivator
		equalityOfCommittedValProof := wit.EqualityOfCommittedValWitness[i].Prove()
		proof.EqualityOfCommittedValProof = append(proof.EqualityOfCommittedValProof, equalityOfCommittedValProof)

		productCommitmentProof, err := wit.ProductCommitmentWitness[i].Prove()
		if err != nil {
			return nil, err
		}
		proof.ProductCommitmentProof = append(proof.ProductCommitmentProof, productCommitmentProof)
	}

	// Proving the knowledge of output coins' openings
	for i := 0; i < numOutputCoins; i++ {
		// Proving the knowledge of output coins' openings
		comOutputOpeningsProof, err := wit.ComOutputOpeningsWitness[i].Prove()
		if err != nil {
			return nil, err
		}
		proof.ComOutputOpeningsProof = append(proof.ComOutputOpeningsProof, comOutputOpeningsProof)
	}

	// Proving that each output values and sum of them does not exceed v_max
	proof.ComOutputMultiRangeProof, err = wit.ComOutputMultiRangeWitness.Prove()
	if err != nil {
		return nil, err
	}
	proof.SumOutRangeProof, err = wit.ComOutputMultiRangeWitness.ProveSum()
	if err != nil {
		return nil, err
	}
	// privacy-protocol/zero-knowledge/zkp_opening.go
	// Proving that sum of all input values is equal to sum of all output values
	proof.ComZeroProof, err = wit.ComZeroWitness.Prove()
	if err != nil {
		return nil, err
	}

	//Calculate new coin commitment

	fmt.Println("Privacy log: PROVING DONE!!!")
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
			if !pro.ProductCommitmentProof[i].Verify() {
				return false
			}

			pubKeyLastByteSender := pubKey[len(pubKey)-1]

			// Check input coins' cm is calculated correctly
			cmTmp := pro.InputCoins[i].CoinDetails.PublicKey
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.VALUE].ScalarMul(big.NewInt(int64(pro.InputCoins[i].CoinDetails.Value))))
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SND].ScalarMul(pro.InputCoins[i].CoinDetails.SNDerivator))
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SHARDID].ScalarMul(new(big.Int).SetBytes([]byte{pubKeyLastByteSender})))
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.RAND].ScalarMul(pro.InputCoins[i].CoinDetails.Randomness))
			if !cmTmp.IsEqual(pro.InputCoins[i].CoinDetails.CoinCommitment) {
				return false
			}

			// Calculate sum of input values
			sumInputValue += pro.InputCoins[i].CoinDetails.Value
		}

		for i := 0; i < len(pro.OutputCoins); i++ {
			// Check output coins' cm is calculated correctly
			cmTmp := pro.OutputCoins[i].CoinDetails.PublicKey
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.VALUE].ScalarMul(big.NewInt(int64(pro.OutputCoins[i].CoinDetails.Value))))
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SND].ScalarMul(pro.OutputCoins[i].CoinDetails.SNDerivator))
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SHARDID].ScalarMul(new(big.Int).SetBytes([]byte{pro.OutputCoins[i].CoinDetails.GetPubKeyLastByte()})))
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.RAND].ScalarMul(pro.OutputCoins[i].CoinDetails.Randomness))
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
	cmInputSum := make([]*privacy.EllipticPoint, len(pro.ComInputOpeningsProof))
	cmInputSumInverse := make([]*privacy.EllipticPoint, len(pro.ComInputOpeningsProof))
	for i := 0; i < len(pro.ComInputOpeningsProof); i++ {
		// Verify the proof for knowledge of input coins' Openings
		if !pro.ComInputOpeningsProof[i].Verify() {
			return false
		}
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

		// get commitments list from CommitmentIndexs
		commitments := make([]*privacy.EllipticPoint, privacy.CMRingSize)
		for j := 0; j < privacy.CMRingSize; j++ {
			commitmentBytes, err := db.GetCommitmentByIndex(tokenID, pro.OneOfManyProof[i].CommitmentIndexs[j], chainId)
			if err != nil {
				fmt.Printf("Error when verify: %v\n", err)
				return false
			}
			commitments[j], err = privacy.DecompressKey(commitmentBytes)
			if err != nil {
				fmt.Printf("Error when verify: %v\n", err)
				return false
			}

			commitments[j] = commitments[j].Add(cmInputSumInverse[i])
		}

		pro.OneOfManyProof[i].Commitments = commitments

		if !pro.OneOfManyProof[i].Verify() {
			return false
		}
		// Verify for the Proof that input coins' serial number is derived from the committed derivator
		if !pro.EqualityOfCommittedValProof[i].Verify() {
			return false
		}
		if !pro.ProductCommitmentProof[i].Verify() {
			return false
		}
	}

	// Verify the proof for knowledge of output coins' openings
	for i := 0; i < len(pro.ComOutputOpeningsProof); i++ {
		if !pro.ComOutputOpeningsProof[i].Verify() {
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
