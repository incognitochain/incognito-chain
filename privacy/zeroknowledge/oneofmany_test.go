package zkp

import (
	"github.com/ninjadotorg/constant/privacy"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

//TestPKOneOfMany test protocol for one of many Commitment is Commitment to zero
func TestPKOneOfMany(t *testing.T) {
	witness := new(PKOneOfManyWitness)

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

	witness.Set(commitments, nil, randoms[indexIsZero], uint64(indexIsZero), privacy.SND)

	proof, err := witness.Prove()
	if err != nil {
		privacy.Logger.Log.Error(err)
	}

	//Convert proof to bytes array
	proofBytes := proof.Bytes()

	//revert bytes array to proof
	proof2 := new(PKOneOfManyProof)
	proof2.SetBytes(proofBytes)

	res := proof.Verify()

	assert.Equal(t, true, res)
}
