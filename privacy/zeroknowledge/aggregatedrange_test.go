package zkp

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestPKComMultiRange(t *testing.T) {
	values := []*big.Int{big.NewInt(10), big.NewInt(20)}

	var witness AggregatedRangeWitness
	witness.Set(values, 64)

	// Testing smallest number in range
	proof, _ := witness.Prove()
	b := proof.Bytes()
	fmt.Printf("Proof size: %v\n", len(b))

	Vproof := new(AggregatedRangeProof)
	Vproof.SetBytes(b)

	res := Vproof.Verify()

	assert.Equal(t,true, res)
}
