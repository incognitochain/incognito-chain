package zkp

import (
	"math/big"
	"sort"

	"github.com/ninjadotorg/constant/privacy-protocol"
)

// PaymentWitness contains all of witness for proving when spending coins
type PaymentWitness struct {
	spendingKey 	*big.Int
	inputCoins  	[]*privacy.InputCoin
	outputCoins 	[]*privacy.OutputCoin
	commitments 	[]*privacy.EllipticPoint
	randCmIndices []*privacy.CMIndex
	myCmPos 			[]uint32


	pkLastByteSender    byte
	pkLastByteReceivers []byte

	ComInputOpeningsWitness       []*PKComOpeningsWitness
	OneOfManyWitness              []*PKOneOfManyWitness
	EqualityOfCommittedValWitness []*PKEqualityOfCommittedValWitness
	ProductCommitmentWitness      []*PKComProductWitness

	ComOutputOpeningsWitness   []*PKComOpeningsWitness
	ComOutputMultiRangeWitness *PKComMultiRangeWitness
	SumOutRangeWitness         *PKComMultiRangeWitness

	ComZeroWitness *PKComZeroWitness
	//ComZeroOneWitness             *PKComZeroOneWitness
}

// PaymentProof contains all of PoK for spending coin
type PaymentProof struct {
	// for input coins
	ComInputOpeningsProof       []*PKComOpeningsProof            //flag -1,-2
	OneOfManyProof              []*PKOneOfManyProof              //flag -3,-4
	EqualityOfCommittedValProof []*PKEqualityOfCommittedValProof //flag -5,-6
	ProductCommitmentProof      []*PKComProductProof             //flag -7,-8
	// for output coins
	ComOutputOpeningsProof   []*PKComOpeningsProof //flag -9,-10
	ComOutputMultiRangeProof *PKComMultiRangeProof //flag -11,-12
	SumOutRangeProof         *PKComMultiRangeProof //flag -13,-14
	// for input = output
	ComZeroProof *PKComZeroProof //flag -15,-16
	// add list input coins' SN to proof for serial number
	// these following attributes just exist when tx doesn't have privacy
	OutputCoins []*privacy.OutputCoin
	InputCoins  []*privacy.InputCoin
}

func (paymentProof *PaymentProof) Bytes() []byte {
	var proofbytes []byte
	return proofbytes

	/*// OpeningsProof in total proof
	// len(openingProof)|| -1 || openingsProof1 || -1 || openingsProof2 || -1 || openingsProof3 || -1 || openingsProof4.....||-2
	var elementsFlag byte
	elementsFlag = -1
	var mainFlag byte
	mainFlag = -2
	proofbytes = append(proofbytes, byte(len(paymentProof.ComInputOpeningsProof)))
	for i := 0; i < len(paymentProof.ComInputOpeningsProof); i++ {
		proofbytes = append(proofbytes, elementsFlag)
		proofbytes = append(proofbytes, paymentProof.ComInputOpeningsProof[i].Bytes()...)
	}
	proofbytes = append(proofbytes, mainFlag)
	// OpeningsProof in total proof
	// len(openingProof)|| -1 || openingsProof1 || -1 || openingsProof2 || -1 || openingsProof3 || -1 || openingsProof4.....||-2
	// OneOfManyProof
	elementsFlag = -3
	mainFlag = -4
	proofbytes = append(proofbytes, byte(len(paymentProof.OneOfManyProof)))
	for i := 0; i < len(paymentProof.OneOfManyProof); i++ {
		proofbytes = append(proofbytes, elementsFlag)
		proofbytes = append(proofbytes, paymentProof.OneOfManyProof[i].Bytes()...)
	}
	proofbytes = append(proofbytes, mainFlag)
	// EqualityOfCommittedValProof
	elementsFlag = -5
	mainFlag = -6
	proofbytes = append(proofbytes, byte(len(paymentProof.EqualityOfCommittedValProof)))
	for i := 0; i < len(paymentProof.EqualityOfCommittedValProof); i++ {
		proofbytes = append(proofbytes, elementsFlag)
		proofbytes = append(proofbytes, paymentProof.EqualityOfCommittedValProof[i].Bytes()...)
	}
	proofbytes = append(proofbytes, mainFlag)
	// ProductCommitmentProof
	elementsFlag = -7
	mainFlag = -8
	proofbytes = append(proofbytes, byte(len(paymentProof.ProductCommitmentProof)))
	for i := 0; i < len(paymentProof.ProductCommitmentProof); i++ {
		proofbytes = append(proofbytes, elementsFlag)
		proofbytes = append(proofbytes, paymentProof.ProductCommitmentProof[i].Bytes()...)
	}
	proofbytes = append(proofbytes, mainFlag)
	// ProductCommitmentProof
	elementsFlag = -7
	mainFlag = -8
	proofbytes = append(proofbytes, byte(len(paymentProof.ProductCommitmentProof)))
	for i := 0; i < len(paymentProof.ProductCommitmentProof); i++ {
		proofbytes = append(proofbytes, elementsFlag)
		proofbytes = append(proofbytes, paymentProof.ProductCommitmentProof[i].Bytes()...)
	}
	proofbytes = append(proofbytes, mainFlag)
	// ComOutputOpeningsProof
	elementsFlag = -9
	mainFlag = -10
	proofbytes = append(proofbytes, byte(len(paymentProof.ComOutputOpeningsProof)))
	for i := 0; i < len(paymentProof.ComOutputOpeningsProof); i++ {
		proofbytes = append(proofbytes, elementsFlag)
		proofbytes = append(proofbytes, paymentProof.ComOutputOpeningsProof[i].Bytes()...)
	}
	proofbytes = append(proofbytes, mainFlag)
	// ComOutputMultiRangeProof
	elementsFlag = -11
	mainFlag = -12
	proofbytes = append(proofbytes, byte(len(paymentProof.ComOutputMultiRangeProof.Bytes())))
	proofbytes = append(proofbytes, elementsFlag)
	proofbytes = append(proofbytes, paymentProof.ComOutputMultiRangeProof.Bytes()...)
	proofbytes = append(proofbytes, mainFlag)
	// SumOutRangeProof
	elementsFlag = -13
	mainFlag = -14
	proofbytes = append(proofbytes, byte(len(paymentProof.SumOutRangeProof.Bytes())))
	proofbytes = append(proofbytes, elementsFlag)
	proofbytes = append(proofbytes, paymentProof.SumOutRangeProof.Bytes()...)
	proofbytes = append(proofbytes, mainFlag)
	// ComZeroProof
	elementsFlag = -15
	mainFlag = -16
	proofbytes = append(proofbytes, byte(len(paymentProof.SumOutRangeProof.Bytes())))
	proofbytes = append(proofbytes, elementsFlag)
	proofbytes = append(proofbytes, paymentProof.SumOutRangeProof.Bytes()...)
	proofbytes = append(proofbytes, mainFlag)
	return proofbytes*/
}

type PaymentProofByte struct {
	lenarrayComInputOpeningsProof       int
	lenarrayComOutputOpeningsProof      int
	lenarrayEqualityOfCommittedValProof int
	lenarrayOneOfManyProof              int

	//It should be constants
	lenComInputOpeningsProof       int
	lenComOutputOpeningsProof      int
	lenOneOfManyProof              int //
	lenEqualityOfCommittedValProof int
	lenComMultiRangeProof          int
	lenComZeroProof                int
	lenComZeroOneProof             int

	/**** proof ****/
	// for input coins
	ComInputOpeningsProof       []byte
	OneOfManyProof              []byte
	EqualityOfCommittedValProof []byte
	ProductCommitmentProof      []byte
	// for output coins
	ComOutputOpeningsProof   []byte
	ComOutputMultiRangeProof []byte
	SumOutRangeProof         []byte

	// for input = output
	ComZeroProof []byte
	//ComZeroOneProof    []byte
}

// Bytes converts payment proof to byte array to send verifiers
//func (paymentProof *PaymentProof) Bytes() []byte {
//	byteArray := new(PaymentProofByte)
//	byteArray.lenarrayComInputOpeningsProof = len(paymentProof.ComInputOpeningsProof)
//	byteArray.lenarrayComOutputOpeningsProof = len(paymentProof.ComOutputOpeningsProof)
//	byteArray.lenarrayEqualityOfCommittedValProof = len(paymentProof.EqualityOfCommittedValProof)
//	byteArray.lenarrayOneOfManyProof = len(paymentProof.OneOfManyProof)
//
//	byteArray.ComInputOpeningsProof = paymentProof.ComInputOpeningsProof[0].Bytes()
//	for i := 1; i < byteArray.lenarrayComInputOpeningsProof; i++ {
//		byteArray.ComInputOpeningsProof = append(byteArray.ComInputOpeningsProof, paymentProof.ComInputOpeningsProof[i].Bytes()...)
//	}
//
//	byteArray.ComOutputOpeningsProof = paymentProof.ComOutputOpeningsProof[0].Bytes()
//	for i := 1; i < byteArray.lenarrayComOutputOpeningsProof; i++ {
//		byteArray.ComOutputOpeningsProof = append(byteArray.ComOutputOpeningsProof, paymentProof.ComOutputOpeningsProof[i].Bytes()...)
//	}
//
//	byteArray.EqualityOfCommittedValProof = paymentProof.EqualityOfCommittedValProof[0].Bytes()
//	for i := 1; i < byteArray.lenarrayEqualityOfCommittedValProof; i++ {
//		byteArray.EqualityOfCommittedValProof = append(byteArray.EqualityOfCommittedValProof, paymentProof.EqualityOfCommittedValProof[i].Bytes()...)
//	}
//
//	byteArray.OneOfManyProof, _ = paymentProof.OneOfManyProof[0].Bytes()
//	for i := 1; i < byteArray.lenarrayOneOfManyProof; i++ {
//		outOfManyProofBytes, _ :=  paymentProof.OneOfManyProof[i].Bytes()
//		byteArray.OneOfManyProof = append(byteArray.OneOfManyProof, outOfManyProofBytes...)
//	}
//
//	// byteArray.ComMultiRangeProof = paymentProof.ComMultiRangeProof.Bytes()
//	byteArray.ComZeroProof = paymentProof.ComZeroProof.Bytes()
//	//byteArray.ComZeroOneProof = paymentProof.ComZeroOneProof.Bytes()
//	return []byte{0}
//
//	//Todo: thunderbird
//	//
//
//}

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
	spendingKey *big.Int, inputCoins []*privacy.InputCoin, outputCoins []*privacy.OutputCoin,
	pkLastByteSender byte, pkLastByteReceivers []byte,
	commitments []*privacy.EllipticPoint, randCmIndices []*privacy.CMIndex, myCmPos []uint32) {

	wit.spendingKey = spendingKey
	wit.inputCoins = inputCoins
	wit.outputCoins = outputCoins
	wit.commitments= commitments
	wit.randCmIndices = randCmIndices
	wit.myCmPos = myCmPos

	// Todo: cmInputPartialSK := g^(sk - last byte)

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

	// commit each component of coin commitment
	for i, inputCoin := range wit.inputCoins {
		randInputValue[i] = privacy.RandInt()
		randInputSND[i] = privacy.RandInt()
		cmInputValue[i] = privacy.PedCom.CommitAtIndex(big.NewInt(int64(inputCoin.CoinDetails.Value)), randInputValue[i], privacy.VALUE)
		cmInputSND[i] = privacy.PedCom.CommitAtIndex(inputCoin.CoinDetails.SNDerivator, randInputSND[i], privacy.SND)

		// Todo: using randInputSK or randInputSND
		cmInputSNDIndexSK[i] = privacy.PedCom.CommitAtIndex(inputCoin.CoinDetails.SNDerivator, randInputSK, privacy.SK)
	}

	// Summing all commitments of each input coin into one commitment and proving the knowledge of its Openings
	cmInputSum := make([]*privacy.EllipticPoint, numInputCoin)
	cmInputSumInverse := make([]*privacy.EllipticPoint, numInputCoin)
	randInputSum := make([]*big.Int, numInputCoin)
	wit.ComInputOpeningsWitness = make([]*PKComOpeningsWitness, numInputCoin)

	// cmInputSumAll is sum of all coin commitments
	cmInputSumAll := privacy.EllipticPoint{big.NewInt(0), big.NewInt(0)}

	// randInputSumAll is sum of all randomess of coin commitments
	randInputSumAll := big.NewInt(0)

	wit.OneOfManyWitness = make([]*PKOneOfManyWitness, numInputCoin)
	wit.EqualityOfCommittedValWitness = make([]*PKEqualityOfCommittedValWitness, numInputCoin)
	indexZKPEqual := make([]*byte, 2)
	indexZKPEqual[0] = new(byte)
	*indexZKPEqual[0] = privacy.SK
	indexZKPEqual[1] = new(byte)
	*indexZKPEqual[1] = privacy.SND

	for i := 0; i < numInputCoin; i++ {
		cmInputSum[i] = cmInputSK
		cmInputSum[i].X, cmInputSum[i].Y = privacy.Curve.Add(cmInputSum[i].X, cmInputSum[i].Y, cmInputValue[i].X, cmInputValue[i].Y)
		cmInputSum[i].X, cmInputSum[i].Y = privacy.Curve.Add(cmInputSum[i].X, cmInputSum[i].Y, cmInputSND[i].X, cmInputSND[i].Y)
		cmInputSum[i].X, cmInputSum[i].Y = privacy.Curve.Add(cmInputSum[i].X, cmInputSum[i].Y, cmInputShardID.X, cmInputShardID.Y)
		cmInputSumInverse[i], _ = cmInputSum[i].Inverse()
		cmInputSumAll.X, cmInputSumAll.Y = privacy.Curve.Add(cmInputSum[i].X, cmInputSum[i].Y, cmInputSumAll.X, cmInputSumAll.Y)

		randInputSum[i] = randInputSK
		randInputSum[i].Add(randInputSum[i], randInputValue[i])
		randInputSum[i].Add(randInputSum[i], randInputSND[i])
		randInputSum[i].Add(randInputSum[i], randInputShardID)

		randInputSumAll.Add(randInputSumAll, randInputSum[i])
		randInputSumAll.Mod(randInputSumAll, privacy.Curve.Params().N)
		// Build witness for proving the knowledge of input coins' Openings
		wit.ComInputOpeningsWitness[i].Set(cmInputSum[i], []*big.Int{wit.spendingKey, big.NewInt(int64(inputCoins[i].CoinDetails.Value)), inputCoins[i].CoinDetails.SNDerivator, big.NewInt(int64(pkLastByteSender)), randInputSum[i]})
		// ---------------
		//  Build witness for proving one-out-of-N commitments is a commitment to the coins being spent
		cmInputRndIndex := new(privacy.CMIndex)
		cmInputRndIndex.GetCmIndex(cmInputSum[i])
		//cmInputRndIndexList, cmInputRndValue, indexInputIsZero := GetCMList(cmInputSum[i], cmInputRndIndex, GetCurrentBlockHeight())
		//commitmentTemps := new(privacy.EllipticPoint)
		//rndInputIsZero := big.NewInt(0).Sub(inputCoins[i].CoinDetails.Randomness, randInputSum[i])
		//rndInputIsZero.Mod(rndInputIsZero, privacy.Curve.Params().N)
		//for j := 0; j < privacy.CMRingSize; j++ {
		//	cmInputRndValue[j].X, cmInputRndValue[j].Y = privacy.Curve.Add(commitments[j].X, commitments[j].Y, cmInputSumInverse[j].X, cmInputSumInverse[j].Y)
		//}
		//wit.OneOfManyWitness[i].Set(cmInputRndValue, &cmInputRndIndexList, rndInputIsZero, indexInputIsZero, privacy.SK)
		// -------------------
		// For ZKP Equal Commitment Value
		wit.EqualityOfCommittedValWitness[i].Set([]*privacy.EllipticPoint{cmInputSNDIndexSK[i], cmInputSND[i]}, indexZKPEqual, []*big.Int{inputCoins[i].CoinDetails.SNDerivator, randInputSK, randInputSND[i]})
		// ------------------------------
	}

	numberOutputCoin := len(wit.outputCoins)
	randOutputSK := privacy.RandInt()
	cmOutputSK := privacy.PedCom.CommitAtIndex(wit.spendingKey, randOutputSK, privacy.SK)
	cmOutputValue := make([]*privacy.EllipticPoint, numberOutputCoin)
	cmOutputSND := make([]*privacy.EllipticPoint, numberOutputCoin)
	randOutputValue := make([]*big.Int, numberOutputCoin)
	randOutputSND := make([]*big.Int, numberOutputCoin)

	for i, outputCoin := range wit.outputCoins {
		randOutputValue[i] = privacy.RandInt()
		randOutputSND[i] = privacy.RandInt()
		cmOutputValue[i] = privacy.PedCom.CommitAtIndex(big.NewInt(int64(outputCoin.CoinDetails.Value)), randOutputValue[i], privacy.VALUE)
		cmOutputSND[i] = privacy.PedCom.CommitAtIndex(outputCoin.CoinDetails.SNDerivator, randOutputSND[i], privacy.SND)
	}

	cmOutputSum := make([]*privacy.EllipticPoint, numberOutputCoin)
	randOutputSum := make([]*big.Int, numberOutputCoin)
	wit.ComOutputOpeningsWitness = make([]*PKComOpeningsWitness, numberOutputCoin)
	wit.ComOutputMultiRangeWitness = new(PKComMultiRangeWitness)

	cmOutputSumAll := new(privacy.EllipticPoint)
	cmOutputSumAll.X = big.NewInt(0)
	cmOutputSumAll.Y = big.NewInt(0)
	cmOutputSumAllInverse := new(privacy.EllipticPoint)
	cmOutputRndAll := big.NewInt(0)

	for i := 0; i < numberOutputCoin; i++ {
		cmOutputSum[i] = cmOutputSK
		cmOutputSum[i].X, cmOutputSum[i].Y = privacy.Curve.Add(cmOutputSum[i].X, cmOutputSum[i].Y, cmOutputValue[i].X, cmOutputValue[i].Y)
		cmOutputSum[i].X, cmOutputSum[i].Y = privacy.Curve.Add(cmOutputSum[i].X, cmOutputSum[i].Y, cmOutputSND[i].X, cmOutputSND[i].Y)
		cmOutputSumAll.X, cmOutputSumAll.Y = privacy.Curve.Add(cmOutputSum[i].X, cmOutputSum[i].Y, cmOutputSumAll.X, cmOutputSumAll.Y)

		// cmOutputSumInverse[i], _ = cmOutputSum[i].Inverse()
		randOutputSum[i] = randOutputSK
		randOutputSum[i].Add(randOutputSum[i], randOutputValue[i])
		randOutputSum[i].Add(randOutputSum[i], randOutputSND[i])
		cmOutputRndAll.Add(cmOutputRndAll, randOutputSum[i])
		cmOutputRndAll.Mod(cmOutputRndAll, privacy.Curve.Params().N)
		// For ZKP Opening
		wit.ComOutputOpeningsWitness[i].Set(cmOutputSum[i], []*big.Int{wit.spendingKey, big.NewInt(int64(outputCoins[i].CoinDetails.Value)), outputCoins[i].CoinDetails.SNDerivator, randOutputSum[i]})
		// ---------------
	}

	// For Multi Range Protocol
	// proving each output value is less than vmax
	// proving sum of output values is less than vmax
	// TODO wit.ComOutputMultiRangeWitness.Set(???)
	outputValue := make([]*big.Int, numberOutputCoin)
	for i := 0; i < numberOutputCoin; i++ {
		outputValue[i] = big.NewInt(int64(outputCoins[i].CoinDetails.Value))
	}
	wit.ComOutputMultiRangeWitness.Set(outputValue, 64)
	// ------------------------

	// TODO Product Commitment

	// TODO Zero Or One

	// For check Sum(Input's value) == Sum(Output's Value)
	cmOutputSumAllInverse, _ = cmOutputSumAll.Inverse()
	cmEqualCoinValue := new(privacy.EllipticPoint)
	cmEqualCoinValue.X, cmEqualCoinValue.Y = privacy.Curve.Add(cmInputSumAll.X, cmInputSumAll.Y, cmOutputSumAllInverse.X, cmOutputSumAllInverse.Y)
	cmEqualCoinValueRnd := big.NewInt(0)
	cmEqualCoinValueRnd.Sub(randInputSumAll, cmOutputRndAll)
	cmEqualCoinValueRnd.Mod(cmEqualCoinValueRnd, privacy.Curve.Params().N)

	wit.ComZeroWitness = new(PKComZeroWitness)
	index := new(byte)
	*index = privacy.VALUE
	wit.ComZeroWitness.Set(cmEqualCoinValue, index, cmEqualCoinValueRnd)
	// ---------------------------------------------------
}

// Prove creates big proof
func (wit *PaymentWitness) Prove(hasPrivacy bool) (*PaymentProof, error) {
	proof := new(PaymentProof)
	var err error

	// if hasPrivacy == false, don't need to create the zero knowledge proof
	// proving user has spending key corresponding with public key in input coins
	// is proved by signing with spending key
	if !hasPrivacy {
		proof.InputCoins = wit.inputCoins
		proof.OutputCoins = wit.outputCoins
		// Proving that serial number is derived from the committed derivator
		for i := 0; i < len(wit.inputCoins); i++ {
			proof.EqualityOfCommittedValProof[i] = new(PKEqualityOfCommittedValProof)
			proof.ProductCommitmentProof[i] = new(PKComProductProof)
			proof.EqualityOfCommittedValProof[i] = wit.EqualityOfCommittedValWitness[i].Prove()
			proof.ProductCommitmentProof[i], err = wit.ProductCommitmentWitness[i].Prove()
			if err != nil {
				return nil, err
			}
		}

		return proof, nil
	}

	// if hasPrivacy == true

	numInputCoins := len(wit.ComInputOpeningsWitness)
	numOutputCoins := len(wit.ComOutputOpeningsWitness)

	proof.ComInputOpeningsProof = make([]*PKComOpeningsProof, numInputCoins)
	proof.ComOutputOpeningsProof = make([]*PKComOpeningsProof, numOutputCoins)
	proof.OneOfManyProof = make([]*PKOneOfManyProof, numInputCoins)

	for i := 0; i < numInputCoins; i++ {
		// Proving the knowledge of input coins' Openings
		proof.ComInputOpeningsProof[i] = new(PKComOpeningsProof)
		proof.ComInputOpeningsProof[i], err = wit.ComInputOpeningsWitness[i].Prove()
		if err != nil {
			return nil, err
		}

		// Proving one-out-of-N commitments is a commitment to the coins being spent
		proof.OneOfManyProof[i] = new(PKOneOfManyProof)
		proof.OneOfManyProof[i], err = wit.OneOfManyWitness[i].Prove()
		if err != nil {
			return nil, err
		}

		// Proving that serial number is derived from the committed derivator
		proof.EqualityOfCommittedValProof[i] = new(PKEqualityOfCommittedValProof)
		proof.ProductCommitmentProof[i] = new(PKComProductProof)
		proof.EqualityOfCommittedValProof[i] = wit.EqualityOfCommittedValWitness[i].Prove()
		proof.ProductCommitmentProof[i], err = wit.ProductCommitmentWitness[i].Prove()
		if err != nil {
			return nil, err
		}
	}

	// Proving the knowledge of output coins' openings
	for i := 0; i < numOutputCoins; i++ {
		// Proving the knowledge of output coins' openings
		proof.ComOutputOpeningsProof[i] = new(PKComOpeningsProof)
		proof.ComOutputOpeningsProof[i], err = wit.ComOutputOpeningsWitness[i].Prove()
		if err != nil {
			return nil, err
		}
	}

	// Proving that each output values does not exceed v_max
	proof.ComOutputMultiRangeProof, err = wit.ComOutputMultiRangeWitness.Prove()
	if err != nil {
		return nil, err
	}

	// Proving that sum of all output values does not exceed v_max
	proof.SumOutRangeProof, err = wit.SumOutRangeWitness.Prove()
	if err != nil {
		return nil, err
	}

	// Proving that sum of all input values is equal to sum of all output values
	proof.ComZeroProof, err = wit.ComZeroWitness.Prove()
	if err != nil {
		return nil, err
	}

	return proof, nil
}

func (pro PaymentProof) Verify(hasPrivacy bool, pubKey privacy.PublicKey) bool {
	// if hasPrivacy == false,
	//numInputCoin := len(pro.InputCoins)
	pubKeyPoint, _ := privacy.DecompressKey(pubKey)

	if !hasPrivacy {
		var sumInputValue, sumOutputValue uint64
		sumInputValue = 0
		sumOutputValue = 0

		for i := 0; i < len(pro.InputCoins); i++ {
			// check if input coin's public key is pubKey or not
			// pubKey is the signing key for tx
			if !pro.InputCoins[i].CoinDetails.PublicKey.IsEqual(pubKeyPoint) {
				return false
			}

			// Check input coins' Serial number is created from input coins' SND and sender's spending key
			if !pro.EqualityOfCommittedValProof[i].Verify() {
				return false
			}
			if !pro.ProductCommitmentProof[i].Verify() {
				return false
			}

			// Check input coins' cm is calculated correctly
			cmTmp := pro.InputCoins[i].CoinDetails.PublicKey
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SND].ScalarMul(pro.InputCoins[i].CoinDetails.SNDerivator))
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.VALUE].ScalarMul(big.NewInt(int64(pro.InputCoins[i].CoinDetails.Value))))
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.RAND].ScalarMul(pro.InputCoins[i].CoinDetails.Randomness))
			if !cmTmp.IsEqual(pro.InputCoins[i].CoinDetails.CoinCommitment) {
				return false
			}

			// Check input coins' cm is exists in cm list (Database)
			//Todo

			// Calculate sum of input values
			sumInputValue += pro.InputCoins[i].CoinDetails.Value

		}

		for i := 0; i < len(pro.OutputCoins); i++ {
			// Check output coins' SND is not exists in SND list (Database)
			// Todo

			// Check output coins' cm is calculated correctly
			cmTmp := pro.OutputCoins[i].CoinDetails.PublicKey
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SND].ScalarMul(pro.OutputCoins[i].CoinDetails.SNDerivator))
			cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.VALUE].ScalarMul(big.NewInt(int64(pro.OutputCoins[i].CoinDetails.Value))))
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

	// Verify the proof that output values do not exceed v_max
	// if !pro.ComMultiRangeProof.Verify() {
	// 	return false
	// }

	// Verify the proof that sum of all output values do not exceed v_max
	if !pro.SumOutRangeProof.Verify() {
		return false
	}

	// Verify the proof that sum of all input values is equal to sum of all output values
	if !pro.ComZeroProof.Verify() {
		return false
	}

	return true
}

// GetCMList returns list of CMRingSize (2^3) commitments and list of corresponding cmIndexs that includes cm corresponding to cmIndex
// And return index of cm in list
// the list is sorted by blockHeight, txIndex, CmId
func GetCMList(cm *privacy.EllipticPoint, cmIndex *privacy.CMIndex, blockHeightCurrent *big.Int) ([]*privacy.CMIndex, []*privacy.EllipticPoint, *int) {

	cmIndexs := make([]*privacy.CMIndex, privacy.CMRingSize)
	cms := make([]*privacy.EllipticPoint, privacy.CMRingSize)

	// Random 7 triples (blockHeight, txID, cmIndex)
	for i := 0; i < privacy.CMRingSize-1; i++ {
		cmIndexs[i].Randomize(blockHeightCurrent)
	}

	// the last element in list is cmIndex
	cmIndexs[privacy.CMRingSize-1] = cmIndex

	// Sort list cmIndexs
	sort.Slice(cmIndexs, func(i, j int) bool {
		if cmIndexs[i].BlockHeight.Cmp(cmIndexs[j].BlockHeight) == -1 {
			return true
		}
		if cmIndexs[i].BlockHeight.Cmp(cmIndexs[j].BlockHeight) == 1 {
			return false
		}
		if cmIndexs[i].TxIndex < cmIndexs[j].TxIndex {
			return true
		}
		if cmIndexs[i].TxIndex < cmIndexs[j].TxIndex {
			return false
		}
		return cmIndexs[i].CmId < cmIndexs[j].CmId
	})

	var index int

	// Get list of commitment from sorted cmIndexs
	for i := 0; i < privacy.CMRingSize; i++ {
		if cmIndexs[i].IsEqual(cmIndex) {
			cms[i] = cm
			index = i
		}
		//cms[i] = cmIndexs[i].GetCommitment()
	}

	return cmIndexs, cms, &index
}

func GetCurrentBlockHeight() *big.Int {
	//TODO
	return big.NewInt(0)
}

// ToBytes converts payment proof to byte array to send verifiers
func (pro PaymentProof) ToBytes() []byte {
	//ToDo
	return []byte{0}
}

// FromBytes reverts bytes array to payment proof for verifying
func (pro *PaymentProof) FromBytes(bytes []byte) {
	//ToDo

}
