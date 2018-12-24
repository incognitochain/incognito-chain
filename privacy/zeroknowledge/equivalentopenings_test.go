package zkp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPKEqualityOfCommittedVal(t *testing.T) {

	witness := new(PKEqualityOfCommittedValWitness)
	witness.randValue()

	proof := new(PKEqualityOfCommittedValProof)
	proof = witness.Prove()

	res := proof.Verify()
	assert.Equal(t, true, res)
}
