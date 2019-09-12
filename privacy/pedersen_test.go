package privacy

import (
	"crypto/rand"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestPedersenCommitAll(t *testing.T) {
	var r = rand.Reader
	openings := make([]*big.Int, len(PedCom.G))
	for i := 0; i < len(openings); i++ {
		openings[i] = RandScalar(r)
	}

	commitment, err := PedCom.commitAll(openings)
	isOnCurve := Curve.IsOnCurve(commitment.x, commitment.y)

	assert.NotEqual(t, commitment, nil)
	assert.Equal(t, true, isOnCurve)
	assert.Equal(t, nil, err)
}

func TestPedersenCommitAtIndex(t *testing.T) {
	var r = rand.Reader
	data := []struct {
		value *big.Int
		rand  *big.Int
		index byte
	}{
		{RandScalar(r), RandScalar(r), PedersenPrivateKeyIndex},
		{RandScalar(r), RandScalar(r), PedersenValueIndex},
		{RandScalar(r), RandScalar(r), PedersenSndIndex},
		{RandScalar(r), RandScalar(r), PedersenShardIDIndex},
	}

	for _, item := range data {
		commitment := PedCom.CommitAtIndex(item.value, item.rand, item.index)
		expectedCm := PedCom.G[item.index].ScalarMult(item.value).Add(PedCom.G[PedersenRandomnessIndex].ScalarMult(item.rand))
		assert.Equal(t, expectedCm, commitment)
	}
}
