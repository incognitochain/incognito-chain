package privacy

import (
	"math/big"
)

// PedersenCommitment represents a commitment that includes 4 generators
type PedersenCommitment interface {
	// Params returns the parameters for the commitment
	Params() *PCParams
	// CommitAll commits
	CommitAll(openings []*big.Int) *EllipticPoint
	// CommitAtIndex commits value at index
	CommitAtIndex(value *big.Int, rand *big.Int, index byte) *EllipticPoint
	//CommitEllipticPoint commit a elliptic point
	CommitEllipticPoint(point *EllipticPoint, rand *big.Int) *EllipticPoint
}

// PCParams represents the parameters for the commitment
type PCParams struct {
	G        []*EllipticPoint // generators
	Capacity int
	// G[0]: public key
	// G[1]: Value
	// G[2]: SNDerivator
	// G[3]: ShardID
	// G[4]: Randomness
}

// newPedersenParams creates new generators
func newPedersenParams() PCParams {
	var pcm PCParams
	pcm.Capacity = 5
	pcm.G = make([]*EllipticPoint, pcm.Capacity)
	pcm.G[0] = new(EllipticPoint)
	pcm.G[0].X, pcm.G[0].Y = Curve.Params().Gx, Curve.Params().Gy

	for i := 1; i < pcm.Capacity; i++ {
		pcm.G[i] = pcm.G[0].Hash(i)
	}
	return pcm
}

var PedCom = newPedersenParams()

// Params returns parameters of commitment
func (com PCParams) Params() PCParams {
	return com
}

// CommitAll commits a list of PCM_CAPACITY value(s)
func (com PCParams) CommitAll(openings []*big.Int) *EllipticPoint {
	if len(openings) != com.Capacity {
		return nil
	}

	commitment := new(EllipticPoint).Zero()
	for i := 0; i < com.Capacity; i++ {
		commitment = commitment.Add(com.G[i].ScalarMult(openings[i]))
	}
	return commitment
}

// CommitAtIndex commits specific value with index and returns 34 bytes
func (com PCParams) CommitAtIndex(value, rand *big.Int, index byte) *EllipticPoint {
	commitment := com.G[com.Capacity-1].ScalarMult(rand).Add(com.G[index].ScalarMult(value))
	return commitment
}
