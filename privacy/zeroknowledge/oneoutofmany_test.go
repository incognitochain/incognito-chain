package zkp

import (
	"fmt"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

//TestPKOneOfMany test protocol for one of many Commitment is Commitment to zero
func TestPKOneOfMany(t *testing.T) {
	witness := new(OneOutOfManyWitness)

	indexIsZero := 2

	// list of commitments
	commitments := make([]*privacy.EllipticPoint, privacy.CMRingSize)
	snDerivators := make([]*big.Int, privacy.CMRingSize)
	randoms := make([]*big.Int, privacy.CMRingSize)

	for i := 0; i < privacy.CMRingSize; i++ {
		snDerivators[i] = privacy.RandScalar()
		randoms[i] = privacy.RandScalar()
		commitments[i] = privacy.PedCom.CommitAtIndex(snDerivators[i], randoms[i], privacy.SND)
	}

	// create Commitment to zero at indexIsZero
	snDerivators[indexIsZero] = big.NewInt(0)
	commitments[indexIsZero] = privacy.PedCom.CommitAtIndex(snDerivators[indexIsZero], randoms[indexIsZero], privacy.SND)

	witness.Set(commitments, randoms[indexIsZero], uint64(indexIsZero))
	start := time.Now()
	proof, err := witness.Prove()
	end := time.Since(start)
	fmt.Printf("One out of many proving time: %v\n", end)
	if err != nil {
		privacy.Logger.Log.Error(err)
	}

	//Convert proof to bytes array
	proofBytes := proof.Bytes()

	fmt.Printf("One out of many proof size: %v\n", len(proofBytes))

	//revert bytes array to proof
	proof2 := new(OneOutOfManyProof).Init()
	proof2.SetBytes(proofBytes)
	proof2.stmt.commitments = commitments

	start = time.Now()

	res := proof2.Verify()
	end = time.Since(start)
	fmt.Printf("One out of many verification time: %v\n", end)

	assert.Equal(t, true, res)
}
