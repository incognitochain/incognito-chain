package privacy

import (
	"fmt"
	"math/big"

	"github.com/minio/blake2b-simd"
)

// ZKProtocols interface
type ZKProtocols interface {
	SetWitness(witnesses [][]byte)
	Prove() ([]byte, error)
	Verify() bool
}

// PKComValProtocol is a protocol for Zero-knowledge Proof of Knowledge of committed values
// include witnesses
type PKComValProtocol struct {
	witnesses [][]byte
	// Proof     *ProofForPKCommittedValues
}

// PKComValProof contains proof's value
type PKComValProof struct {
	Alpha  []byte
	Gammas [][]byte
}

// SetWitness sets witnesses
func (pro *PKComValProtocol) SetWitness(witnesses [][]byte) {
	pro.witnesses = make([][]byte, len(witnesses))
	for i := 0; i < len(witnesses); i++ {
		pro.witnesses[i] = make([]byte, len(witnesses[i]))
		copy(pro.witnesses[i], witnesses[i])
	}
}

// Prove creates zero knowledge proof for an opening of a Pedersen commitment
func (pro *PKComValProtocol) Prove(commitmentValue []byte) (*PKComValProof, error) {
	if len(pro.witnesses) != 4 {
		return nil, fmt.Errorf("len of witnesses must be equal to 4")
	}

	proof := new(PKComValProof)

	// Calculate random numbers
	r := make([][]byte, CM_CAPACITY)
	for i := 0; i < CM_CAPACITY; i++ {
		r[i] = RandBytes(32)
	}

	// Calculate alpha
	alpha := new(EllipticPoint)
	tmp := new(EllipticPoint)
	alpha.X, alpha.Y = Curve.ScalarMult(Pcm.G[0].X, Pcm.G[0].Y, r[0])
	for i := 1; i < CM_CAPACITY; i++ {
		tmp.X, tmp.Y = Curve.ScalarMult(Pcm.G[i].X, Pcm.G[i].Y, r[i])
		alpha.X, alpha.Y = Curve.Add(alpha.X, alpha.Y, tmp.X, tmp.Y)
	}

	proof.Alpha = make([]byte, 33)
	copy(proof.Alpha, CompressKey(*alpha))

	// calculate beta
	hashFunc := blake2b.New256()
	values := [][]byte{
		commitmentValue,
		CompressKey((*alpha)),
	}

	appendStr := Pcm.getHashOfValues(values)
	hashFunc.Write(appendStr)
	beta := hashFunc.Sum(nil)

	// Calculate gammas
	b := new(big.Int)
	witness := new(big.Int)
	bMulWitness := new(big.Int)
	randTmp := new(big.Int)

	b.SetBytes(beta)
	proof.Gammas = make([][]byte, CM_CAPACITY)

	for i := 0; i < CM_CAPACITY; i++ {
		witness.SetBytes(pro.witnesses[i])
		bMulWitness.Mul(b, witness)
		proof.Gammas[i] = make([]byte, 32)
		copy(proof.Gammas[i], bMulWitness.Add(bMulWitness, randTmp.SetBytes(r[i])).Bytes())
	}
	return proof, nil
}

//Verify check the proof's value
//in this function, we check c
func (pro *PKComValProtocol) Verify(proof PKComValProof, commitmentValue []byte) bool {
	// re-calculate beta and check whether it is equal to beta of proof or not
	beta := Pcm.getHashOfValues([][]byte{commitmentValue, proof.Alpha})

	// Calculate right point:
	rightPoint := EllipticPoint{big.NewInt(0), big.NewInt(0)}
	tmpPoint := new(EllipticPoint)
	for i := 0; i < CM_CAPACITY; i++ {
		tmpPoint.X, tmpPoint.Y = Curve.ScalarMult(Pcm.G[i].X, Pcm.G[i].Y, proof.Gammas[i])
		rightPoint.X, rightPoint.Y = Curve.Add(rightPoint.X, rightPoint.Y, tmpPoint.X, tmpPoint.Y)
	}

	//Logger.log.Infof("commitment value: %v\n", commitmentValue)
	commitmentPoint, err := DecompressCommitment(commitmentValue)
	if err != nil {
		//	Logger.log.Errorf("Decompress commitment error: %v\n", err.Error())
	}

	alphaPoint, err := DecompressKey(proof.Alpha)
	if err != nil {
		//	Logger.log.Errorf("Decompress alpha error: %v\n", err.Error())
	}
	// Calculate left point:
	xY, yY := Curve.ScalarMult(commitmentPoint.X, commitmentPoint.Y, beta)
	LeftPoint := new(EllipticPoint)
	LeftPoint.X, LeftPoint.Y = Curve.Add(xY, yY, alphaPoint.X, alphaPoint.Y)

	// Check whether right point is equal left point or not
	if (rightPoint.X.CmpAbs(LeftPoint.X) == 0) && (rightPoint.Y.CmpAbs(LeftPoint.Y) == 0) {
		return false
	}
	return true
}
