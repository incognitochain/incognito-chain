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

	ComOpeningsWitness            []*PKComOpeningsWitness
	OneOfManyWitness              []*PKOneOfManyWitness
	EqualityOfCommittedValWitness *PKEqualityOfCommittedValWitness
	ComMultiRangeWitness          *PKComMultiRangeWitness
	ComZeroWitness                *PKComZeroWitness
	ComZeroOneWitness             *PKComZeroOneWitness
}

// BEGIN--------------------------------------------------------------------------------------------------------------------------------------------

// PaymentProof contains all of PoK for sending coin
type PaymentProof struct {
	ComOpeningsProof            []*PKComOpeningsProof
	OneOfManyProof              []*PKOneOfManyProof
	EqualityOfCommittedValProof *PKEqualityOfCommittedValProof
	ComMultiRangeProof          *PKComMultiRangeProof
	ComZeroProof                *PKComZeroProof
	ComZeroOneProof             *PKComZeroOneProof

	// these following attributes just exist when tx doesn't have privacy
	OutputCoins									[]*privacy.OutputCoin
	InputCoins									[]*privacy.InputCoin
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
	randSK := privacy.RandInt()
	// Commit each component of coins being spent
	cmSK := privacy.PedCom.CommitAtIndex(wit.spendingKey, randSK, privacy.SK)
	cmValue := make([]*privacy.EllipticPoint, numberInputCoin)
	cmSND := make([]*privacy.EllipticPoint, numberInputCoin)
	// cmAll := make([]*privacy.EllipticPoint, numberInputCoin)
	randValue := make([]*big.Int, numberInputCoin)
	randSND := make([]*big.Int, numberInputCoin)
	for i, inputCoin := range wit.inputCoins {
		// cmAll[i] = privacy.PedCom.CommitAll([]*big.Int{spendingKey, big.NewInt(int64(inputCoin.CoinDetails.Value)), inputCoin.CoinDetails.SNDerivator, randValue})
		randValue[i] = privacy.RandInt()
		randSND[i] = privacy.RandInt()
		cmValue[i] = privacy.PedCom.CommitAtIndex(big.NewInt(int64(inputCoin.CoinDetails.Value)), randValue[i], privacy.VALUE)
		cmSND[i] = privacy.PedCom.CommitAtIndex(inputCoin.CoinDetails.SNDerivator, randSND[i], privacy.SND)
	}

	// Summing all commitments of each input coin into one commitment and proving the knowledge of its Openings
	cmSum := make([]*privacy.EllipticPoint, numberInputCoin)
	cmSumInverse := make([]*privacy.EllipticPoint, numberInputCoin)
	randSum := make([]*big.Int, numberInputCoin)
	wit.ComOpeningsWitness = make([]*PKComOpeningsWitness, numberInputCoin)
	for i := 0; i < numberInputCoin; i++ {
		cmSum[i] = cmSK
		cmSum[i].X, cmSum[i].Y = privacy.Curve.Add(cmSum[i].X, cmSum[i].Y, cmValue[i].X, cmValue[i].Y)
		cmSum[i].X, cmSum[i].Y = privacy.Curve.Add(cmSum[i].X, cmSum[i].Y, cmSND[i].X, cmSND[i].Y)
		cmSumInverse[i], _ = cmSum[i].Inverse()
		randSum[i] = randSK
		randSum[i].Add(randSum[i], randValue[i])
		randSum[i].Add(randSum[i], randSND[i])

		// For ZKP Opening
		wit.ComOpeningsWitness[i].Set(cmSum[i], []*big.Int{wit.spendingKey, big.NewInt(int64(inputCoins[i].CoinDetails.Value)), inputCoins[i].CoinDetails.SNDerivator, randSum[i]})

		// For ZKP One Of Many
		cmRndIndex := new(privacy.CMIndex)
		cmRndIndex.GetCmIndex(cmSum[i])
		cmRndIndexList, cmRndValue, indexIsZero := GetCMList(cmSum[i], cmRndIndex, GetCurrentBlockHeight())
		rndIsZero := big.NewInt(0).Sub(inputCoins[i].CoinDetails.Randomness, randSum[i])
		rndIsZero.Mod(rndIsZero, privacy.Curve.Params().N)
		for j := 0; j < CMRingSize; j++ {
			cmRndValue[j].X, cmRndValue[j].Y = privacy.Curve.Add(cmRndValue[j].X, cmRndValue[j].Y, cmSumInverse[j].X, cmSumInverse[j].Y)
		}
		wit.OneOfManyWitness.Set(cmRndValue, &cmRndIndexList, rndIsZero, indexIsZero, privacy.SK)

	}
	//todo

}

// Prove creates big proof
func (wit *PaymentWitness) Prove(hasPrivacy bool) (*PaymentProof, error) {
	proof := new(PaymentProof)
	// if hasPrivacy == false, don't need to create the zero knowledge proof
	// proving user has spending key corresponding with public key in input coins
	// is proved by signing with spending key
	if !hasPrivacy{
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

	for i:=0; i < numInputCoins; i++{
		proof.ComOpeningsProof[i] = new(PKComOpeningsProof)
		proof.ComOpeningsProof[i], err = wit.ComOpeningsWitness[i].Prove()
		if err != nil{
			return nil, err
		}

		proof.OneOfManyProof[i] = new(PKOneOfManyProof)
		proof.OneOfManyProof[i], err = wit.OneOfManyWitness[i].Prove()
		if err != nil{
			return nil, err
		}
	}

	// Proving that serial number is derived from the committed derivator
	// Todo: 0xKraken

	// Proving that output values do not exceed v_max
	proof.ComMultiRangeProof, err = wit.ComMultiRangeWitness.Prove()
	if err != nil{
		return nil, err
	}
	// Proving that sum of all output values do not exceed v_max
	// Todo: 0xKraken


	return proof, nil
}

func (pro PaymentProof) Verify(hasPrivacy bool) bool {
	if !pro.ComOpeningsProof[0].Verify() {
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
	if !pro.OneOfManyProof.Verify() {
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
