package pedersen

import (
	"testing"

	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/stretchr/testify/assert"
)

func TestPedersenCommitAll(t *testing.T) {
	for i := 0; i < 100; i++ {
		openings := make([]*Scalar, len(pedCom.G))
		for i := 0; i < len(openings); i++ {
			openings[i] = operation.RandomScalar()
		}

		commitment, err := pedCom.CommitAll(openings)
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
			{operation.RandomScalar(), operation.RandomScalar(), PedersenPrivateKeyIndex},
			{operation.RandomScalar(), operation.RandomScalar(), PedersenValueIndex},
			{operation.RandomScalar(), operation.RandomScalar(), PedersenSndIndex},
			{operation.RandomScalar(), operation.RandomScalar(), PedersenShardIDIndex},
		}

		for _, item := range data {
			commitment := pedCom.CommitAtIndex(item.value, item.rand, item.index)
			expectedCm := new(Point).ScalarMult(pedCom.G[item.index], item.value)
			expectedCm.Add(expectedCm, new(Point).ScalarMult(pedCom.G[PedersenRandomnessIndex], item.rand))
			assert.Equal(t, expectedCm, commitment)
		}
	}
}
