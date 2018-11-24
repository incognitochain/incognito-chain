package zkp


import (
	"crypto/rand"
	"math/big"

	"github.com/ninjadotorg/constant/privacy-protocol"
)

// type PKComZeroProtocol struct {
// 	Witness PKComZeroWitnees
// 	Proof   PKComZeroProof
// }

type PKComZeroProof struct {
	commitmentValue *privacy.EllipticPoint //statement
	index           byte                   //statement
	commitmentZeroS *privacy.EllipticPoint
	z               *big.Int
}

type PKComZeroWitnees struct {
	commitmentRnd *big.Int
}

/*Protocol for opening a PedersenCommitment to 0
Prove:
	commitmentValue is PedersenCommitment value of Zero, that is statement needed to prove
	commitmentValue is calculated by Comm_ck(Value,PRDNumber)
	commitmentRnd is PRDNumber, which is used to calculate commitmentValue
	s <- Zp; P is privacy.Curve base point's order, is N
	B <- Comm_ck(0,s);  Comm_ck is PedersenCommit function using public params - privacy.Curve.Params() (G0,G1...)
						but is just commit special value (in this case, special value is 0),
						which is stick with G[Index] (in this case, Index is the Index stick with commitmentValue)
						B is a.k.a commitmentZeroS
	x <- Hash(G0||G1||G2||G3||commitmentvalue) x is pseudorandom number, which could be computed easily by Verifier
	z <- rx + s; z in Zp, r is commitmentRnd
	return commitmentZeroS, z

Verify:
	commitmentValue is PedersenCommitment value of Zero, that is statement needed to prove
	commitmentValue is calculated by Comm_ck(Value,PRDNumber), a.k.a A
	commitmentZeroS, z are output of Prove function, commitmentZeroS is a.k.a B
	x <- Hash(G0||G1||G2||G3||commitmentvalue)
	boolValue <- (Comm_ck(0,z) == A.x + B); in this case, A and B needed to convert to privacy.privacy.EllipticPoint
	return boolValue
)
*/

// SetValue dosomethings
func (wit *PKComZeroWitnees) SetValue(commitmentRnd *big.Int) {
	pro.commitmentRnd = commitmentRnd
}

// SetValue dosomethings
func (pro *PKComZeroProof) SetValue(
	commitmentValue *privacy.EllipticPoint, //statement
	index byte, //statement
	commitmentZeroS *privacy.EllipticPoint,
	z *big.Int) {

	if pro.commitmentValue == nil {
		pro.commitmentValue = new(privacy.EllipticPoint)
		pro.commitmentValue.X = big.NewInt(0)
		pro.commitmentValue.Y = big.NewInt(0)
	}
	*(pro.commitmentValue.X) = *(commitmentValue.X)
	*(pro.commitmentValue.Y) = *(commitmentValue.Y)

	if pro.commitmentZeroS == nil {
		pro.commitmentZeroS = new(privacy.EllipticPoint)
		pro.commitmentZeroS.X = big.NewInt(0)
		pro.commitmentZeroS.Y = big.NewInt(0)
	}
	*(pro.commitmentZeroS.X) = *(commitmentZeroS.X)
	*(pro.commitmentZeroS.Y) = *(commitmentZeroS.Y)

	pro.index = index
	if pro.z == nil {
		pro.z = big.NewInt(0)
	}
	*(pro.z) = *z

}

//Prove generate a Proof prove that the PedersenCommitment is zero
func (wit PKComZeroWitnees) Prove(commitmentValue *privacy.EllipticPoint, index byte) (*PKComZeroProof, error) { //???
	//var x big.Int
	//s is a random number in Zp, with p is N, which is order of base point of privacy.Curve
	sRnd, _ := rand.Int(rand.Reader, privacy.Curve.Params().N)
	// if err != nil {
	// 	panic(err)
	// }

	//Calculate B = commitmentZeroS = comm_ck(0,s,Index)
	commitmentZeroS := privacy.PedCom.CommitAtIndex(big.NewInt(0), sRnd, index)

	//Generate challenge x in Zp
	xChallenge := GenerateChallenge([][]byte{commitmentValue.Bytes()})

	//Calculate z=r*x + s (mod N)
	z := big.NewInt(0)
	*z = *(wit.commitmentRnd)
	z.Mul(z, xChallenge)
	z.Mod(z, privacy.Curve.Params().N)
	z.Add(z, sRnd)
	z.Mod(z, privacy.Curve.Params().N)

	proof := new(PKComZeroProof)
	proof.SetValue(commitmentValue, index, commitmentZeroS, z)
	return proof, nil
}

//Verify verify that under PedersenCommitment is zero
func (pro *PKComZeroProof) Verify() bool {
	//Generate challenge x in Zp
	xChallenge := GenerateChallenge([][]byte{pro.commitmentValue.Bytes()})

	//convert commitmentValue []byte to Point in ECC
	// commitmentValuePoint, err := privacy.DecompressKey(commitmentValue)
	// if err != nil {
	// 	return false
	// }
	// if (!privacy.Curve.IsOnCurve(commitmentValuePoint.X, commitmentValuePoint.Y)) || (z.Cmp(privacy.Curve.Params().N) > -1) {
	// 	return false
	// }

	//verifyPoint is result of A.x + B (in ECC)
	verifyPoint := new(privacy.EllipticPoint)
	verifyPoint.X = big.NewInt(0)
	verifyPoint.Y = big.NewInt(0)
	//Set verifyPoint = A
	*(verifyPoint.X) = *(pro.commitmentValue.X)
	*(verifyPoint.Y) = *(pro.commitmentValue.Y)
	//verifyPoint = verifyPoint.x
	verifyPoint.X, verifyPoint.Y = privacy.Curve.ScalarMult(verifyPoint.X, verifyPoint.Y, xChallenge.Bytes())
	//verifyPoint = verifyPoint + B
	verifyPoint.X, verifyPoint.Y = privacy.Curve.Add(verifyPoint.X, verifyPoint.Y, pro.commitmentZeroS.X, pro.commitmentZeroS.Y)

	//Generate Zero number
	zeroInt := big.NewInt(0)

	//Calculate comm_ck(0,z, Index)
	commitmentZeroZ := privacy.PedCom.CommitAtIndex(zeroInt, pro.Proof.z, pro.Proof.index)

	if commitmentZeroZ.X.CmpAbs(verifyPoint.X) != 0 {
		return false
	}
	if commitmentZeroZ.Y.CmpAbs(verifyPoint.Y) != 0 {
		return false
	}

	return true
}

//TestProofIsZero test prove and verify function
func TestProofIsZero() bool {
	return false
}
