package zkp

import (
	"math/big"

	"github.com/ninjadotorg/constant/privacy-protocol"
)

const (
	CMRingSize = 16
)

type PaymentWitness struct{
	spendingKey *big.Int
	inputCoins []*privacy.InputCoin
	outputCoins []*privacy.OutputCoin
}

// BEGIN--------------------------------------------------------------------------------------------------------------------------------------------

// ProofOfPayment contains all of PoK for sending coin
type PaymentProof struct {
	ComMultiRangeProof *PKComMultiRangeProof
	ComOpeningsProof   *PKComOpeningsProof
	ComZeroOneProof    *PKComZeroOneProof
	ComZeroProof       *PKComZeroProof
	InEqualOutProof    *PKInEqualOutProof
}

// END----------------------------------------------------------------------------------------------------------------------------------------------


func (wit *PaymentWitness) Set(spendingKey *big.Int, inputCoins []*privacy.InputCoin, outputCoins []*privacy.OutputCoin){
	wit.spendingKey = spendingKey
	wit.inputCoins = inputCoins
	wit.outputCoins = outputCoins
}
// Prove creates big proof
func (wit *PaymentWitness) Prove() {

	numberInputCoin := len(wit.inputCoins)
	// Commit each component of coins being spent
	cmSK := make([]*privacy.EllipticPoint, numberInputCoin)
	cmValue := make([]*privacy.EllipticPoint, numberInputCoin)
	cmSND := make([]*privacy.EllipticPoint, numberInputCoin)

	randSK := make([]*big.Int, numberInputCoin)
	randValue := make([]*big.Int, numberInputCoin)
	randSND := make([]*big.Int, numberInputCoin)
	for i, inputCoin := range wit.inputCoins {
		randSK[i] = privacy.RandInt()
		randValue[i] = privacy.RandInt()
		randSND[i] = privacy.RandInt()

		cmSK[i] = privacy.PedCom.CommitAtIndex(wit.spendingKey, randSK[i], privacy.SK)
		cmValue[i] = privacy.PedCom.CommitAtIndex(big.NewInt(int64(inputCoin.CoinDetails.Value)), randValue[i], privacy.VALUE)
		cmSND[i] = privacy.PedCom.CommitAtIndex(inputCoin.CoinDetails.SNDerivator, randSND[i], privacy.SND)
	}

	// Summing all commitments of each input coin into one commitment and proving the knowledge of its openings
	cmSum := make([]*privacy.EllipticPoint, numberInputCoin)
	randSum := make([]*big.Int, numberInputCoin)
	for i := 0; i < numberInputCoin; i++ {
		cmSum[i] = cmSK[i]
		cmSum[i].X, cmSum[i].Y = privacy.Curve.Add(cmSum[i].X, cmSum[i].Y, cmValue[i].X, cmValue[i].Y)
		cmSum[i].X, cmSum[i].Y = privacy.Curve.Add(cmSum[i].X, cmSum[i].Y, cmSND[i].X, cmSND[i].Y)

		randSum[i] = randSK[i]
		randSum[i].Add(randSum[i], randValue[i])
		randSum[i].Add(randSum[i], randSND[i])
	}

	// Call protocol proving knowledge of each sum commitment's openings

	// Proving one-out-of-N commitments is a commitment to the coins being spent
	cmSumInverse := make([]*privacy.EllipticPoint, numberInputCoin)
	cmLists := make([][]*privacy.EllipticPoint, numberInputCoin)
	//witnessOneOutOfN := make([]*PKOne, len(inputCoins))
	for i := 0; i < numberInputCoin; i++ {
		// get sum commitment inverse
		cmSumInverse[i], _ = cmSum[i].Inverse()

		// Prepare list of commitments for each commitmentSum that includes 2^8 commiments
		// Get all commitments in inputCoin[i]'s BlockHeight and other block (if needed)
		cmLists[i] = make([]*privacy.EllipticPoint, CMRingSize)
		cmLists[i] = GetCMList(wit.inputCoins[i].CoinDetails.CoinCommitment, wit.inputCoins[i].BlockHeight)
		for j := 0; j < CMRingSize; j++ {
			cmLists[i][j].X, cmLists[i][j].Y = privacy.Curve.Add(cmLists[i][j].X, cmLists[i][j].Y, cmSumInverse[i].X, cmSumInverse[i].Y)
		}

		// Prepare witness for protocol one-out-of-N
		//witnessOneOutOfN[i].Set()

	}

	// Proving that serial number is derived from the committed derivator

	// Proving that output values do not exceed v_max

	//BEGIN--------------------------------------------------------------------------------------------------------------------------------------------

	// Calculate COMM(sk,r1)+COMM(snd,r2)
	// Calculate G[x]*(1/(sk+snd))

	// Proving that sum of inputs equals sum of outputs
	// prove ( cmvaluein cmvalueout) (commit + s...)
	cmValueIn := new(privacy.EllipticPoint)
	cmValueIn.X, cmValueIn.Y = big.NewInt(0), big.NewInt(0)
	cmValueRndIn := big.NewInt(0)
	//------------
	cmValueOut := new(privacy.EllipticPoint)
	cmValueOut.X, cmValueOut.Y = big.NewInt(0), big.NewInt(0)
	cmValueRndOut := big.NewInt(0)
	//------------
	for i := 0; i < numberInputCoin; i++ {
		cmValueIn.X, cmValueIn.Y = privacy.Curve.Add(cmValueIn.X, cmValueIn.Y, cmValue[i].X, cmValue[i].Y)
		cmValueRndIn = cmValueRndIn.Add(cmValueRndIn, randValue[i])
		cmValueRndIn = cmValueRndIn.Mod(cmValueRndIn, privacy.Curve.Params().N)
	}

	//cmEqualValue.X, cmEqualValue.Y = big.NewInt(0), big.NewInt(0)
	cmEqualValue, _ := cmValueIn.Inverse()
	cmEqualValue.X, cmEqualValue.Y = privacy.Curve.Add(cmEqualValue.X, cmEqualValue.Y, cmValueOut.X, cmValueOut.Y)
	cmEqualValueRnd := big.NewInt(0)
	*cmEqualValueRnd = *cmValueRndIn
	cmEqualValueRnd = cmEqualValueRnd.Sub(cmEqualValueRnd, cmValueRndOut)
	cmEqualValueRnd = cmEqualValueRnd.Mod(cmEqualValueRnd, privacy.Curve.Params().N)

	witnessEqualValue := new(PKComZeroWitness)
	var cmEqualIndex byte
	cmEqualIndex = privacy.VALUE
	witnessEqualValue.Set(cmEqualValue, &cmEqualIndex, cmEqualValueRnd)
	proofEqualValue, _ := witnessEqualValue.Prove()
	proofEqualValue.Verify()
	//END----------------------------------------------------------------------------------------------------------------------------------------------

}

func (pro PaymentProof) Verify() bool{
	return true
}

// GetCMList returns list CMRingSize (2^4) commitments that includes cm in blockHeight
func GetCMList(cm *privacy.EllipticPoint, blockHeight *big.Int) []*privacy.EllipticPoint {
	return nil
}
