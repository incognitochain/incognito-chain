package privacy

import (
	"math/big"
)

const (
	SK      = byte(0x00)
	VALUE   = byte(0x01)
	SND     = byte(0x02)
	SHARDID = byte(0x03)
	RAND    = byte(0x04)
	FULL    = byte(0x05)
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

func setPedersenParams() PedersenCommitment {
	var pcm PedersenCommitment
	const capacity = 5 // fixed value
	pcm.G = make([]*EllipticPoint, capacity, capacity)

	for i := 0; i < capacity; i++ {
		pcm.G[i] = new(EllipticPoint)
		pcm.G[i].Set(PubParams.G[i].X, PubParams.G[i].Y)
	}
	return pcm
}

var PedCom = setPedersenParams()

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
