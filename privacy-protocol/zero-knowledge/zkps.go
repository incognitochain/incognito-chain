package zkp

import (
	"math/big"
	"sort"

	"github.com/ninjadotorg/constant/privacy-protocol"
)

const (
	CMRingSize = 8
)

type PaymentWitness struct {
	spendingKey *big.Int
	inputCoins  []*privacy.InputCoin
	outputCoins []*privacy.OutputCoin

	ComInputOpeningsWitness       []*PKComOpeningsWitness
	ComOutputOpeningsWitness      []*PKComOpeningsWitness
	OneOfManyWitness              []*PKOneOfManyWitness
	EqualityOfCommittedValWitness *PKEqualityOfCommittedValWitness
	ComMultiRangeWitness          *PKComMultiRangeWitness
	ComZeroWitness                *PKComZeroWitness
	ComZeroOneWitness             *PKComZeroOneWitness
}

// BEGIN--------------------------------------------------------------------------------------------------------------------------------------------

// PaymentProof contains all of PoK for sending coin
type PaymentProof struct {
	ComInputOpeningsProof       []*PKComOpeningsProof
	ComOutputOpeningsProof      []*PKComOpeningsProof
	OneOfManyProof              []*PKOneOfManyProof
	EqualityOfCommittedValProof *PKEqualityOfCommittedValProof
	ComMultiRangeProof          *PKComMultiRangeProof
	ComZeroProof                *PKComZeroProof
	ComZeroOneProof             *PKComZeroOneProof

	// these following attributes just exist when tx doesn't have privacy
	OutputCoins []*privacy.OutputCoin
	InputCoins  []*privacy.InputCoin
}

func (paymentProof *PaymentProof) Bytes() []byte {
	return []byte{0}
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
func (wit *PaymentWitness) Build(hasPrivacy bool, spendingKey *big.Int, inputCoins []*privacy.InputCoin, outputCoins []*privacy.OutputCoin) {
	wit.spendingKey = spendingKey
	wit.inputCoins = inputCoins
	wit.outputCoins = outputCoins

	numberInputCoin := len(wit.inputCoins)
	randInputSK := privacy.RandInt()
	cmInputSK := privacy.PedCom.CommitAtIndex(wit.spendingKey, randInputSK, privacy.SK)
	cmInputValue := make([]*privacy.EllipticPoint, numberInputCoin)
	cmInputSND := make([]*privacy.EllipticPoint, numberInputCoin)
	randInputValue := make([]*big.Int, numberInputCoin)
	randInputSND := make([]*big.Int, numberInputCoin)

	for i, inputCoin := range wit.inputCoins {
		randInputValue[i] = privacy.RandInt()
		randInputSND[i] = privacy.RandInt()
		cmInputValue[i] = privacy.PedCom.CommitAtIndex(big.NewInt(int64(inputCoin.CoinDetails.Value)), randInputValue[i], privacy.VALUE)
		cmInputSND[i] = privacy.PedCom.CommitAtIndex(inputCoin.CoinDetails.SNDerivator, randInputSND[i], privacy.SND)
	}

	// Summing all commitments of each input coin into one commitment and proving the knowledge of its Openings
	cmInputSum := make([]*privacy.EllipticPoint, numberInputCoin)
	cmInputSumInverse := make([]*privacy.EllipticPoint, numberInputCoin)
	randInputSum := make([]*big.Int, numberInputCoin)
	wit.ComInputOpeningsWitness = make([]*PKComOpeningsWitness, numberInputCoin)

	wit.OneOfManyWitness = make([]*PKOneOfManyWitness, numberInputCoin)
	for i := 0; i < numberInputCoin; i++ {
		cmInputSum[i] = cmInputSK
		cmInputSum[i].X, cmInputSum[i].Y = privacy.Curve.Add(cmInputSum[i].X, cmInputSum[i].Y, cmInputValue[i].X, cmInputValue[i].Y)
		cmInputSum[i].X, cmInputSum[i].Y = privacy.Curve.Add(cmInputSum[i].X, cmInputSum[i].Y, cmInputSND[i].X, cmInputSND[i].Y)
		cmInputSumInverse[i], _ = cmInputSum[i].Inverse()
		randInputSum[i] = randInputSK
		randInputSum[i].Add(randInputSum[i], randInputValue[i])
		randInputSum[i].Add(randInputSum[i], randInputSND[i])

		// For ZKP Opening
		wit.ComInputOpeningsWitness[i].Set(cmInputSum[i], []*big.Int{wit.spendingKey, big.NewInt(int64(inputCoins[i].CoinDetails.Value)), inputCoins[i].CoinDetails.SNDerivator, randInputSum[i]})

		// For ZKP One Of Many
		cmInputRndIndex := new(privacy.CMIndex)
		cmInputRndIndex.GetCmIndex(cmInputSum[i])
		cmInputRndIndexList, cmInputRndValue, indexInputIsZero := GetCMList(cmInputSum[i], cmInputRndIndex, GetCurrentBlockHeight())
		rndInputIsZero := big.NewInt(0).Sub(inputCoins[i].CoinDetails.Randomness, randInputSum[i])
		rndInputIsZero.Mod(rndInputIsZero, privacy.Curve.Params().N)
		for j := 0; j < CMRingSize; j++ {
			cmInputRndValue[j].X, cmInputRndValue[j].Y = privacy.Curve.Add(cmInputRndValue[j].X, cmInputRndValue[j].Y, cmInputSumInverse[j].X, cmInputSumInverse[j].Y)
		}
		wit.OneOfManyWitness[i].Set(cmInputRndValue, &cmInputRndIndexList, rndInputIsZero, indexInputIsZero, privacy.SK)

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
	cmOutputSumInverse := make([]*privacy.EllipticPoint, numberOutputCoin)
	randOutputSum := make([]*big.Int, numberOutputCoin)
	wit.ComOutputOpeningsWitness = make([]*PKComOpeningsWitness, numberOutputCoin)
	for i := 0; i < numberOutputCoin; i++ {
		cmOutputSum[i] = cmOutputSK
		cmOutputSum[i].X, cmOutputSum[i].Y = privacy.Curve.Add(cmOutputSum[i].X, cmOutputSum[i].Y, cmOutputValue[i].X, cmOutputValue[i].Y)
		cmOutputSum[i].X, cmOutputSum[i].Y = privacy.Curve.Add(cmOutputSum[i].X, cmOutputSum[i].Y, cmOutputSND[i].X, cmOutputSND[i].Y)
		cmOutputSumInverse[i], _ = cmOutputSum[i].Inverse()
		randOutputSum[i] = randOutputSK
		randOutputSum[i].Add(randOutputSum[i], randOutputValue[i])
		randOutputSum[i].Add(randOutputSum[i], randOutputSND[i])

		// For ZKP Opening
		wit.ComOutputOpeningsWitness[i].Set(cmOutputSum[i], []*big.Int{wit.spendingKey, big.NewInt(int64(outputCoins[i].CoinDetails.Value)), outputCoins[i].CoinDetails.SNDerivator, randOutputSum[i]})
	}

}

// Prove creates big proof
func (wit *PaymentWitness) Prove(hasPrivacy bool) (*PaymentProof, error) {
	proof := new(PaymentProof)
	// if hasPrivacy == false, don't need to create the zero knowledge proof
	// proving user has spending key corresponding with public key in input coins
	// is proved by signing with spending key
	if !hasPrivacy {
		proof.InputCoins = wit.inputCoins
		proof.OutputCoins = wit.outputCoins
	}

	// if hasPrivacy == true
	var err error
	numInputCoins := len(wit.ComOpeningsWitness)
	// Proving the knowledge of input coins' Openings, output coins' openings
	proof.ComOpeningsProof = make([]*PKComOpeningsProof, numInputCoins)
	// Proving one-out-of-N commitments is a commitment to the coins being spent
	proof.OneOfManyProof = make([]*PKOneOfManyProof, numInputCoins)

	for i := 0; i < numInputCoins; i++ {
		proof.ComOpeningsProof[i] = new(PKComOpeningsProof)
		proof.ComOpeningsProof[i], err = wit.ComOpeningsWitness[i].Prove()
		if err != nil {
			return nil, err
		}

		proof.OneOfManyProof[i] = new(PKOneOfManyProof)
		proof.OneOfManyProof[i], err = wit.OneOfManyWitness[i].Prove()
		if err != nil {
			return nil, err
		}
	}

	// Proving that serial number is derived from the committed derivator
	// Todo: 0xKraken

	// Proving that output values do not exceed v_max
	proof.ComMultiRangeProof, err = wit.ComMultiRangeWitness.Prove()
	if err != nil {
		return nil, err
	}
	// Proving that sum of all output values do not exceed v_max
	// Todo: 0xKraken

	return proof, nil
}

func (pro PaymentProof) Verify() bool {
	if !pro.ComInputOpeningsProof[0].Verify() {
		return false
	}
	//if !pro.ComMultiRangeProof
	if !pro.ComZeroOneProof.Verify() {
		return false
	}
	if !pro.ComZeroProof.Verify() {
		return false
	}
	if !pro.EqualityOfCommittedValProof.Verify() {
		return false
	}
	if !pro.OneOfManyProof[0].Verify() {
		return false
	}
	return true
}

// GetCMList returns list of CMRingSize (2^4) commitments and list of corresponding cmIndexs that includes cm corresponding to cmIndex
// And return index of cm in list
// the list is sorted by blockHeight, txIndex, CmId
func GetCMList(cm *privacy.EllipticPoint, cmIndex *privacy.CMIndex, blockHeightCurrent *big.Int) ([]*privacy.CMIndex, []*privacy.EllipticPoint, *int) {

	cmIndexs := make([]*privacy.CMIndex, CMRingSize)
	cms := make([]*privacy.EllipticPoint, CMRingSize)

	// Random 7 triples (blockHeight, txID, cmIndex)
	for i := 0; i < CMRingSize-1; i++ {
		cmIndexs[i].Randomize(blockHeightCurrent)
	}

	// the last element in list is cmIndex
	cmIndexs[CMRingSize-1] = cmIndex

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
	for i := 0; i < CMRingSize; i++ {
		if cmIndexs[i].IsEqual(cmIndex) {
			cms[i] = cm
			index = i
		}
		cms[i] = cmIndexs[i].GetCommitment()
	}

	return cmIndexs, cms, &index
}

func GetCurrentBlockHeight() *big.Int {
	//TODO
	return big.NewInt(1224)
}
