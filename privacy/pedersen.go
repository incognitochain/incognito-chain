package privacy

import (
	"github.com/pkg/errors"
	"math/big"
)

const (
	PedersenPrivateKeyIndex = byte(0x00)
	PedersenValueIndex      = byte(0x01)
	PedersenSndIndex        = byte(0x02)
	PedersenShardIDIndex    = byte(0x03)
	PedersenRandomnessIndex = byte(0x04)
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

func newPedersenParams() PedersenCommitment {
	var pcm PedersenCommitment
	const capacity = 5 // fixed Value = 5
	pcm.G = make([]*EllipticPoint, capacity)
	pcm.G[0] = new(EllipticPoint)
	pcm.G[0].Set(Curve.Params().Gx, Curve.Params().Gy)

	for i := 1; i < len(pcm.G); i++ {
		pcm.G[i] = pcm.G[0].Hash(int64(i))
	}
	return pcm
}

var PedCom = newPedersenParams()

// CommitAll commits a list of PCM_CAPACITY Value(s)
func (com PedersenCommitment) commitAll(openings []*big.Int) (*EllipticPoint, error) {
	if len(openings) != len(com.G) {
		return nil, errors.New("invalid length of openings to commit")
	}

	commitment := new(EllipticPoint)
	commitment.Zero()
	for i := 0; i < len(com.G); i++ {
		commitment = commitment.Add(com.G[i].ScalarMult(openings[i]))
	}
	return commitment, nil
}

// CommitAtIndex commits specific Value with index and returns 34 bytes
func (com PedersenCommitment) CommitAtIndex(value, rand *big.Int, index byte) *EllipticPoint {
	commitment := com.G[PedersenRandomnessIndex].ScalarMult(rand).Add(com.G[index].ScalarMult(value))
	return commitment
}
