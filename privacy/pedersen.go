package privacy

import (
	"math/big"
)

// PedersenCommitment represents the parameters for the commitment
type PedersenCommitment struct {
	G []*EllipticPoint // generators
	// G[0]: public key
	// G[1]: Value
	// G[2]: SNDerivator
	// G[3]: ShardID
	// G[4]: Randomness
}

// newPedersenCommitment creates new generators
func newPedersenCommitment() PedersenCommitment {
	var pcm PedersenCommitment
	capacity := 5 // fixed value
	pcm.G = make([]*EllipticPoint, capacity, capacity)
	pcm.G[0] = new(EllipticPoint)
	pcm.G[0].X, pcm.G[0].Y = Curve.Params().Gx, Curve.Params().Gy

	for i := 1; i < len(pcm.G); i++ {
		pcm.G[i] = pcm.G[0].Hash(i)
	}
	return pcm
}

var PedCom = newPedersenCommitment()

// Params returns parameters of commitment
func (com PedersenCommitment) Params() PedersenCommitment {
	return com
}

// CommitAll commits a list of PCM_CAPACITY value(s)
func (com PedersenCommitment) CommitAll(openings []*big.Int) *EllipticPoint {
	if len(openings) != len(com.G) {
		return nil
	}

	commitment := new(EllipticPoint).Zero()
	for i := 0; i < len(com.G); i++ {
		commitment = commitment.Add(com.G[i].ScalarMult(openings[i]))
	}
	return commitment
}

// CommitAtIndex commits specific value with index and returns 34 bytes
func (com PedersenCommitment) CommitAtIndex(value, rand *big.Int, index byte) *EllipticPoint {
	commitment := com.G[len(com.G)-1].ScalarMult(rand).Add(com.G[index].ScalarMult(value))
	return commitment
}
