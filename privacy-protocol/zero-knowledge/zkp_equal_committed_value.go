package zkp

import (
	"crypto/rand"
	"math/big"

	"github.com/ninjadotorg/constant/privacy-protocol"
)

// Protocol proving in ZK ... https://link.springer.com/chapter/10.1007/3-540-48910-X_8
// PKEqualityOfCommittedVal

// PKEqualityOfCommittedValProof contains...
type PKEqualityOfCommittedValProof struct {
	C     []*privacy.EllipticPoint //Statement
	Index []*byte                  //Statement
	T     []*privacy.EllipticPoint
	Z     []*big.Int
}

// PKEqualityOfCommittedValWitness contains...
type PKEqualityOfCommittedValWitness struct {
	C     []*privacy.EllipticPoint //Statement
	Index []*byte                  //Statement
	X     []*big.Int
}

// randValue ...
func (wit *PKEqualityOfCommittedValWitness) randValue() {
	X := make([]*big.Int, 3)
	for i := 0; i < 3; i++ {
		X[i], _ = rand.Int(rand.Reader, privacy.Curve.Params().N)
	}
	C := make([]*privacy.EllipticPoint, 2)
	index := make([]*byte, 2)

	index[0] = new(byte)
	*index[0] = 1
	index[1] = new(byte)
	*index[1] = 2
	for i := 0; i < 2; i++ {
		C[i] = privacy.PedCom.CommitAtIndex(X[0], X[i+1], *index[i])
		// C[1] = privacy.PedCom.CommitAtIndex(X[0], X[2], 1)
	}
	wit.Set(C, index, X)
}

// Set - witness setter
func (wit *PKEqualityOfCommittedValWitness) Set(
	C []*privacy.EllipticPoint, //Statement
	Index []*byte, //Statement
	X []*big.Int) {
	wit.C = C
	wit.Index = Index
	wit.X = X
}

// Set - proof setter
func (pro *PKEqualityOfCommittedValProof) Set(
	C []*privacy.EllipticPoint, //Statement
	Index []*byte, //Statement
	T []*privacy.EllipticPoint,
	Z []*big.Int) {
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
		//t[i].Y = new(big.Int)
		t[i].X, t[i].Y = privacy.Curve.Add(privacy.PedCom.G[*wit.Index[i]].X, privacy.PedCom.G[*wit.Index[i]].Y, privacy.PedCom.G[3].X, privacy.PedCom.G[3].Y)
		t[i].X, t[i].Y = privacy.Curve.ScalarMult(t[i].X, t[i].Y, wRand.Bytes())
		// t[1].X, t[1].Y = privacy.Curve.Add(privacy.PedCom.G[*wit.Index[0]].X, privacy.PedCom.G[*wit.Index[0]].Y, privacy.PedCom.G[*wit.Index[3]].X, privacy.PedCom.G[*wit.Index[3]].Y)
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
		rightPoint.X, rightPoint.Y = privacy.Curve.ScalarMult(privacy.PedCom.G[*pro.Index[i]].X, privacy.PedCom.G[*pro.Index[i]].Y, pro.Z[0].Bytes())
		tmpPoint := new(privacy.EllipticPoint)
		tmpPoint.X, tmpPoint.Y = privacy.Curve.ScalarMult(privacy.PedCom.G[3].X, privacy.PedCom.G[3].Y, pro.Z[i+1].Bytes())
		rightPoint.X, rightPoint.Y = privacy.Curve.Add(rightPoint.X, rightPoint.Y, tmpPoint.X, tmpPoint.Y)
		tmpPoint.X, tmpPoint.Y = privacy.Curve.ScalarMult(pro.C[i].X, pro.C[i].Y, xChallenge.Bytes())
		rightPoint.X, rightPoint.Y = privacy.Curve.Add(rightPoint.X, rightPoint.Y, tmpPoint.X, tmpPoint.Y)
		if !rightPoint.IsEqual(*pro.T[i]) {
			return false
		}
	}
	return true
}

// TestPKEqualityOfCommittedVal ...
func TestPKEqualityOfCommittedVal() bool {
	witness := new(PKEqualityOfCommittedValWitness)
	witness.randValue()
	proof := new(PKEqualityOfCommittedValProof)
	proof = witness.Prove()
	return proof.Verify()
}
