package zkp

import (
	"fmt"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

//TestPKOneOfMany test protocol for one of many Commitment is Commitment to zero
func TestPKOneOfMany(t *testing.T) {
	witness := new(OneOutOfManyWitness)

	indexIsZero := 2

	// list of commitments
	commitments := make([]*privacy.EllipticPoint, privacy.CMRingSize)
	SNDerivators := make([]*big.Int, privacy.CMRingSize)
	randoms := make([]*big.Int, privacy.CMRingSize)

	for i := 0; i < privacy.CMRingSize; i++ {
		SNDerivators[i] = privacy.RandInt()
		randoms[i] = privacy.RandInt()
		commitments[i] = privacy.PedCom.CommitAtIndex(SNDerivators[i], randoms[i], privacy.SND)
	}

	// create Commitment to zero at indexIsZero
	SNDerivators[indexIsZero] = big.NewInt(0)
	commitments[indexIsZero] = privacy.PedCom.CommitAtIndex(SNDerivators[indexIsZero], randoms[indexIsZero], privacy.SND)

	witness.Set(commitments, []uint64{1,4,5,8,9,10,23,45}, randoms[indexIsZero], uint64(indexIsZero), privacy.SND)

	proof, err := witness.Prove()
	if err != nil {
		privacy.Logger.Log.Error(err)
	}

	//Convert proof to bytes array
	proofBytes := proof.Bytes()

	fmt.Printf("One out of many proof size: %v\n", len(proofBytes))

	//revert bytes array to proof
	proof2 := new(OneOutOfManyProof)
	proof2.SetBytes(proofBytes)
	proof2.Commitments = commitments

	res := proof2.Verify()

	assert.Equal(t, true, res)
}
