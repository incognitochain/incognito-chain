package zkp

import (
	"github.com/stretchr/testify/assert"
	"github.com/ninjadotorg/constant/privacy"
	"math/big"
	"testing"
)

func TestPKComMultiRange(t *testing.T) {
	numValues := 2
	values := make([]*big.Int, numValues)

	for i := 0; i < numValues; i++ {
		values[i] = new(big.Int).SetBytes(privacy.RandBytes(1))
	}

	var witness AggregatedRangeWitness
	witness.Set(values, 64)

	// Testing smallest number in range
	proof, _ := witness.Prove()
	b := proof.Bytes()

	Vproof := new(AggregatedRangeProof)
	Vproof.SetBytes(b)

	res := Vproof.Verify()

	assert.Equal(t,true, res)
}
