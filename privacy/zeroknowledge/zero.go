package zkp

import (
	"errors"
	"math/big"

	"github.com/ninjadotorg/constant/privacy"
)

// ComZeroProof contains Proof's value
type ComZeroProof struct {
	commitmentValue *privacy.EllipticPoint //statement
	index           *byte                  //statement
	commitmentZeroS *privacy.EllipticPoint
	z               *big.Int
}

// ComZeroWitness contains Witness's value
type ComZeroWitness struct {
	commitmentValue *privacy.EllipticPoint //statement
	index           *byte                  //statement
	commitmentRnd   *big.Int
}

//Protocol for opening a commitment to 0 https://link.springer.com/chapter/10.1007/978-3-319-43005-8_1 (Fig. 5)
func (pro *ComZeroProof) Init() *ComZeroProof {
	pro.index = new(byte)
	pro.commitmentValue = new(privacy.EllipticPoint).Zero()
	pro.commitmentZeroS = new(privacy.EllipticPoint).Zero()
	pro.z = new(big.Int)
	return pro
}

func (pro *ComZeroProof) isNil() bool {
	if (pro.commitmentValue == nil) || (pro.commitmentZeroS == nil) || (pro.index == nil) || (pro.z == nil) {
		return true
	}
	return false
}

// Set dosomethings
func (wit *ComZeroWitness) Set(
	commitmentValue *privacy.EllipticPoint, //statement
	index *byte,                            //statement
	commitmentRnd *big.Int) {
	if wit == nil {
		wit = new(ComZeroWitness)
	}

	wit.commitmentRnd = commitmentRnd
	wit.commitmentValue = commitmentValue
	wit.index = index
}

// Bytes ...
func (pro ComZeroProof) Bytes() []byte {
	if pro.isNil() {
		return []byte{}
	}
	var res []byte
	// don't need to send commitmentValue, because verifiers can recalculate.
	res = append(res, pro.commitmentZeroS.Compress()...)
	res = append(res, privacy.AddPaddingBigInt(pro.z, privacy.BigIntSize)...)
	res = append(res, *pro.index)
	return res
}

// SetBytes ...
func (pro *ComZeroProof) SetBytes(bytes []byte) error {
	if pro == nil {
		pro = pro.Init()
	}

	if len(bytes) == 0 {
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

	offset := 0
	//err := pro.commitmentValue.Decompress(bytes[offset : offset + privacy.CompressedPointSize])
	//if err != nil {
	//	return errors.New("Decompressed failed!")
	//}
	//offset += privacy.CompressedPointSize

	err := pro.commitmentZeroS.Decompress(bytes[offset: offset+privacy.CompressedPointSize])
	if err != nil {
		return errors.New("Decompressed failed!")
	}
	offset += privacy.CompressedPointSize

	pro.z.SetBytes(bytes[offset: offset+privacy.BigIntSize])
	offset += privacy.BigIntSize

	*pro.index = bytes[offset]
	return nil
}

// Set dosomethings
func (pro *ComZeroProof) Set(
	commitmentValue *privacy.EllipticPoint, //statement
	index *byte,                            //statement
	commitmentZeroS *privacy.EllipticPoint,
	z *big.Int) {

	if pro == nil {
		pro = new(ComZeroProof)
	}
	pro.commitmentValue = commitmentValue
	pro.commitmentZeroS = commitmentZeroS
	pro.index = index
	pro.z = z
}

//Prove generate a Proof prove that the PedersenCommitment is zero
func (wit ComZeroWitness) Prove() (*ComZeroProof, error) {
	//var x big.Int
	//s is a random number in Zp, with p is N, which is order of base point of privacy.Curve
	sRnd := privacy.RandInt()

	//Calculate B = commitmentZeroS = comm_ck(0,s,Index)
	commitmentZeroS := privacy.PedCom.CommitAtIndex(big.NewInt(0), sRnd, *wit.index)

	//Generate challenge x in Zp
	xChallenge := generateChallengeFromPoint([]*privacy.EllipticPoint{wit.commitmentValue})

	//Calculate z=r*x + s (mod N)
	z := new(big.Int).Mul(wit.commitmentRnd, xChallenge)
	z.Add(z, sRnd)
	z.Mod(z, privacy.Curve.Params().N)

	proof := new(ComZeroProof)
	proof.Set(wit.commitmentValue, wit.index, commitmentZeroS, z)
	return proof, nil
}

//Verify verify that under PedersenCommitment is zero
func (pro *ComZeroProof) Verify() bool {
	//Generate challenge x in Zp
	xChallenge := generateChallengeFromPoint([]*privacy.EllipticPoint{pro.commitmentValue})

	//verifyPoint is result of A.x + B (in ECC)
	verifyPoint := pro.commitmentZeroS.Add(pro.commitmentValue.ScalarMult(xChallenge))

	//Generate Zero number
	zeroInt := big.NewInt(0)

	//Calculate comm_ck(0,z, Index)
	commitmentZeroZ := privacy.PedCom.CommitAtIndex(zeroInt, pro.z, *pro.index)

	if !commitmentZeroZ.IsEqual(verifyPoint) {
		return false
	}

	return true
}
