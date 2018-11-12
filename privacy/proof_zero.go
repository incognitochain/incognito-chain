package privacy

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

/*Protocol for opening a commitment to 0
Prove:
	commitmentValue is commitment value of Zero, that is statement needed to prove
	commitmentValue is calculated by Comm_ck(Value,PRDNumber)
	commitmentRnd is PRDNumber, which is used to calculate commitmentValue
	s <- Zp; P is Curve base point's order, is N
	B <- Comm_ck(0,s);  Comm_ck is PedersenCommit function using public params - Curve.Params() (G0,G1...)
						but is just commit special value (in this case, special value is 0),
						which is stick with G[index] (in this case, index is the index stick with commitmentValue)
						B is a.k.a commitmentZeroS
	x <- Hash(G0||G1||G2||G3||commitmentvalue) x is pseudorandom number, which could be computed easily by Verifier
	z <- rx + s; z in Zp, r is commitmentRnd
	return commitmentZeroS, z

Verify:
	commitmentValue is commitment value of Zero, that is statement needed to prove
	commitmentValue is calculated by Comm_ck(Value,PRDNumber), a.k.a A
	commitmentZeroS, z are output of Prove function, commitmentZeroS is a.k.a B
	x <- Hash(G0||G1||G2||G3||commitmentvalue)
	boolValue <- (Comm_ck(0,z) == A.x + B); in this case, A and B needed to convert to EllipticPoint
	return boolValue
)
*/

//ProveIsZero generate a proof prove that the commitment is zero
func ProveIsZero(commitmentValue, commitmentRnd []byte, index byte) ([]byte, *big.Int) {
	//var x big.Int
	//s is a random number in Zp, with p is N, which is order of base point of Curve
	sRnd, err := rand.Int(rand.Reader, Curve.Params().N)
	if err != nil {
		panic(err)
	}

	//Generate zero number to commit
	zeroInt := big.NewInt(0)

	//Calculate B = commitmentZeroS = comm_ck(0,s,index)
	commitmentZeroS := Pcm.CommitSpecValue(zeroInt.Bytes(), sRnd.Bytes(), index)

	//Generate random x in Zp
	xRnd := big.NewInt(0)
	xRnd.SetBytes(Pcm.getHashOfValues([][]byte{commitmentValue}))
	xRnd.Mod(xRnd, Curve.Params().N)

	//Calculate z=r*x + s (mod N)
	z := big.NewInt(0)
	z.SetBytes(commitmentRnd)
	z.Mul(z, xRnd)
	z.Mod(z, Curve.Params().N)
	z.Add(z, sRnd)
	z.Mod(z, Curve.Params().N)

	//return B, z
	return commitmentZeroS, z
}

//VerifyIsZero verify that under commitment is zero
func VerifyIsZero(commitmentValue, commitmentZeroS []byte, index byte, z *big.Int) bool {
	//Calculate x
	xRnd := big.NewInt(0)
	xRnd.SetBytes(Pcm.getHashOfValues([][]byte{commitmentValue}))
	xRnd.Mod(xRnd, Curve.Params().N)

	//convert commitmentValue []byte to Point in ECC
	commitmentValuePoint, err := DecompressKey(commitmentValue)
	if err != nil {
		return false
	}
	if (!Curve.IsOnCurve(commitmentValuePoint.X, commitmentValuePoint.Y)) || (z.Cmp(Curve.Params().N) > -1) {
		return false
	}

	//convert commitmentZeroS (a.k.a B) to Point in ECC
	commitmentZeroSPoint, err := DecompressCommitment(commitmentZeroS)
	if err != nil {
		return false
	}
	if (!Curve.IsOnCurve(commitmentZeroSPoint.X, commitmentZeroSPoint.Y)) || (z.Cmp(Curve.Params().N) > -1) {
		return false
	}

	//verifyPoint is result of A.x + B (in ECC)
	verifyPoint := new(EllipticPoint)
	verifyPoint.X = big.NewInt(0)
	verifyPoint.Y = big.NewInt(0)
	//Set verifyPoint = A
	verifyPoint.X.SetBytes(commitmentValuePoint.X.Bytes())
	verifyPoint.Y.SetBytes(commitmentValuePoint.Y.Bytes())
	//verifyPoint = verifyPoint.x
	verifyPoint.X, verifyPoint.Y = Curve.ScalarMult(verifyPoint.X, verifyPoint.Y, xRnd.Bytes())
	//verifyPoint = verifyPoint + B
	verifyPoint.X, verifyPoint.Y = Curve.Add(verifyPoint.X, verifyPoint.Y, commitmentZeroSPoint.X, commitmentZeroSPoint.Y)

	//Generate Zero number
	zeroInt := big.NewInt(0)

	//Calculate comm_ck(0,z, index)
	commitmentZeroZ := Pcm.CommitSpecValue(zeroInt.Bytes(), z.Bytes(), index)

	//convert result to point
	commitmentZeroZPoint, err := DecompressCommitment(commitmentZeroZ)
	if err != nil {
		return false
	}
	if (!Curve.IsOnCurve(commitmentZeroZPoint.X, commitmentZeroZPoint.Y)) || (z.Cmp(Curve.Params().N) > -1) {
		return false
	}

	if commitmentZeroZPoint.X.CmpAbs(verifyPoint.X) != 0 {
		return false
	}
	if commitmentZeroZPoint.Y.CmpAbs(verifyPoint.Y) != 0 {
		return false
	}

	return true
}

//TestProofIsZero test prove and verify function
func TestProofIsZero() bool {
	//Generate a random commitment

	//First, generate random value to commit and calculate two commitment with different PRDNumber
	//Random value
	serialNumber := RandBytes(32)

	//Random two PRDNumber in Zp
	r1Int := big.NewInt(0)
	r2Int := big.NewInt(0)
	r1 := RandBytes(32)
	r2 := RandBytes(32)
	r1Int.SetBytes(r1)
	r2Int.SetBytes(r2)
	r1Int.Mod(r1Int, Curve.Params().N)
	r2Int.Mod(r2Int, Curve.Params().N)
	r1 = r1Int.Bytes()
	r2 = r2Int.Bytes()

	//Calculate two Pedersen commitment
	committemp1 := Pcm.CommitSpecValue(serialNumber, r1, 0)
	committemp2 := Pcm.CommitSpecValue(serialNumber, r2, 0)

	//Converting them to ECC Point
	committemp1Point, err := DecompressCommitment(committemp1)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	committemp2Point, err := DecompressCommitment(committemp2)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	//Compute inverse of commitment2 cuz we wanna calculate A1 + A2^-1 in ECC
	//Inverse of A(x,y) in ECC is A'(x,P-y) with P is order of field
	inverse_committemp2Point := new(EllipticPoint)
	inverse_committemp2Point.X = big.NewInt(0)
	inverse_committemp2Point.Y = big.NewInt(0)
	inverse_committemp2Point.X.SetBytes(committemp2Point.X.Bytes())
	inverse_committemp2Point.Y.SetBytes(committemp2Point.Y.Bytes())
	inverse_committemp2Point.Y.Sub(Curve.Params().P, committemp2Point.Y)

	//So, when we have A1+A2^-1, we need compute r = r1 - r2 (mod N), which is r of zero commitment
	rInt := big.NewInt(0)
	rInt.Sub(r1Int, r2Int)
	rInt.Mod(rInt, Curve.Params().N)

	//Convert result of A1 + A2^-1 to ECC Point
	resPoint := EllipticPoint{big.NewInt(0), big.NewInt(0)}
	resPoint.X, resPoint.Y = Curve.Add(committemp1Point.X, committemp1Point.Y, inverse_committemp2Point.X, inverse_committemp2Point.Y)

	//Convert it to byte array
	commitZero := CompressKey(resPoint)

	//Compute proof
	proofZero, z := ProveIsZero(commitZero, rInt.Bytes(), 0)

	//verify proof
	boolValue := VerifyIsZero(commitZero, proofZero, 0, z)
	fmt.Println("Test ProofIsZero resulit: ", boolValue)
	return boolValue
}
