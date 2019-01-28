package zkp

import (
	"crypto/rand"
	"fmt"
	"github.com/ninjadotorg/constant/privacy"
	"math/big"
	"testing"
	"github.com/stretchr/testify/assert"
)


// randValue return random witness value for testing
func (wit *ComZeroWitness) randValue(testcase bool) {
	switch testcase {
	case false:
		commitmentValue := new(privacy.EllipticPoint)
		commitmentValue.Randomize()
		index := byte(3)
		commitmentRnd, _ := rand.Int(rand.Reader, privacy.Curve.Params().N)
		wit.Set(commitmentValue, &index, commitmentRnd)
	case true:
		index := byte(3)
		commitmentRnd, _ := rand.Int(rand.Reader, privacy.Curve.Params().N)
		commitmentValue := privacy.PedCom.CommitAtIndex(big.NewInt(0), commitmentRnd, index)
		wit.Set(commitmentValue, &index, commitmentRnd)
	}
}

//TestProofIsZero test prove and verify function
func TestPKComZero(t* testing.T){
	witness := new(ComZeroWitness)
	witness.randValue(true)

	proof, _ := witness.Prove()

	proofBytes := proof.Bytes()
	fmt.Printf("Zero commitment len proof: %v\n", len(proofBytes))

	//proof2 := new(ComZeroProof)
	//proof2.SetBytes(proofBytes)

	res := proof.Verify()

	assert.Equal(t, true, res)
}
