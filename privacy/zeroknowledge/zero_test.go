package zkp

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

//TestProofIsZero test prove and verify function
func TestPKComZero(t* testing.T){
	witness := new(PKComZeroWitness)
	witness.randValue(true)

	proof, _ := witness.Prove()
	res := proof.Verify()

	assert.Equal(t, true, res)
}
