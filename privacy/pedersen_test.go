package privacy

import (
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestPedersenCommitAll(t *testing.T) {
	openings := make([]*big.Int, len(PedCom.G))
	for i := 0; i < len(openings); i++ {
		openings[i] = RandScalar()
	}

	commitment, err := PedCom.commitAll(openings)
	isOnCurve := Curve.IsOnCurve(commitment.X, commitment.Y)

	assert.NotEqual(t, commitment, nil)
	assert.Equal(t, true, isOnCurve)
	assert.Equal(t, nil, err)
}

func TestPedersenCommitAtIndex(t *testing.T) {
	data := []struct {
		value *big.Int
		rand  *big.Int
		index byte
	}{
		{RandScalar(), RandScalar(), PedersenPrivateKeyIndex},
		{RandScalar(), RandScalar(), PedersenValueIndex},
		{RandScalar(), RandScalar(), PedersenSndIndex},
		{RandScalar(), RandScalar(), PedersenShardIDIndex},
	}

	for _, item := range data {
		commitment := PedCom.CommitAtIndex(item.value, item.rand, item.index)
		expectedCm := PedCom.G[item.index].ScalarMult(item.value).Add(PedCom.G[PedersenRandomnessIndex].ScalarMult(item.rand))
		assert.Equal(t, expectedCm, commitment)
	}
}
