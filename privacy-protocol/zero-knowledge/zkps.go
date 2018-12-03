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
	OneOfManyWitness              *PKOneOfManyWitness
	EqualityOfCommittedValWitness *PKEqualityOfCommittedValWitness
	ComMultiRangeWitness          *PKComMultiRangeWitness
	ComZeroWitness                *PKComZeroWitness
	ComZeroOneWitness             *PKComZeroOneWitness
}

// BEGIN--------------------------------------------------------------------------------------------------------------------------------------------

// PaymentProof contains all of PoK for sending coin
type PaymentProof struct {
	ComOpeningsProof            []*PKComOpeningsProof
	OneOfManyProof              *PKOneOfManyProof
	EqualityOfCommittedValProof *PKEqualityOfCommittedValProof
	ComMultiRangeProof          *PKComMultiRangeProof
	ComZeroProof                *PKComZeroProof
	ComZeroOneProof             *PKComZeroOneProof

	// these following attributes just exist when tx is no privacy
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
func (wit *PaymentWitness) Prove(hasPrivacy bool) *PaymentProof {
	// if no privacy, don't need to create the zero knowledge proof
	if !hasPrivacy{

	}

	// Call protocol proving knowledge of each sum commitment's Openings


	// Proving one-out-of-N commitments is a commitment to the coins being spent

	//cmLists := make([][]*privacy.EllipticPoint, numberInputCoin)
	//witnessOneOutOfN := make([]*PKOne, len(inputCoins))

	/*for i := 0; i < numberInputCoin; i++ {

	}*/
	// get sum commitment inverse


	// Prepare list of commitments for each commitmentSum that includes 2^3 commiments
	// Get all commitments in inputCoin[i]'s BlockHeight and other block (if needed)
	//cmLists[i] = make([]*privacy.EllipticPoint, CMRingSize)
	//cmLists[i] = GetCMList(wit.inputCoins[i].CoinDetails.CoinCommitment)
	//for j := 0; j < CMRingSize; j++ {
	//	cmLists[i][j].X, cmLists[i][j].Y = privacy.Curve.Add(cmLists[i][j].X, cmLists[i][j].Y, cmSumInverse[i].X, cmSumInverse[i].Y)
	//}

	// Prepare witness for protocol one-out-of-N
	//witnessOneOutOfN[i].Set()

	// }

	// Proving that serial number is derived from the committed derivator

	// Proving that output values do not exceed v_max
	//

	//BEGIN--------------------------------------------------------------------------------------------------------------------------------------------

	// Calculate COMM(sk,r1)+COMM(snd,r2)
	// Calculate G[x]*(1/(sk+snd))

	// Proving that sum of inputs equals sum of outputs
	// prove ( cmvaluein cmvalueout) (commit + s...)
	// cmValueIn := new(privacy.EllipticPoint)
	// cmValueIn.X, cmValueIn.Y = big.NewInt(0), big.NewInt(0)
	// cmValueRndIn := big.NewInt(0)
	// //------------
	// cmValueOut := new(privacy.EllipticPoint)
	// cmValueOut.X, cmValueOut.Y = big.NewInt(0), big.NewInt(0)
	// cmValueRndOut := big.NewInt(0)
	// //------------
	// for i := 0; i < numberInputCoin; i++ {
	// 	cmValueIn.X, cmValueIn.Y = privacy.Curve.Add(cmValueIn.X, cmValueIn.Y, cmValue[i].X, cmValue[i].Y)
	// 	cmValueRndIn = cmValueRndIn.Add(cmValueRndIn, randValue[i])
	// 	cmValueRndIn = cmValueRndIn.Mod(cmValueRndIn, privacy.Curve.Params().N)
	// }

	// //cmEqualValue.X, cmEqualValue.Y = big.NewInt(0), big.NewInt(0)
	// cmEqualValue, _ := cmValueIn.Inverse()
	// cmEqualValue.X, cmEqualValue.Y = privacy.Curve.Add(cmEqualValue.X, cmEqualValue.Y, cmValueOut.X, cmValueOut.Y)
	// cmEqualValueRnd := big.NewInt(0)
	// *cmEqualValueRnd = *cmValueRndIn
	// cmEqualValueRnd = cmEqualValueRnd.Sub(cmEqualValueRnd, cmValueRndOut)
	// cmEqualValueRnd = cmEqualValueRnd.Mod(cmEqualValueRnd, privacy.Curve.Params().N)

	// witnessEqualValue := new(PKComZeroWitness)
	// var cmEqualIndex byte
	// cmEqualIndex = privacy.VALUE
	// witnessEqualValue.Set(cmEqualValue, &cmEqualIndex, cmEqualValueRnd)
	// proofEqualValue, _ := witnessEqualValue.Prove()
	// proofEqualValue.Verify()
	// //END----------------------------------------------------------------------------------------------------------------------------------------------

	return nil
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
