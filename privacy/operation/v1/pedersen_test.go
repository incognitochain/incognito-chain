package operation

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestPedersenCommitAll(t *testing.T) {
	for i := 0; i < 100; i++ {
		openings := make([]*Scalar, len(PedCom.G))
		for i := 0; i < len(openings); i++ {
			openings[i] = RandomScalar()
		}

		commitment, err := PedCom.CommitAll(openings)
		isValid := commitment.PointValid()

		assert.NotEqual(t, commitment, nil)
		assert.Equal(t, true, isValid)
		assert.Equal(t, nil, err)
	}
}

func TestPedersenCommitAtIndex(t *testing.T) {
	for i := 0; i < 100; i++ {
		data := []struct {
			value *Scalar
			rand  *Scalar
			index byte
		}{
			{RandomScalar(), RandomScalar(), PedersenPrivateKeyIndex},
			{RandomScalar(), RandomScalar(), PedersenValueIndex},
			{RandomScalar(), RandomScalar(), PedersenSndIndex},
			{RandomScalar(), RandomScalar(), PedersenShardIDIndex},
		}

		for _, item := range data {
			commitment := PedCom.CommitAtIndex(item.value, item.rand, item.index)
			expectedCm := new(Point).ScalarMult(PedCom.G[item.index], item.value)
			expectedCm.Add(expectedCm, new(Point).ScalarMult(PedCom.G[PedersenRandomnessIndex], item.rand))
			assert.Equal(t, expectedCm, commitment)
		}
	}
}
