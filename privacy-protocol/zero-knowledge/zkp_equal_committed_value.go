package zkp

import (
	"crypto/rand"
	"errors"
	"math/big"

	privacy "github.com/ninjadotorg/constant/privacy-protocol"
)

// Protocol proving in ZK ... https://link.springer.com/chapter/10.1007/3-540-48910-X_8
// PKEqualityOfCommittedVal

// PKEqualityOfCommittedValProof contains...
type PKEqualityOfCommittedValProof struct {
	C     []*privacy.EllipticPoint //Statement    // 2*sz(EcPoint)
	Index []byte                   //Statement		// 2*sz(byte)
	T     []*privacy.EllipticPoint
	Z     []*big.Int
}

// PKEqualityOfCommittedValWitness contains...
type PKEqualityOfCommittedValWitness struct {
	C     []*privacy.EllipticPoint //Statement
	Index []byte                   //Statement
	X     []*big.Int
}

// Init create PKEqualityOfCommittedValProof element with default value
func (pro *PKEqualityOfCommittedValProof) Init() *PKEqualityOfCommittedValProof {
	return &PKEqualityOfCommittedValProof{
		C:     []*privacy.EllipticPoint{},
		Index: []byte{},
		T:     []*privacy.EllipticPoint{},
		Z:     []*big.Int{},
	}
}

// IsNil return true if one of each field is null
func (pro *PKEqualityOfCommittedValProof) IsNil() bool {
	if (len(pro.C) == 0) || (len(pro.Index) == 0) || (pro.T == nil) || (pro.Z == nil) {
		return true
	}
	return false
}

// randValue ...
func (wit *PKEqualityOfCommittedValWitness) randValue() {
	X := make([]*big.Int, 3)
	for i := 0; i < 3; i++ {
		X[i], _ = rand.Int(rand.Reader, privacy.Curve.Params().N)
	}
	C := make([]*privacy.EllipticPoint, 2)
	index := make([]byte, 2)

	index[0] = 1
	index[1] = 2
	for i := 0; i < 2; i++ {
		C[i] = privacy.PedCom.CommitAtIndex(X[0], X[i+1], index[i])
		// C[1] = privacy.PedCom.CommitAtIndex(X[0], X[2], 1)
	}
	wit.Set(C, index, X)
}

// Set - witness setter
func (wit *PKEqualityOfCommittedValWitness) Set(
	C []*privacy.EllipticPoint, //Statement
	Index []byte, //Statement
	X []*big.Int) {

	if wit == nil {
		wit = new(PKEqualityOfCommittedValWitness)
	}
	wit.C = C
	wit.Index = Index
	wit.X = X
}

// Bytes convert PKEqualityOfCommittedValProof's value to byte array and return
func (pro PKEqualityOfCommittedValProof) Bytes() []byte {
	if pro.IsNil() {
		return []byte{}
	}
	var res []byte

	res = append(pro.C[0].Compress(), pro.C[1].Compress()...)
	res = append(res, []byte{pro.Index[0], pro.Index[1]}...)

	for i := 0; i < len(pro.T); i++ {
		res = append(res, pro.T[i].Compress()...)
	}
	for i := 0; i < len(pro.Z); i++ {
		temp := pro.Z[i].Bytes()
		for j := 0; j < privacy.BigIntSize-len(temp); j++ {
			temp = append([]byte{0}, temp...)
		}
		res = append(res, temp...)
	}
	return res
}

// SetBytes convert byte array to PKEqualityOfCommittedValProof
func (pro *PKEqualityOfCommittedValProof) SetBytes(proofbytes []byte) error {
	if pro == nil {
		pro = pro.Init()
	}

	if len(proofbytes) == 0 {
		return nil
	}
	pro.C = make([]*privacy.EllipticPoint, 2)
	for i := 0; i < len(pro.C); i++ {
		pro.C[i] = new(privacy.EllipticPoint)
		err := pro.C[i].Decompress(proofbytes[i*privacy.CompressedPointSize : (i+1)*privacy.CompressedPointSize])
		if err != nil {
			return err
		}
		if !pro.C[i].IsSafe() {
			return errors.New("Decompressed failed!")
		}
	}
	pro.Index = make([]byte, 2)
	for i := 0; i < len(pro.Index); i++ {
		pro.Index[i] = proofbytes[i+len(pro.C)*privacy.CompressedPointSize]
	}

	pro.T = make([]*privacy.EllipticPoint, 2)
	for i := 0; i < len(pro.T); i++ {
		pro.T[i] = new(privacy.EllipticPoint)
		pro.T[i].Decompress(proofbytes[len(pro.Index)+len(pro.C)*privacy.CompressedPointSize+i*privacy.CompressedPointSize : len(pro.Index)+len(pro.C)*privacy.CompressedPointSize+(i+1)*privacy.CompressedPointSize])
		if !pro.T[i].IsSafe() {
			return errors.New("Decompressed failed!")
		}
	}
	pro.Z = make([]*big.Int, 3)
	for i := 0; i < len(pro.Z); i++ {
		pro.Z[i] = big.NewInt(0)
		pro.Z[i].SetBytes(proofbytes[len(pro.Index)+len(pro.C)*privacy.CompressedPointSize+len(pro.T)*privacy.CompressedPointSize+i*privacy.BigIntSize : len(pro.Index)+len(pro.C)*privacy.CompressedPointSize+len(pro.T)*privacy.CompressedPointSize+(i+1)*privacy.BigIntSize])
	}
	return nil
}

// Set - proof setter
func (pro *PKEqualityOfCommittedValProof) Set(
	C []*privacy.EllipticPoint, //Statement
	Index []byte, //Statement
	T []*privacy.EllipticPoint,
	Z []*big.Int) {

	if pro == nil {
		pro = new(PKEqualityOfCommittedValProof)
	}
	pro.C = C
	pro.Index = Index
	pro.T = T
	pro.Z = Z
}

// Prove ...
func (wit *PKEqualityOfCommittedValWitness) Prove() *PKEqualityOfCommittedValProof {
	wRand, _ := rand.Int(rand.Reader, privacy.Curve.Params().N)
	xChallenge := GenerateChallengeFromPoint(wit.C)
	z := make([]*big.Int, 3)
	for i := 0; i < 3; i++ {
		z[i] = big.NewInt(0).Mul(wit.X[i], xChallenge)
		z[i] = z[i].Sub(wRand, z[i])
		z[i] = z[i].Mod(z[i], privacy.Curve.Params().N)
	}
	t := make([]*privacy.EllipticPoint, 2)
	for i := 0; i < 2; i++ {
		t[i] = new(privacy.EllipticPoint)
		t[i].X, t[i].Y = privacy.Curve.Add(privacy.PedCom.G[wit.Index[i]].X, privacy.PedCom.G[wit.Index[i]].Y, privacy.PedCom.G[privacy.PedCom.Capacity-1].X, privacy.PedCom.G[privacy.PedCom.Capacity-1].Y)
		t[i].X, t[i].Y = privacy.Curve.ScalarMult(t[i].X, t[i].Y, wRand.Bytes())
	}
	proof := new(PKEqualityOfCommittedValProof)
	proof.Set(wit.C, wit.Index, t, z)
	return proof
}

// Verify ...
func (pro *PKEqualityOfCommittedValProof) Verify() bool {
	xChallenge := GenerateChallengeFromPoint(pro.C)
	for i := 0; i < 2; i++ {
		rightPoint := new(privacy.EllipticPoint)
		rightPoint.X, rightPoint.Y = privacy.Curve.ScalarMult(privacy.PedCom.G[pro.Index[i]].X, privacy.PedCom.G[pro.Index[i]].Y, pro.Z[0].Bytes())

		tmpPoint := new(privacy.EllipticPoint)

		tmpPoint.X, tmpPoint.Y = privacy.Curve.ScalarMult(privacy.PedCom.G[privacy.PedCom.Capacity-1].X, privacy.PedCom.G[privacy.PedCom.Capacity-1].Y, pro.Z[i+1].Bytes())
		rightPoint.X, rightPoint.Y = privacy.Curve.Add(rightPoint.X, rightPoint.Y, tmpPoint.X, tmpPoint.Y)

		tmpPoint.X, tmpPoint.Y = privacy.Curve.ScalarMult(pro.C[i].X, pro.C[i].Y, xChallenge.Bytes())
		rightPoint.X, rightPoint.Y = privacy.Curve.Add(rightPoint.X, rightPoint.Y, tmpPoint.X, tmpPoint.Y)
		if !rightPoint.IsEqual(pro.T[i]) {
			return false
		}
	}
	return true
}
