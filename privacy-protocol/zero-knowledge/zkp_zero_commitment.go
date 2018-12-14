package zkp

import (
	"crypto/rand"
	"errors"
	"math/big"

	privacy "github.com/ninjadotorg/constant/privacy-protocol"
)

// PKComZeroProof contains Proof's value
type PKComZeroProof struct {
	commitmentValue *privacy.EllipticPoint //statement
	index           *byte                  //statement
	commitmentZeroS *privacy.EllipticPoint
	z               *big.Int
}

// PKComZeroWitness contains Witness's value
type PKComZeroWitness struct {
	commitmentValue *privacy.EllipticPoint //statement
	index           *byte                  //statement
	commitmentRnd   *big.Int
}

//Protocol for opening a commitment to 0 https://link.springer.com/chapter/10.1007/978-3-319-43005-8_1 (Fig. 5)

/*Protocol for opening a PedersenCommitment to 0
Prove:
	commitmentValue is PedersenCommitment value of Zero, that is statement needed to prove
	commitmentValue is calculated by Comm_ck(H,PRDNumber)
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
	commitmentValue is calculated by Comm_ck(H,PRDNumber), a.k.a A
	commitmentZeroS, z are output of Prove function, commitmentZeroS is a.k.a B
	x <- Hash(G0||G1||G2||G3||commitmentvalue)
	boolValue <- (Comm_ck(0,z) == A.x + B); in this case, A and B needed to convert to privacy.privacy.EllipticPoint
	return boolValue
)
*/
func (pro *PKComZeroProof) Init() *PKComZeroProof {
	pro.index = new(byte)
	pro.commitmentValue = new(privacy.EllipticPoint).Zero()
	pro.commitmentZeroS = new(privacy.EllipticPoint).Zero()
	pro.z = new(big.Int)
	return pro
}

func (pro *PKComZeroProof) IsNil() bool {
	if (pro.commitmentValue == nil) || (pro.commitmentZeroS == nil) || (pro.index == nil) || (pro.z == nil) {
		return true
	}
	return false
}

// randValue return random witness value for testing
func (wit *PKComZeroWitness) randValue(testcase bool) {
	switch testcase {
	case false:
		commitmentValue := new(privacy.EllipticPoint)
		commitmentValue.Randomize()
		index := byte(3)
		commitmentRnd, _ := rand.Int(rand.Reader, privacy.Curve.Params().N)
		wit.Set(commitmentValue, &index, commitmentRnd)
		break
	case true:
		// commitmentValue := new(privacy.EllipticPoint)
		// commitmentValue.Randomize()
		index := byte(3)
		commitmentRnd, _ := rand.Int(rand.Reader, privacy.Curve.Params().N)
		commitmentValue := privacy.PedCom.CommitAtIndex(big.NewInt(0), commitmentRnd, index)
		wit.Set(commitmentValue, &index, commitmentRnd)
		break
	}

}

// Set dosomethings
func (wit *PKComZeroWitness) Set(
	commitmentValue *privacy.EllipticPoint, //statement
	index *byte, //statement
	commitmentRnd *big.Int) {
	if wit == nil {
		wit = new(PKComZeroWitness)
	}

	wit.commitmentRnd = commitmentRnd
	wit.commitmentValue = commitmentValue
	wit.index = index
}

// Bytes ...
func (pro PKComZeroProof) Bytes() []byte {
	if pro.IsNil() {
		return []byte{}
	}
	var res []byte
	res = append(pro.commitmentValue.Compress(), pro.commitmentZeroS.Compress()...)

	temp := pro.z.Bytes()
	for j := 0; j < privacy.BigIntSize-len(temp); j++ {
		temp = append([]byte{0}, temp...)
	}
	res = append(res, temp...)
	//res = append(res, *pro.index)
	res = append(res, []byte{*pro.index}...)
	return res
}

// SetBytes ...
func (pro *PKComZeroProof) SetBytes(bytestr []byte) error {
	if pro == nil {
		pro = pro.Init()
	}

	if len(bytestr) == 0 {
		return nil
	}
	if pro.commitmentValue == nil {
		pro.commitmentValue = new(privacy.EllipticPoint)
	}
	if pro.commitmentZeroS == nil {
		pro.commitmentZeroS = new(privacy.EllipticPoint)
	}
	if pro.z == nil {
		pro.z = big.NewInt(0)
	}
	if pro.index == nil {
		pro.index = new(byte)
	}

	err := pro.commitmentValue.Decompress(bytestr[0:privacy.CompressedPointSize])
	if err != nil {
		return errors.New("Decompressed failed!")
	}
	err = pro.commitmentZeroS.Decompress(bytestr[privacy.CompressedPointSize : 2*privacy.CompressedPointSize])
	if err != nil {
		return errors.New("Decompressed failed!")
	}
	pro.z.SetBytes(bytestr[2*privacy.CompressedPointSize : 2*privacy.CompressedPointSize+privacy.BigIntSize])
	*pro.index = bytestr[2*privacy.CompressedPointSize+privacy.BigIntSize]
	return nil
}

// Set dosomethings
func (pro *PKComZeroProof) Set(
	commitmentValue *privacy.EllipticPoint, //statement
	index *byte, //statement
	commitmentZeroS *privacy.EllipticPoint,
	z *big.Int) {

	if pro == nil {
		pro = new(PKComZeroProof)
	}
	pro.commitmentValue = commitmentValue
	pro.commitmentZeroS = commitmentZeroS
	pro.index = index
	pro.z = z
}

//Prove generate a Proof prove that the PedersenCommitment is zero
func (wit PKComZeroWitness) Prove() (*PKComZeroProof, error) {
	//var x big.Int
	//s is a random number in Zp, with p is N, which is order of base point of privacy.Curve
	sRnd, _ := rand.Int(rand.Reader, privacy.Curve.Params().N)

	//Calculate B = commitmentZeroS = comm_ck(0,s,Index)
	commitmentZeroS := privacy.PedCom.CommitAtIndex(big.NewInt(0), sRnd, *wit.index)

	//Generate challenge x in Zp
	xChallenge := GenerateChallengeFromPoint([]*privacy.EllipticPoint{wit.commitmentValue})

	//Calculate z=r*x + s (mod N)
	z := big.NewInt(0)
	z.Set(wit.commitmentRnd)
	z.Mul(z, xChallenge)
	z.Mod(z, privacy.Curve.Params().N)
	z.Add(z, sRnd)
	z.Mod(z, privacy.Curve.Params().N)

	proof := new(PKComZeroProof)
	proof.Set(wit.commitmentValue, wit.index, commitmentZeroS, z)
	return proof, nil
}

//Verify verify that under PedersenCommitment is zero
func (pro *PKComZeroProof) Verify() bool {
	//Generate challenge x in Zp
	xChallenge := GenerateChallengeFromPoint([]*privacy.EllipticPoint{pro.commitmentValue})

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
	verifyPoint.X.Set(pro.commitmentValue.X)
	verifyPoint.Y.Set(pro.commitmentValue.Y)
	//verifyPoint = verifyPoint.x
	verifyPoint.X, verifyPoint.Y = privacy.Curve.ScalarMult(verifyPoint.X, verifyPoint.Y, xChallenge.Bytes())
	//verifyPoint = verifyPoint + B
	verifyPoint.X, verifyPoint.Y = privacy.Curve.Add(verifyPoint.X, verifyPoint.Y, pro.commitmentZeroS.X, pro.commitmentZeroS.Y)

	//Generate Zero number
	zeroInt := big.NewInt(0)

	//Calculate comm_ck(0,z, Index)
	commitmentZeroZ := privacy.PedCom.CommitAtIndex(zeroInt, pro.z, *pro.index)

	if commitmentZeroZ.X.CmpAbs(verifyPoint.X) != 0 {
		return false
	}
	if commitmentZeroZ.Y.CmpAbs(verifyPoint.Y) != 0 {
		return false
	}

	return true
}
