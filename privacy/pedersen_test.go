package privacy

import (
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestPedersenCommitAll(t *testing.T){
	openings := make([]*big.Int, len(PedCom.G))
	for i :=0; i<len(openings); i++ {
		openings[i] = RandScalar()
	}

	commitment := PedCom.CommitAll(openings)
	isOnCurve := Curve.IsOnCurve(commitment.X, commitment.Y)

	assert.NotEqual(t, commitment, nil)
	assert.Equal(t, true, isOnCurve)
}

func TestPedersenCommitAtIndex(t *testing.T){
	data := []struct{
		value *big.Int
		rand *big.Int
		index byte
	}{
		{RandScalar(), RandScalar(), SK},
		{RandScalar(), RandScalar(), VALUE},
		{RandScalar(), RandScalar(), SND},
		{RandScalar(), RandScalar(), SHARDID},
	}

	for _, item := range data {
		commitment := PedCom.CommitAtIndex(item.value, item.rand, item.index)
		expectedCm := PedCom.G[item.index].ScalarMult(item.value).Add(PedCom.G[RAND].ScalarMult(item.rand))
		assert.Equal(t, expectedCm, commitment)
	}
}