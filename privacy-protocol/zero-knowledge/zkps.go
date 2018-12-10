package zkp

import (
	"github.com/ninjadotorg/constant/privacy-protocol"
	"math/big"
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
	OutputCoins []*privacy.OutputCoin
	InputCoins  []*privacy.InputCoin

	ComOutputValue   []*privacy.EllipticPoint
	ComOutputSND     []*privacy.EllipticPoint
	ComOutputShardID []*privacy.EllipticPoint

	PubKeyLastByteSender byte
}

func (paymentProof *PaymentProof) Bytes() []byte {
	var proofbytes []byte
	// ComInputOpeningsProof
	proofbytes = append(proofbytes, byte(len(paymentProof.ComInputOpeningsProof)))
	for i := 0; i < len(paymentProof.ComInputOpeningsProof); i++ {
		proofbytes = append(proofbytes, paymentProof.ComInputOpeningsProof[i].Bytes()...)
	}
	// OneOfManyProof
	proofbytes = append(proofbytes, byte(len(paymentProof.OneOfManyProof)))
	for i := 0; i < len(paymentProof.OneOfManyProof); i++ {
		proofbytes = append(proofbytes, byte(len(paymentProof.OneOfManyProof[i].Bytes())))
		proofbytes = append(proofbytes, paymentProof.OneOfManyProof[i].Bytes()...)
	}
	// EqualityOfCommittedValProof
	proofbytes = append(proofbytes, byte(len(paymentProof.EqualityOfCommittedValProof)))
	for i := 0; i < len(paymentProof.EqualityOfCommittedValProof); i++ {
		proofbytes = append(proofbytes, paymentProof.EqualityOfCommittedValProof[i].Bytes()...)
	}
	// ProductCommitmentProof
	proofbytes = append(proofbytes, byte(len(paymentProof.ProductCommitmentProof)))
	for i := 0; i < len(paymentProof.ProductCommitmentProof); i++ {
		proofbytes = append(proofbytes, paymentProof.ProductCommitmentProof[i].Bytes()...)
	}
	//ComOutputOpeningsProof
	proofbytes = append(proofbytes, byte(len(paymentProof.ComOutputOpeningsProof)))
	for i := 0; i < len(paymentProof.ComOutputOpeningsProof); i++ {
		proofbytes = append(proofbytes, paymentProof.ComOutputOpeningsProof[i].Bytes()...)
	}
	// ComOutputMultiRangeProof
	/*proofbytes = append(proofbytes, byte(len(paymentProof.ComOutputMultiRangeProof.Bytes())))
	proofbytes = append(proofbytes, paymentProof.ComOutputMultiRangeProof.Bytes()...)
	// ComZeroProof
	proofbytes = append(proofbytes, byte(len(paymentProof.ComZeroProof.Bytes())))
	proofbytes = append(proofbytes, paymentProof.ComZeroProof.Bytes()...)
*///	OutputCoins
	/*proofbytes = append(proofbytes, byte(len(paymentProof.OutputCoins)))
	for i := 0; i < len(paymentProof.OutputCoins); i++ {
		proofbytes = append(proofbytes, paymentProof.OutputCoins[i].Bytes()...)
	}
	//	InputCoins
	proofbytes = append(proofbytes, byte(len(paymentProof.InputCoins)))
	for i := 0; i < len(paymentProof.InputCoins); i++ {
		proofbytes = append(proofbytes, paymentProof.InputCoins[i].Bytes()...)
	}
	//


	proofbytes = append(proofbytes, byte(len(paymentProof.ComInputSK)))*/

	return proofbytes
}

func (paymentProof *PaymentProof) SetBytes(proofbytes []byte) {
	offset := 0
	// Set ComInputOpeningsProof
	lenComInputOpeningsProofArray := int(proofbytes[offset])
	ComInputOpeningsProof := make([]*PKComOpeningsProof, lenComInputOpeningsProofArray)
	for i := 0; i < lenComInputOpeningsProofArray; i++ {
		ComInputOpeningsProof[i] = new(PKComOpeningsProof)
		ComInputOpeningsProof[i].SetBytes(proofbytes[offset:offset+privacy.ComInputOpeningsProofSize])
		offset += privacy.ComInputOpeningsProofSize
	}
	// Set OneOfManyProof
	lenOneOfManyProofArray := int(proofbytes[offset])
	OneOfManyProof := make([]*PKOneOfManyProof, lenOneOfManyProofArray)
	for i := 0; i < lenComInputOpeningsProofArray; i++ {
		offset += 1
		size := int(proofbytes[offset])
		OneOfManyProof[i] = new(PKOneOfManyProof)
		OneOfManyProof[i].SetBytes(proofbytes[offset:offset+size])
		offset += size
	}
	// Set EqualityOfCommittedValProof
	lenEqualityOfCommittedValProof := int(proofbytes[offset])
	EqualityOfCommittedValProof := make([]*PKEqualityOfCommittedValProof, lenEqualityOfCommittedValProof)
	for i := 0; i < lenEqualityOfCommittedValProof; i++ {
		EqualityOfCommittedValProof[i] = new(PKEqualityOfCommittedValProof)
		EqualityOfCommittedValProof[i].SetBytes(proofbytes[offset:offset+privacy.EqualityOfCommittedValProofSize])
		offset += privacy.EqualityOfCommittedValProofSize
	}
	// Set ProductCommitmentProof
	lenProductCommitmentProofArray := int(proofbytes[offset])
	ProductCommitmentProof := make([]*PKComProductProof, lenProductCommitmentProofArray)
	for i := 0; i < lenProductCommitmentProofArray; i++ {
		ProductCommitmentProof[i] = new(PKComProductProof)
		ProductCommitmentProof[i].SetBytes(proofbytes[offset:offset+privacy.ProductCommitmentProofSize])
		offset += privacy.ProductCommitmentProofSize
	}
	//Set ComOutputOpeningsProof
	lenComOutputOpeningsProofArray := int(proofbytes[offset])
	ComOutputOpeningsProof := make([]*PKComOpeningsProof, lenComOutputOpeningsProofArray)
	for i := 0; i < lenComOutputOpeningsProofArray; i++ {
		ComOutputOpeningsProof[i] = new(PKComOpeningsProof)
		ComOutputOpeningsProof[i].SetBytes(proofbytes[offset:offset+privacy.ComOutputOpeningsProofSize])
		offset += privacy.ComOutputOpeningsProofSize
	}
	// Set InputCoin
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
	indexZKPEqual := make([]*byte, 2)
	indexZKPEqual[0] = new(byte)
	*indexZKPEqual[0] = privacy.SK
	indexZKPEqual[1] = new(byte)
	*indexZKPEqual[1] = privacy.SND

	for i := 0; i < numInputCoin; i++ {
		/***** Build witness for proving the knowledge of input coins' Openings  *****/
		cmInputSum[i] = cmInputSK
		cmInputSum[i].X, cmInputSum[i].Y = privacy.Curve.Add(cmInputSum[i].X, cmInputSum[i].Y, cmInputValue[i].X, cmInputValue[i].Y)
		cmInputSum[i].X, cmInputSum[i].Y = privacy.Curve.Add(cmInputSum[i].X, cmInputSum[i].Y, cmInputSND[i].X, cmInputSND[i].Y)

		randInputSum[i] = randInputSK
		randInputSum[i].Add(randInputSum[i], randInputValue[i])
		randInputSum[i].Add(randInputSum[i], randInputSND[i])

		if wit.ComInputOpeningsWitness[i] == nil {
			wit.ComInputOpeningsWitness[i] = new(PKComOpeningsWitness)
		}
		wit.ComInputOpeningsWitness[i].Set(cmInputSum[i], []*big.Int{wit.spendingKey, big.NewInt(int64(inputCoins[i].CoinDetails.Value)), inputCoins[i].CoinDetails.SNDerivator, randInputSum[i]})

		/***** Build witness for proving one-out-of-N commitments is a commitment to the coins being spent *****/
		// commitmentTemps is a list of commitments for protocol one-out-of-N
		commitmentTemps := make([]*privacy.EllipticPoint, numInputCoin*privacy.CMRingSize)
		randInputSum[i].Add(randInputSum[i], randInputShardID)
		rndInputIsZero := big.NewInt(0).Sub(inputCoins[i].CoinDetails.Randomness, randInputSum[i])
		rndInputIsZero.Mod(rndInputIsZero, privacy.Curve.Params().N)

		cmInputSum[i].X, cmInputSum[i].Y = privacy.Curve.Add(cmInputSum[i].X, cmInputSum[i].Y, cmInputShardID.X, cmInputShardID.Y)
		cmInputSumInverse[i], _ = cmInputSum[i].Inverse()

		randInputSumAll.Add(randInputSumAll, randInputSum[i])
		randInputSumAll.Mod(randInputSumAll, privacy.Curve.Params().N)

		//cmInputSumAll.X, cmInputSumAll.Y = privacy.Curve.Add(cmInputSum[i].X, cmInputSum[i].Y, cmInputSumAll.X, cmInputSumAll.Y)

		for j := 0; j < numInputCoin*privacy.CMRingSize; j++ {
			commitmentTemps[j] = new(privacy.EllipticPoint)
			commitmentTemps[j].X = big.NewInt(0)
			commitmentTemps[j].Y = big.NewInt(0)
			commitmentTemps[j].X, commitmentTemps[j].Y = privacy.Curve.Add(commitments[j].X, commitments[j].Y, cmInputSumInverse[i].X, cmInputSumInverse[i].Y)
		}

		wit.OneOfManyWitness[i].Set(commitmentTemps, commitmentIndexs, rndInputIsZero, myCommitmentIndexs[i], privacy.SK)

		/***** Build witness for proving that serial number is derived from the committed derivator *****/
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

		cmOutputSum[i] = new(privacy.EllipticPoint)
		cmOutputSum[i].X, cmOutputSum[i].Y = cmOutputValue[i].X, cmOutputValue[i].Y
		cmOutputSum[i].X, cmOutputSum[i].Y = privacy.Curve.Add(cmOutputSum[i].X, cmOutputSum[i].Y, cmOutputSND[i].X, cmOutputSND[i].Y)

		cmOutputValueAll.Add(cmOutputValue[i])
		randOutputValueAll.Add(randOutputValueAll, randOutputValue[i])

		/***** Build witness for proving the knowledge of output coins' Openings (value, snd, randomness) *****/
		if wit.ComOutputOpeningsWitness[i] == nil {
			wit.ComOutputOpeningsWitness[i] = new(PKComOpeningsWitness)
		}
		wit.ComOutputOpeningsWitness[i].Set(cmOutputSum[i], []*big.Int{big.NewInt(int64(outputCoins[i].CoinDetails.Value)), outputCoins[i].CoinDetails.SNDerivator, randOutputSum[i]})
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

		outputCoins[i].CoinDetails.CoinCommitment = cmOutputSum[i]
		outputCoins[i].CoinDetails.Randomness = randOutputSum[i]

		cmOutputSumAll.X, cmOutputSumAll.Y = privacy.Curve.Add(cmOutputSum[i].X, cmOutputSum[i].Y, cmOutputSumAll.X, cmOutputSumAll.Y)

	}

	// For Multi Range Protocol
	// proving each output value is less than vmax
	// proving sum of output values is less than vmax
	outputValue := make([]*big.Int, numOutputCoin)
	for i := 0; i < numOutputCoin; i++ {
		outputValue[i] = big.NewInt(int64(outputCoins[i].CoinDetails.Value))
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
	proof := new(PaymentProof)
	var err error

	proof.InputCoins = wit.inputCoins
	proof.OutputCoins = wit.outputCoins

	// if hasPrivacy == false, don't need to create the zero knowledge proof
	// proving user has spending key corresponding with public key in input coins
	// is proved by signing with spending key
	if !hasPrivacy {
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

	// Proving that each output values and sum of them does not exceed v_max
	proof.ComOutputMultiRangeProof, err = wit.ComOutputMultiRangeWitness.Prove()
	var err1 error
	proof.SumOutRangeProof, err = wit.ComOutputMultiRangeWitness.ProveSum()
	if err != nil && err1 != nil {
		return nil, err
	}

	// Proving that sum of all input values is equal to sum of all output values
	proof.ComZeroProof, err = wit.ComZeroWitness.Prove()
	if err != nil {
		return nil, err
	}

	// hide

	return proof, nil
}

func (pro PaymentProof) Verify(hasPrivacy bool, pubKey privacy.PublicKey, commitmentsDB []*privacy.EllipticPoint) bool {
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
		if !pro.OneOfManyProof[i].Verify(commitmentsDB) {
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
