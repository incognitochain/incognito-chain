package zkp

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestOpeningsProtocol(t *testing.T) {
	witness := new(PKComOpeningsWitness)
	witness.randValue(true)

	proof, _ := witness.Prove()

	proof2 := new(PKComOpeningsProof)
	proof2.SetBytes(proof.Bytes())

	res := proof2.Verify()
	assert.Equal(t, true, res)
}
