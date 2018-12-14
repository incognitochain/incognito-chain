package zkp

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy-protocol"
)

// PaymentWitness contains all of witness for proving when spending coins
type PaymentWitness struct {
	spendingKey        *big.Int
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

	PubKeyLastByteSender byte
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
		PubKeyLastByteSender:        byte(0x00),
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
	fmt.Printf("Byte - lenComInputOpeningsProofArray: %v\n", lenComInputOpeningsProofArray)

	for i := 0; i < len(paymentProof.ComInputOpeningsProof); i++ {
		comInputOpeningProof := paymentProof.ComInputOpeningsProof[i].Bytes()
		fmt.Printf("Byte - lenComInputOpeningsProof: %v\n", len(comInputOpeningProof))
		//fmt.Printf("Byte - ComInputOpeningsProof: %v\n", comInputOpeningProof)

		proofbytes = append(proofbytes, byte(len(comInputOpeningProof)))
		proofbytes = append(proofbytes, comInputOpeningProof...)
	}
	// OneOfManyProof
	lenOneOfManyProofArray := len(paymentProof.OneOfManyProof)
	proofbytes = append(proofbytes, byte(lenOneOfManyProofArray))
	fmt.Printf("Byte - lenOneOfManyProofArray: %v\n", lenOneOfManyProofArray)

	for i := 0; i < len(paymentProof.OneOfManyProof); i++ {
		oneOfManyProof := paymentProof.OneOfManyProof[i].Bytes()

		fmt.Printf("Byte - lenOneOfManyProof: %v\n", len(oneOfManyProof))
		//fmt.Printf("Byte - OneOfManyProof: %v\n", oneOfManyProof)

		proofbytes = append(proofbytes, privacy.IntToByteArr(len(oneOfManyProof))...)
		proofbytes = append(proofbytes, oneOfManyProof...)
	}
	// EqualityOfCommittedValProof
	lenEqualityOfCommittedValProofArray := len(paymentProof.EqualityOfCommittedValProof)
	proofbytes = append(proofbytes, byte(lenEqualityOfCommittedValProofArray))
	fmt.Printf("Byte - lenEqualityOfCommittedValProofArray: %v\n", lenEqualityOfCommittedValProofArray)

	for i := 0; i < len(paymentProof.EqualityOfCommittedValProof); i++ {
		equalityOfCommittedValProof := paymentProof.EqualityOfCommittedValProof[i].Bytes()
		fmt.Printf("Byte - LenEqualityOfCommittedValProof: %v\n", len(equalityOfCommittedValProof))
		//fmt.Printf("Byte - equalityOfCommittedValProof: %v\n", equalityOfCommittedValProof)

		proofbytes = append(proofbytes, byte(len(equalityOfCommittedValProof)))
		proofbytes = append(proofbytes, equalityOfCommittedValProof...)
	}
	// ProductCommitmentProof
	proofbytes = append(proofbytes, byte(len(paymentProof.ProductCommitmentProof)))
	fmt.Printf("Byte - lenProductCommitmentProofArray: %v\n", len(paymentProof.ProductCommitmentProof))
	for i := 0; i < len(paymentProof.ProductCommitmentProof); i++ {
		productCommitmentProof := paymentProof.ProductCommitmentProof[i].Bytes()
		proofbytes = append(proofbytes, byte(len(productCommitmentProof)))
		proofbytes = append(proofbytes, paymentProof.ProductCommitmentProof[i].Bytes()...)
		fmt.Printf("Byte - LenEqualityOfCommittedValProof: %v\n", len(productCommitmentProof))

	}
	//ComOutputOpeningsProof
	proofbytes = append(proofbytes, byte(len(paymentProof.ComOutputOpeningsProof)))
	fmt.Printf("Byte - lenComOutputOpeningsProofArray: %v\n", len(paymentProof.ComOutputOpeningsProof))
	for i := 0; i < len(paymentProof.ComOutputOpeningsProof); i++ {
		comOutputOpeningsProof := paymentProof.ComOutputOpeningsProof[i].Bytes()
		fmt.Printf("Byte - LencomOutputOpeningsProof: %v\n", len(comOutputOpeningsProof))
		//fmt.Printf("Byte - comOutputOpeningsProof: %v\n", comOutputOpeningsProof)

		proofbytes = append(proofbytes, byte(len(comOutputOpeningsProof)))
		proofbytes = append(proofbytes, comOutputOpeningsProof...)
	}
	// ComOutputMultiRangeProof
	if paymentProof.ComOutputMultiRangeProof != nil {
		comOutputMultiRangeProof := paymentProof.ComOutputMultiRangeProof.Bytes()
		fmt.Printf("Byte - LencomOutputMultiRangeProof: %v\n", len(comOutputMultiRangeProof))
		//fmt.Printf("Byte - comOutputMultiRangeProof: %v\n", comOutputMultiRangeProof)

		proofbytes = append(proofbytes, privacy.IntToByteArr(len(comOutputMultiRangeProof))...)
		proofbytes = append(proofbytes, comOutputMultiRangeProof...)

		fmt.Println("DATTTTTTTTTTTTTTTTTTTT - BYTE: ", comOutputMultiRangeProof)
	} else {
		proofbytes = append(proofbytes, byte(0))
	}

	// SumOutRangeProof
	//paymentProof.SumOutRangeProof = nil
	if paymentProof.SumOutRangeProof != nil {
		sumOutRangeProof := paymentProof.SumOutRangeProof.Bytes()
		fmt.Printf("Byte - LensumOutRangeProof: %v\n", len(sumOutRangeProof))
		//fmt.Printf("Byte - sumOutRangeProof: %v\n", sumOutRangeProof)
		proofbytes = append(proofbytes, byte(len(sumOutRangeProof)))
		proofbytes = append(proofbytes, sumOutRangeProof...)

		fmt.Println("HYYYYYYYYYYYYYYYYYYYY - BYTE: ", sumOutRangeProof)
	} else {
		proofbytes = append(proofbytes, byte(0))
	}

	// ComZeroProof
	//paymentProof.ComZeroProof = nil
	if paymentProof.ComZeroProof != nil {
		comZeroProof := paymentProof.ComZeroProof.Bytes()
		fmt.Printf("Byte - LencomZeroProof: %v\n", len(comZeroProof))
		//fmt.Printf("Byte - comZeroProof: %v\n", comZeroProof)
		proofbytes = append(proofbytes, byte(len(comZeroProof)))
		proofbytes = append(proofbytes, comZeroProof...)
	} else {
		proofbytes = append(proofbytes, byte(0))
	}

	// InputCoins
	proofbytes = append(proofbytes, byte(len(paymentProof.InputCoins)))
	fmt.Printf("Byte - LeninputCoinsArray: %v\n", len(paymentProof.InputCoins))

	for i := 0; i < len(paymentProof.InputCoins); i++ {
		inputCoins := paymentProof.InputCoins[i].Bytes()
		fmt.Printf("Byte - LeninputCoins: %v\n", len(inputCoins))
		proofbytes = append(proofbytes, byte(len(inputCoins)))
		proofbytes = append(proofbytes, inputCoins...)
	}
	// OutputCoins
	proofbytes = append(proofbytes, byte(len(paymentProof.OutputCoins)))
	fmt.Printf("Byte - LenoutputCoinsArray: %v\n", len(paymentProof.OutputCoins))
	for i := 0; i < len(paymentProof.OutputCoins); i++ {
		outputCoins := paymentProof.OutputCoins[i].Bytes()
		lenOutputCoins := len(outputCoins)
		//if lenOutputCoins != 222{
		//
		//}
		fmt.Printf("Byte - LenoutputCoins: %v\n", lenOutputCoins)
		//fmt.Printf("Byte - outputCoins: %v\n", outputCoins)
		proofbytes = append(proofbytes, byte(len(outputCoins)))
		proofbytes = append(proofbytes, outputCoins...)
	}
	// ComOutputValue
	proofbytes = append(proofbytes, byte(len(paymentProof.ComOutputValue)))
	fmt.Printf("Byte - LenComOutputValueArray: %v\n", len(paymentProof.ComOutputValue))
	for i := 0; i < len(paymentProof.ComOutputValue); i++ {
		comOutputValue := paymentProof.ComOutputValue[i].Compress()
		fmt.Printf("Byte - LenComOutputValue: %v\n", len(comOutputValue))
		//fmt.Printf("Byte - ComOutputValue: %v\n", comOutputValue)
		proofbytes = append(proofbytes, byte(len(comOutputValue)))
		proofbytes = append(proofbytes, comOutputValue...)
	}
	// ComOutputSND
	proofbytes = append(proofbytes, byte(len(paymentProof.ComOutputSND)))
	fmt.Printf("Byte - LenCComOutputSNDArray: %v\n", len(paymentProof.ComOutputSND))
	for i := 0; i < len(paymentProof.ComOutputSND); i++ {
		comOutputSND := paymentProof.ComOutputSND[i].Compress()
		fmt.Printf("Byte - LencomOutputSND: %v\n", len(comOutputSND))
		proofbytes = append(proofbytes, byte(len(comOutputSND)))
		proofbytes = append(proofbytes, comOutputSND...)
	}
	// ComOutputShardID
	proofbytes = append(proofbytes, byte(len(paymentProof.ComOutputShardID)))
	fmt.Printf("Byte - LenComOutputShardIDArray: %v\n", len(paymentProof.ComOutputShardID))
	for i := 0; i < len(paymentProof.ComOutputShardID); i++ {
		comOutputShardID := paymentProof.ComOutputShardID[i].Compress()
		fmt.Printf("Byte - LencomOutputShardID: %v\n", len(comOutputShardID))
		proofbytes = append(proofbytes, byte(len(comOutputShardID)))
		proofbytes = append(proofbytes, comOutputShardID...)
	}
	// PubKeyLastByteSender
	//proofbytes = append(proofbytes, byte(1))
	proofbytes = append(proofbytes, paymentProof.PubKeyLastByteSender)
	//fmt.Println("********************** LEN - BYTES ",len(proofbytes))
	fmt.Printf("***************BYTE - PROOF %v\n", proofbytes)
	return proofbytes
}

func (proof *PaymentProof) SetBytes(proofbytes []byte) (err error) {

	fmt.Printf("***************BEFORE SETBYTE - PROOF %v\n", proofbytes)
	//proof.Init()
	offset := 0
	// Set ComInputOpeningsProof
	lenComInputOpeningsProofArray := int(proofbytes[offset])
	fmt.Printf("Set Byte - lenComInputOpeningsProofArray: %v\n", lenComInputOpeningsProofArray)
	offset += 1
	proof.ComInputOpeningsProof = make([]*PKComOpeningsProof, lenComInputOpeningsProofArray)
	for i := 0; i < lenComInputOpeningsProofArray; i++ {
		lenComInputOpeningsProof := int(proofbytes[offset])
		offset += 1
		fmt.Printf("Set Byte - lenComInputOpeningsProof: %v\n", lenComInputOpeningsProof)
		//fmt.Printf("Set Byte - ComInputOpeningsProof: %v\n", proofbytes[offset: offset+lenComInputOpeningsProof])

		proof.ComInputOpeningsProof[i] = new(PKComOpeningsProof).Init()
		proof.ComInputOpeningsProof[i].SetBytes(proofbytes[offset : offset+lenComInputOpeningsProof])
		offset += lenComInputOpeningsProof
	}
	// Set OneOfManyProof
	lenOneOfManyProofArray := int(proofbytes[offset])
	offset += 1
	fmt.Printf("Set Byte - lenOneOfManyProofArray: %v\n", lenOneOfManyProofArray)
	proof.OneOfManyProof = make([]*PKOneOfManyProof, lenOneOfManyProofArray)
	for i := 0; i < lenOneOfManyProofArray; i++ {
		//fmt.Printf("%v - %v\n", proofbytes[offset: offset+2], privacy.ByteArrToInt(proofbytes[offset: offset+2]))
		lenOneOfManyProof := privacy.ByteArrToInt(proofbytes[offset : offset+2])
		offset += 2
		fmt.Printf("Set Byte - lenOneOfManyProof: %v\n", lenOneOfManyProof)
		//fmt.Printf("Set Byte - OneOfManyProof: %v\n", proofbytes[offset: offset+lenOneOfManyProof])

		proof.OneOfManyProof[i] = new(PKOneOfManyProof).Init()
		proof.OneOfManyProof[i].SetBytes(proofbytes[offset : offset+lenOneOfManyProof])
		offset += lenOneOfManyProof
	}
	// Set EqualityOfCommittedValProof
	lenEqualityOfCommittedValProofArray := int(proofbytes[offset])
	offset += 1
	fmt.Printf("Set Byte - lenEqualityOfCommittedValProofArray: %v\n", lenEqualityOfCommittedValProofArray)
	proof.EqualityOfCommittedValProof = make([]*PKEqualityOfCommittedValProof, lenEqualityOfCommittedValProofArray)
	for i := 0; i < lenEqualityOfCommittedValProofArray; i++ {
		lenEqualityOfCommittedValProof := int(proofbytes[offset])
		offset += 1
		fmt.Printf("Set Byte - lenEqualityOfCommittedValProof: %v\n", lenEqualityOfCommittedValProof)
		//fmt.Printf("Set Byte - EqualityOfCommittedValProof: %v\n", proofbytes[offset: offset+lenEqualityOfCommittedValProof])

		proof.EqualityOfCommittedValProof[i] = new(PKEqualityOfCommittedValProof).Init()
		err := proof.EqualityOfCommittedValProof[i].SetBytes(proofbytes[offset : offset+lenEqualityOfCommittedValProof])
		if err != nil {
			return err
		}
		offset += lenEqualityOfCommittedValProof
	}
	// Set ProductCommitmentProof
	lenProductCommitmentProofArray := int(proofbytes[offset])
	offset += 1
	fmt.Printf("Set Byte - lenProductCommitmentProofArray: %v\n", lenProductCommitmentProofArray)
	proof.ProductCommitmentProof = make([]*PKComProductProof, lenProductCommitmentProofArray)
	for i := 0; i < lenProductCommitmentProofArray; i++ {
		lenProductCommitmentProof := int(proofbytes[offset])
		offset += 1

		fmt.Printf("Set Byte - lenProductCommitmentProof: %v\n", lenProductCommitmentProof)
		//fmt.Printf("Set Byte - ProductCommitmentProof: %v\n", proofbytes[offset: offset+lenProductCommitmentProof])
		proof.ProductCommitmentProof[i] = new(PKComProductProof).Init()
		proof.ProductCommitmentProof[i].SetBytes(proofbytes[offset : offset+lenProductCommitmentProof])
		offset += lenProductCommitmentProof
	}
	//Set ComOutputOpeningsProof
	lenComOutputOpeningsProofArray := int(proofbytes[offset])
	offset += 1
	fmt.Printf("Set Byte - lenComOutputOpeningsProofArray: %v\n", lenComOutputOpeningsProofArray)
	proof.ComOutputOpeningsProof = make([]*PKComOpeningsProof, lenComOutputOpeningsProofArray)
	for i := 0; i < lenComOutputOpeningsProofArray; i++ {
		lenComOutputOpeningsProof := int(proofbytes[offset])
		offset += 1

		fmt.Printf("Set Byte - lenComOutputOpeningsProof: %v\n", lenComOutputOpeningsProof)
		//fmt.Printf("Set Byte - ComOutputOpeningsProof: %v\n", proofbytes[offset: offset+lenComOutputOpeningsProof])
		proof.ComOutputOpeningsProof[i] = new(PKComOpeningsProof).Init()
		proof.ComOutputOpeningsProof[i].SetBytes(proofbytes[offset : offset+lenComOutputOpeningsProof])
		offset += lenComOutputOpeningsProof
	}

	//ComOutputMultiRangeProof *PKComMultiRangeProof
	lenComOutputMultiRangeProof := privacy.ByteArrToInt(proofbytes[offset : offset+2])
	offset += 2
	fmt.Printf("Set Byte - lenComOutputMultiRangeProof: %v\n", lenComOutputMultiRangeProof)
	if lenComOutputMultiRangeProof > 0 {
		//fmt.Printf("Set Byte - ComOutputMultiRangeProof: %v\n", proofbytes[offset: offset+lenComOutputMultiRangeProof])
		proof.ComOutputMultiRangeProof = new(PKComMultiRangeProof).Init()
		proof.ComOutputMultiRangeProof.SetBytes(proofbytes[offset : offset+lenComOutputMultiRangeProof])
		offset += lenComOutputMultiRangeProof
		fmt.Println("DATTTTTTTTTTTTTTTTTTTT - SET BYTE: ", proof.ComOutputMultiRangeProof.Bytes())
	}

	//SumOutRangeProof *PKComZeroProof
	lenSumOutRangeProof := int(proofbytes[offset])
	offset += 1
	fmt.Printf("Set Byte - lenSumOutRangeProof: %v\n", lenSumOutRangeProof)
	if lenSumOutRangeProof > 0 {
		//fmt.Printf("Set Byte - SumOutRangeProof: %v\n", proofbytes[offset: offset+lenSumOutRangeProof])
		proof.SumOutRangeProof = new(PKComZeroProof).Init()
		proof.SumOutRangeProof.SetBytes(proofbytes[offset : offset+lenSumOutRangeProof])
		offset += lenSumOutRangeProof

		fmt.Println("HYYYYYYYYYYYYYYYYYYYY - SET BYTE: ", proof.SumOutRangeProof.Bytes())
	}

	//ComZeroProof *PKComZeroProof
	lenComZeroProof := int(proofbytes[offset])
	offset += 1
	fmt.Printf("Set Byte - lenComZeroProof: %v\n", lenComZeroProof)
	if lenComZeroProof > 0 {
		//fmt.Printf("Set Byte - ComZeroProof: %v\n", proofbytes[offset: offset+lenComZeroProof])
		proof.ComZeroProof = new(PKComZeroProof).Init()
		proof.ComZeroProof.SetBytes(proofbytes[offset : offset+lenComZeroProof])
		offset += lenComZeroProof
	}

	//InputCoins  []*privacy.InputCoin
	lenInputCoinsArray := int(proofbytes[offset])
	fmt.Printf("Set Byte - lenInputCoinsArray: %v\n", lenInputCoinsArray)
	offset += 1
	proof.InputCoins = make([]*privacy.InputCoin, lenInputCoinsArray)
	for i := 0; i < lenInputCoinsArray; i++ {
		lenInputCoin := int(proofbytes[offset])
		fmt.Printf("Set Byte - lenInputCoin: %v\n", lenInputCoin)
		offset += 1
		proof.InputCoins[i] = new(privacy.InputCoin)
		proof.InputCoins[i].SetBytes(proofbytes[offset : offset+lenInputCoin])
		offset += lenInputCoin
	}
	//OutputCoins []*privacy.OutputCoin
	lenOutputCoinsArray := int(proofbytes[offset])
	offset += 1
	fmt.Printf("Set Byte - lenOutputCoinsArray: %v\n", lenOutputCoinsArray)
	proof.OutputCoins = make([]*privacy.OutputCoin, lenOutputCoinsArray)
	for i := 0; i < lenOutputCoinsArray; i++ {
		lenOutputCoin := int(proofbytes[offset])
		offset += 1
		fmt.Printf("Set Byte - lenOutputCoin: %v\n", lenOutputCoin)
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
	fmt.Printf("Set Byte - lenComOutputValueArray: %v\n", lenComOutputValueArray)
	proof.ComOutputValue = make([]*privacy.EllipticPoint, lenComOutputValueArray)
	for i := 0; i < lenComOutputValueArray; i++ {
		lenComOutputValue := int(proofbytes[offset])
		offset += 1
		fmt.Printf("Set Byte - lenComOutputValue: %v\n", lenComOutputValue)
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
	fmt.Printf("Set Byte - lenComOutputSNDArray: %v\n", lenComOutputSNDArray)
	proof.ComOutputSND = make([]*privacy.EllipticPoint, lenComOutputSNDArray)
	for i := 0; i < lenComOutputSNDArray; i++ {
		lenComOutputSND := int(proofbytes[offset])
		offset += 1
		fmt.Printf("Set Byte - lenComOutputSND: %v\n", lenComOutputSND)
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
	fmt.Printf("Set Byte - lenComOutputShardIdArray: %v\n", lenComOutputShardIdArray)
	proof.ComOutputShardID = make([]*privacy.EllipticPoint, lenComOutputShardIdArray)
	for i := 0; i < lenComOutputShardIdArray; i++ {
		lenComOutputShardId := int(proofbytes[offset])
		offset += 1
		fmt.Printf("Set Byte - lenComOutputShardId: %v\n", lenComOutputShardId)
		proof.ComOutputShardID[i] = new(privacy.EllipticPoint)
		proof.ComOutputShardID[i], err = privacy.DecompressKey(proofbytes[offset : offset+lenComOutputShardId])
		if err != nil {
			return err
		}
		offset += lenComOutputShardId
	}
	//PubKeyLastByteSender byte
	proof.PubKeyLastByteSender = proofbytes[offset]
	//fmt.Println("***********-----------LEN - SET BYTES ",len(proof.Bytes()))
	newBytes := proof.Bytes()
	fmt.Printf("***************AFTER SETBYTE - length %v - PROOF %v\n", len(newBytes), newBytes)
	return nil
}

// END----------------------------------------------------------------------------------------------------------------------------------------------

func (wit *PaymentWitness) Set(spendingKey *big.Int, inputCoins []*privacy.InputCoin, outputCoins []*privacy.OutputCoin) {
	wit.spendingKey = spendingKey
	wit.inputCoins = inputCoins
	wit.outputCoins = outputCoins
}

// Build prepares witnesses for all protocol need to be proved when create tx
// if hashPrivacy = false, witness includes spending key, input coins, output coins
// otherwise, witness includes all attributes in PaymentWitness struct
func (wit *PaymentWitness) Build(hasPrivacy bool,
	spendingKey *big.Int,
	inputCoins []*privacy.InputCoin, outputCoins []*privacy.OutputCoin,
	pkLastByteSender byte, pkLastByteReceivers []byte,
	commitments []*privacy.EllipticPoint, commitmentIndexs []uint64, myCommitmentIndexs []uint64,
	fee uint64) (err error) {

	wit.spendingKey = spendingKey
	wit.inputCoins = inputCoins
	wit.outputCoins = outputCoins
	wit.commitmentIndexs = commitmentIndexs
	wit.myCommitmentIndexs = myCommitmentIndexs

	numInputCoin := len(wit.inputCoins)

	randInputSK := privacy.RandInt()
	cmInputSK := privacy.PedCom.CommitAtIndex(wit.spendingKey, randInputSK, privacy.SK)

	randInputShardID := privacy.RandInt()
	cmInputShardID := privacy.PedCom.CommitAtIndex(big.NewInt(int64(wit.pkLastByteSender)), randInputShardID, privacy.SHARDID)

	cmInputValue := make([]*privacy.EllipticPoint, numInputCoin)
	cmInputSND := make([]*privacy.EllipticPoint, numInputCoin)
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

		cmInputValue[i] = privacy.PedCom.CommitAtIndex(big.NewInt(int64(inputCoin.CoinDetails.Value)), randInputValue[i], privacy.VALUE)
		cmInputSND[i] = privacy.PedCom.CommitAtIndex(inputCoin.CoinDetails.SNDerivator, randInputSND[i], privacy.SND)
		cmInputSNDIndexSK[i] = privacy.PedCom.CommitAtIndex(inputCoin.CoinDetails.SNDerivator, randInputSNDIndexSK[i], privacy.SK)

		cmInputValueAll.Add(cmInputValue[i])
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

	wit.ComInputOpeningsWitness = make([]*PKComOpeningsWitness, numInputCoin)
	wit.OneOfManyWitness = make([]*PKOneOfManyWitness, numInputCoin)
	wit.ProductCommitmentWitness = make([]*PKComProductWitness, numInputCoin)
	wit.EqualityOfCommittedValWitness = make([]*PKEqualityOfCommittedValWitness, numInputCoin)
	indexZKPEqual := make([]byte, 2)
	indexZKPEqual[0] = privacy.SK
	indexZKPEqual[1] = privacy.SND

	preIndex := 0
	for i := 0; i < numInputCoin; i++ {
		/***** Build witness for proving the knowledge of input coins' Openings  *****/
		cmInputSum[i] = cmInputSK
		cmInputSum[i].X, cmInputSum[i].Y = privacy.Curve.Add(cmInputSum[i].X, cmInputSum[i].Y, cmInputValue[i].X, cmInputValue[i].Y)
		cmInputSum[i].X, cmInputSum[i].Y = privacy.Curve.Add(cmInputSum[i].X, cmInputSum[i].Y, cmInputSND[i].X, cmInputSND[i].Y)

		randInputSum[i] = randInputSK
		randInputSum[i].Add(randInputSum[i], randInputValue[i])
		randInputSum[i].Add(randInputSum[i], randInputSND[i])
		randInputSum[i].Mod(randInputSum[i], privacy.Curve.Params().N)

		if wit.ComInputOpeningsWitness[i] == nil {
			wit.ComInputOpeningsWitness[i] = new(PKComOpeningsWitness)
		}
		wit.ComInputOpeningsWitness[i].Set(cmInputSum[i],
			[]*big.Int{wit.spendingKey, big.NewInt(int64(inputCoins[i].CoinDetails.Value)), inputCoins[i].CoinDetails.SNDerivator, randInputSum[i]},
			[]byte{privacy.SK, privacy.VALUE, privacy.SND, privacy.RAND})

		/***** Build witness for proving one-out-of-N commitments is a commitment to the coins being spent *****/
		// commitmentTemps is a list of commitments for protocol one-out-of-N
		commitmentTemps := make([]*privacy.EllipticPoint, privacy.CMRingSize)
		randInputSum[i].Add(randInputSum[i], randInputShardID)
		rndInputIsZero := big.NewInt(0).Sub(inputCoins[i].CoinDetails.Randomness, randInputSum[i])
		rndInputIsZero.Mod(rndInputIsZero, privacy.Curve.Params().N)

		cmInputSum[i].X, cmInputSum[i].Y = privacy.Curve.Add(cmInputSum[i].X, cmInputSum[i].Y, cmInputShardID.X, cmInputShardID.Y)
		cmInputSumInverse[i], _ = cmInputSum[i].Inverse()

		randInputSumAll.Add(randInputSumAll, randInputSum[i])
		randInputSumAll.Mod(randInputSumAll, privacy.Curve.Params().N)

		//cmInputSumAll.X, cmInputSumAll.Y = privacy.Curve.Add(cmInputSum[i].X, cmInputSum[i].Y, cmInputSumAll.X, cmInputSumAll.Y)

		for j := 0; j < privacy.CMRingSize; j++ {
			commitmentTemps[j] = new(privacy.EllipticPoint)
			commitmentTemps[j].X = big.NewInt(0)
			commitmentTemps[j].Y = big.NewInt(0)
			commitmentTemps[j].X, commitmentTemps[j].Y = privacy.Curve.Add(commitments[preIndex+j].X, commitments[preIndex+j].Y, cmInputSumInverse[i].X, cmInputSumInverse[i].Y)
		}
		preIndex = privacy.CMRingSize * (i + 1)

		if wit.OneOfManyWitness[i] == nil {
			wit.OneOfManyWitness[i] = new(PKOneOfManyWitness)
		}
		wit.OneOfManyWitness[i].Set(commitmentTemps, commitmentIndexs, rndInputIsZero, myCommitmentIndexs[i], privacy.SK)

		/***** Build witness for proving that serial number is derived from the committed derivator *****/
		if wit.EqualityOfCommittedValWitness[i] == nil {
			wit.EqualityOfCommittedValWitness[i] = new(PKEqualityOfCommittedValWitness)
		}
		wit.EqualityOfCommittedValWitness[i].Set([]*privacy.EllipticPoint{cmInputSNDIndexSK[i], cmInputSND[i]}, indexZKPEqual, []*big.Int{inputCoins[i].CoinDetails.SNDerivator, randInputSK, randInputSND[i]})

		/****Build witness for proving that the commitment of serial number is equivalent to Mul(com(sk), com(snd))****/
		witnesssA := new(big.Int)
		witnesssA.Add(wit.spendingKey, inputCoins[i].CoinDetails.SNDerivator)
		randA := new(big.Int)
		randA.Add(randInputSK, randInputSND[i])
		witnessAInverse := new(big.Int)
		witnessAInverse.ModInverse(witnesssA, privacy.Curve.Params().N)
		randAInverse := privacy.RandInt()
		cmInputInverseSum := privacy.PedCom.CommitAtIndex(witnessAInverse, randAInverse, privacy.SK)
		witIndex := new(byte)
		*witIndex = privacy.SK
		if wit.ProductCommitmentWitness[i] == nil {
			wit.ProductCommitmentWitness[i] = new(PKComProductWitness)
		}
		wit.ProductCommitmentWitness[i].Set(witnesssA, randA, cmInputInverseSum, witIndex)
		// ------------------------------
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

	wit.ComOutputOpeningsWitness = make([]*PKComOpeningsWitness, numOutputCoin)

	for i, outputCoin := range wit.outputCoins {
		randOutputValue[i] = privacy.RandInt()
		randOutputSND[i] = privacy.RandInt()
		cmOutputValue[i] = privacy.PedCom.CommitAtIndex(big.NewInt(int64(outputCoin.CoinDetails.Value)), randOutputValue[i], privacy.VALUE)
		cmOutputSND[i] = privacy.PedCom.CommitAtIndex(outputCoin.CoinDetails.SNDerivator, randOutputSND[i], privacy.SND)

		randOutputSum[i] = randOutputValue[i]
		randOutputSum[i].Add(randOutputSum[i], randOutputSND[i])
		randOutputSum[i].Mod(randOutputSum[i], privacy.Curve.Params().N)

		cmOutputSum[i] = new(privacy.EllipticPoint)
		cmOutputSum[i].X, cmOutputSum[i].Y = cmOutputValue[i].X, cmOutputValue[i].Y
		cmOutputSum[i].X, cmOutputSum[i].Y = privacy.Curve.Add(cmOutputSum[i].X, cmOutputSum[i].Y, cmOutputSND[i].X, cmOutputSND[i].Y)

		cmOutputValueAll = *(cmOutputValueAll.Add(cmOutputValue[i]))
		randOutputValueAll.Add(randOutputValueAll, randOutputValue[i])

		/***** Build witness for proving the knowledge of output coins' Openings (value, snd, randomness) *****/
		if wit.ComOutputOpeningsWitness[i] == nil {
			wit.ComOutputOpeningsWitness[i] = new(PKComOpeningsWitness)
		}
		wit.ComOutputOpeningsWitness[i].Set(cmOutputSum[i],
			[]*big.Int{big.NewInt(int64(outputCoins[i].CoinDetails.Value)), outputCoins[i].CoinDetails.SNDerivator, randOutputSum[i]},
			[]byte{privacy.VALUE, privacy.SND, privacy.RAND})
	}

	randOutputShardID := make([]*big.Int, numOutputCoin)
	cmOutputShardID := make([]*privacy.EllipticPoint, numOutputCoin)

	for i := 0; i < numOutputCoin; i++ {
		// calculate final commitment for output coins
		randOutputShardID[i] = privacy.RandInt()
		cmOutputShardID[i] = privacy.PedCom.CommitAtIndex(big.NewInt(int64(outputCoins[i].CoinDetails.GetPubKeyLastByte())), randOutputShardID[i], privacy.SHARDID)

		cmOutputSum[i].Add(outputCoins[i].CoinDetails.PublicKey)
		cmOutputSum[i].Add(cmOutputShardID[i])

		randOutputSum[i].Add(randOutputSum[i], randOutputShardID[i])
		randOutputSum[i].Mod(randOutputSum[i], privacy.Curve.Params().N)

		outputCoins[i].CoinDetails.CoinCommitment = cmOutputSum[i]
		outputCoins[i].CoinDetails.Randomness = randOutputSum[i]

		cmOutputSumAll.X, cmOutputSumAll.Y = privacy.Curve.Add(cmOutputSum[i].X, cmOutputSum[i].Y, cmOutputSumAll.X, cmOutputSumAll.Y)
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

	//printf output's value to test
	for i, outCoin := range wit.outputCoins {
		fmt.Printf("Output coin's value %v: %v\n", i, outCoin.CoinDetails.Value)
	}

	// if hasPrivacy == false, don't need to create the zero knowledge proof
	// proving user has spending key corresponding with public key in input coins
	// is proved by signing with spending key
	if !hasPrivacy {
		// Proving that serial number is derived from the committed derivator
		for i := 0; i < len(wit.inputCoins); i++ {
			equalityOfCommittedValProof := wit.EqualityOfCommittedValWitness[i].Prove()
			proof.EqualityOfCommittedValProof = append(proof.EqualityOfCommittedValProof, equalityOfCommittedValProof)

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

	// Proving that sum of all input values is equal to sum of all output values
	proof.ComZeroProof, err = wit.ComZeroWitness.Prove()
	if err != nil {
		return nil, err
	}

	fmt.Println("PROVING DONE!!!")
	return proof, nil
}

func (pro PaymentProof) Verify(hasPrivacy bool, pubKey privacy.PublicKey, db database.DatabaseInterface, chainId byte) bool {
	return true
	// has no privacy
	if !hasPrivacy {
		var sumInputValue, sumOutputValue uint64
		sumInputValue = 0
		sumOutputValue = 0

		for i := 0; i < len(pro.InputCoins); i++ {
			// Check input coins' Serial number is created from input coins' SND and sender's spending key
			// Todo: check
			if !pro.EqualityOfCommittedValProof[i].Verify() {
				return false
			}
			if !pro.ProductCommitmentProof[i].Verify() {
				return false
			}

			// Check input coins' cm is calculated correctly
			cmTmp := pro.InputCoins[i].CoinDetails.PublicKey
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.VALUE].ScalarMul(big.NewInt(int64(pro.InputCoins[i].CoinDetails.Value))))
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SND].ScalarMul(pro.InputCoins[i].CoinDetails.SNDerivator))
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SHARDID].ScalarMul(new(big.Int).SetBytes([]byte{pro.PubKeyLastByteSender})))
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
	for i := 0; i < len(pro.ComInputOpeningsProof); i++ {
		// Verify the proof for knowledge of input coins' Openings
		if !pro.ComInputOpeningsProof[i].Verify() {
			return false
		}
		// Verify for the proof one-out-of-N commitments is a commitment to the coins being spent
		if !pro.OneOfManyProof[i].Verify(db, chainId) {
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
