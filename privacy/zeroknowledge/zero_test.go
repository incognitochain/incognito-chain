package zkp

import (
	"crypto/rand"
	"github.com/ninjadotorg/constant/privacy"
	"math/big"
	"testing"
	"github.com/stretchr/testify/assert"
)


// randValue return random witness value for testing
func (wit *PKComZeroWitness) randValue(testcase bool) {
	switch testcase {
	case false:
		commitmentValue := new(privacy.EllipticPoint)
		commitmentValue.Randomize()
		index := byte(3)
		commitmentRnd, _ := rand.Int(rand.Reader, privacy.Curve.Params().N)
		wit.Set(commitmentValue, &index, commitmentRnd)
		break
	case true:
		index := byte(3)
		commitmentRnd, _ := rand.Int(rand.Reader, privacy.Curve.Params().N)
		commitmentValue := privacy.PedCom.CommitAtIndex(big.NewInt(0), commitmentRnd, index)
		wit.Set(commitmentValue, &index, commitmentRnd)
		break
	}
}

//TestProofIsZero test prove and verify function
func TestPKComZero(t* testing.T){
	witness := new(PKComZeroWitness)
	witness.randValue(true)

	proof, _ := witness.Prove()

	proofBytes := proof.Bytes()

	proof2 := new(PKComZeroProof)
	proof2.SetBytes(proofBytes)

	res := proof2.Verify()

	assert.Equal(t, true, res)
}
