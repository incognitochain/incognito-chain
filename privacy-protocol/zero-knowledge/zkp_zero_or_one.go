package zkp

import (
	"fmt"
	"math/big"

	"github.com/ninjadotorg/constant/privacy-protocol"
)

// PKComZeroOneProtocol is a protocol for Zero-knowledge Proof of Knowledge of committed zero or one
// include Witness: commitedValue, r []byte
//type PKComZeroOneProtocol struct {
//	Witness PKComZeroOneWitness
//	Proof   PKComZeroOneProof
//}

// PKComZeroOneWitness contains Witness's value
type PKComZeroOneWitness struct {
	commitedValue *big.Int
	rand          *big.Int
	//general info
	commitment *privacy.EllipticPoint
	index      byte
}

// PKComZeroOneProof contains Proof's value
type PKComZeroOneProof struct {
	ca, cb    *privacy.EllipticPoint
	f, za, zb *big.Int
	//general info
	commitment *privacy.EllipticPoint
	index      byte
}

// Set sets Witness
func (wit *PKComZeroOneWitness) Set(
	commitedValue *big.Int,
	rand *big.Int,
	commitment *privacy.EllipticPoint,
	index byte) {

	wit.commitedValue = commitedValue
	wit.rand = rand
	wit.commitment = commitment
	wit.index = index
}

// Get returns witness
func (wit *PKComZeroOneWitness) Get() *PKComZeroOneWitness {
	return wit
}

// Set sets Proof
func (proof *PKComZeroOneProof) Set(
	ca, cb *privacy.EllipticPoint,
	f, za, zb *big.Int,
	commitment *privacy.EllipticPoint,
	index byte) {

	proof.ca, proof.cb = ca, cb
	proof.f, proof.za, proof.zb = f, za, zb
	proof.commitment = commitment
	proof.index = index
}

// Get returns proof
func (proof *PKComZeroOneProof) Get() *PKComZeroOneProof {
	return proof
}

func getindex(bigint *big.Int) int {
	return 32 - len(bigint.Bytes())
}

// Prove creates a Proof for PedersenCommitment to zero or one
func (wit *PKComZeroOneWitness) Prove() (*PKComZeroOneProof, error) {
	// Check index
	if wit.index < privacy.SK || wit.index > privacy.RAND {
		return nil, fmt.Errorf("index must be between SK index and RAND index")
	}

	// Check whether commited value is zero or one or not
	if wit.commitedValue.Cmp(big.NewInt(0)) != 0 && wit.commitedValue.Cmp(big.NewInt(0)) != 1 {
		return nil, fmt.Errorf("commited value must be zero or one")
	}

	proof := new(PKComZeroOneProof)

	// Generate random numbers
	a := new(big.Int).SetBytes(privacy.RandBytes(32))
	a.Mod(a, privacy.Curve.Params().N)

	s := new(big.Int).SetBytes(privacy.RandBytes(32))
	s.Mod(s, privacy.Curve.Params().N)

	t := new(big.Int).SetBytes(privacy.RandBytes(32))
	t.Mod(t, privacy.Curve.Params().N)

	// Calculate ca, cb
	proof.ca = privacy.PedCom.CommitAtIndex(a, s, wit.index)

	// Calulate am = a*commitedValue
	am := big.NewInt(0)
	am.Mul(a, wit.commitedValue)
	proof.cb = privacy.PedCom.CommitAtIndex(am, t, wit.index)

	// Calculate x = hash (G0||G1||G2||G3||ca||cb||cm)
	x := GenerateChallengeFromPoint([]*privacy.EllipticPoint{proof.ca, proof.cb, wit.commitment})
	x.Mod(x, privacy.Curve.Params().N)

	// Calculate f = mx + a
	proof.f = big.NewInt(0)
	proof.f.Mul(wit.commitedValue, x)
	proof.f.Add(proof.f, a)
	proof.f.Mod(proof.f, privacy.Curve.Params().N)

	// Calculate za = rx + s
	proof.za = big.NewInt(0)
	proof.za.Mul(wit.rand, x)
	proof.za.Add(proof.za, s)
	proof.za.Mod(proof.za, privacy.Curve.Params().N)

	// Calculate zb = r(x-f) + t
	proof.zb = big.NewInt(1)
	xSubF := new(big.Int).Sub(x, proof.f)
	xSubF.Mod(xSubF, privacy.Curve.Params().N)
	proof.zb.Mul(wit.rand, xSubF)
	proof.zb.Add(proof.zb, t)
	proof.zb.Mod(proof.zb, privacy.Curve.Params().N)

	proof.commitment = wit.commitment
	proof.index = wit.index

	//
	//xm := big.NewInt(1)
	//	//xm.Mul(x, new(big.Int).SetBytes(wit.Witness.CommitedValue))
	//	//point := privacy.EllipticPoint{big.NewInt(0), big.NewInt(0)}
	//	//point.X, point.Y = privacy.PedCom.G[wit.Witness.Index].X, privacy.PedCom.G[wit.Witness.Index].Y
	//	//point.X, point.Y = privacy.Curve.ScalarMult(point.X, point.Y, xm.Bytes())
	//	//
	//	//xr := big.NewInt(1)
	//	//xr.Mul(x, new(big.Int).SetBytes(wit.Witness.Rand))
	//	//point2 := privacy.EllipticPoint{big.NewInt(0), big.NewInt(0)}
	//	//point2.X, point2.Y = privacy.PedCom.G[privacy.CM_CAPACITY-1].X, privacy.PedCom.G[privacy.CM_CAPACITY-1].Y
	//	//point2.X, point2.Y = privacy.Curve.ScalarMult(point2.X, point2.Y, xr.Bytes())
	//	//
	//	//point.X, point.Y = privacy.Curve.Add(point.X, point.Y, point2.X, point2.Y)
	//	//
	//	//aG := privacy.PedCom.G[wit.Witness.Index]
	//	//aG.X, aG.Y = privacy.Curve.ScalarMult(aG.X, aG.Y, a.Bytes())
	//	//sH := privacy.PedCom.G[privacy.CM_CAPACITY-1]
	//	//sH.X, sH.Y = privacy.Curve.ScalarMult(sH.X, sH.Y, s.Bytes())
	//	//aG.X, aG.Y = privacy.Curve.Add(aG.X, aG.Y, sH.X, sH.Y)
	//	//aG.X, aG.Y = privacy.Curve.Add(aG.X, aG.Y, point.X, point.Y)
	//	//
	//	//proof.commitment = wit.Witness.Commitment.CompressPoint()
	//	//proof.index = wit.Witness.Index

	return proof, nil
}

// Verify verifies the Proof for PedersenCommitment to zero or one
func (proof *PKComZeroOneProof) Verify() bool {
	////Decompress PedersenCommitment  value
	//comPoint, err := privacy.DecompressCommitment(proof.Proof.commitment)
	//if err != nil {
	//	fmt.Printf("Can not decompress PedersenCommitment value to ECC point")
	//	return false
	//}

	// Calculate x = hash (G0||G1||G2||G3||ca||cb||cm)

	fmt.Printf("verify proof ca: %v\n", proof.ca)
	fmt.Printf("verify proof ca: %v\n", proof.cb)
	fmt.Printf("verify proof ca: %v\n", proof.commitment)

	x := GenerateChallengeFromPoint([]*privacy.EllipticPoint{proof.ca, proof.cb, proof.commitment})
	x.Mod(x, privacy.Curve.Params().N)

	// Calculate leftPoint1 = c^x * ca
	leftPoint1 := privacy.EllipticPoint{big.NewInt(0), big.NewInt(0)}
	leftPoint1.X, leftPoint1.Y = privacy.Curve.ScalarMult(proof.commitment.X, proof.commitment.Y, x.Bytes())
	leftPoint1.X, leftPoint1.Y = privacy.Curve.Add(leftPoint1.X, leftPoint1.Y, proof.ca.X, proof.ca.Y)

	// Calculate rightPoint1 = Com(f, za)
	rightPoint1 := privacy.PedCom.CommitAtIndex(proof.f, proof.za, proof.index)

	// Calculate leftPoint2 = c^(x-f) * cb
	leftPoint2 := privacy.EllipticPoint{big.NewInt(0), big.NewInt(0)}
	xSubF := new(big.Int)
	xSubF.Sub(x, proof.f)
	xSubF.Mod(xSubF, privacy.Curve.Params().N)
	leftPoint2.X, leftPoint2.Y = privacy.Curve.ScalarMult(proof.commitment.X, proof.commitment.Y, xSubF.Bytes())
	leftPoint2.X, leftPoint2.Y = privacy.Curve.Add(leftPoint2.X, leftPoint2.Y, proof.cb.X, proof.cb.Y)

	// Calculate rightPoint1 = Com(0, zb)
	rightPoint2 := privacy.PedCom.CommitAtIndex(big.NewInt(0), proof.zb, proof.index)

	if leftPoint1.X.Cmp(rightPoint1.X) == 0 && leftPoint1.Y.Cmp(rightPoint1.Y) == 0 && leftPoint2.X.Cmp(rightPoint2.X) == 0 && leftPoint2.Y.Cmp(rightPoint2.Y) == 0 {
		return true
	}

	return false
}

// TestPKComZeroOne tests prove and verify function for PK for PedersenCommitment to zero or one
func TestPKComZeroOne() {
	res := true
	for res {
		// generate Openings
		valueRand := privacy.RandBytes(32)
		vInt := new(big.Int).SetBytes(valueRand)
		vInt.Mod(vInt, big.NewInt(2))
		rand := new(big.Int).SetBytes(privacy.RandBytes(32))

		// CommitAll
		cm := privacy.PedCom.CommitAtIndex(vInt, rand, privacy.VALUE)

		// create witness for proving
		var witness PKComZeroOneWitness
		witness.Set(vInt, rand, cm, privacy.VALUE)

		// Proving
		proof, _ := witness.Prove()
		fmt.Printf("Proof: %+v\n", proof)

		// Set proof for verifying
		Proof := new(PKComZeroOneProof)
		Proof.Set(proof.ca, proof.cb, proof.f, proof.za, proof.zb, proof.commitment, proof.index)

		res = Proof.Verify()
		fmt.Println(res)
	}
}
