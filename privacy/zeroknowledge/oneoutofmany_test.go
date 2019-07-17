package zkp

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"math/big"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	m.Run()
}

var _ = func() (_ struct{}) {
	fmt.Println("This runs before init()!")
	privacy.Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

//TestPKOneOfMany test protocol for one of many Commitment is Commitment to zero
func TestPKOneOfMany(t *testing.T) {
	// prepare witness for Out out of many protocol
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
	assert.Equal(t, nil, err)
	end := time.Since(start)
	fmt.Printf("One out of many proving time: %v\n", end)

	// validate sanity for proof
	isValidSanity := proof.ValidateSanity()
	assert.Equal(t, true, isValidSanity)

	//Convert proof to bytes array
	proofBytes := proof.Bytes()
	assert.Equal(t, privacy.OneOfManyProofSize, len(proofBytes))

	// revert bytes array to proof
	proof2 := new(OneOutOfManyProof).Init()
	proof2.SetBytes(proofBytes)
	proof2.stmt.commitments = commitments
	assert.Equal(t, proof, proof2)

	// verify the proof
	start = time.Now()
	res := proof.Verify()
	end = time.Since(start)
	fmt.Printf("One out of many verification time: %v\n", end)
	assert.Equal(t, true, res)
}

func TestGetCoefficient(t *testing.T) {
	a := make([]*big.Int, 3)

	a[0] = new(big.Int).SetBytes([]byte{28, 30, 162, 177, 161, 127, 119, 10, 195, 106, 31, 125, 252, 56, 111, 229, 236, 245, 202, 172, 27, 54, 110, 9, 9, 8, 56, 189, 248, 100, 190, 129})
	a[1] = new(big.Int).SetBytes([]byte{144, 245, 78, 232, 93, 155, 71, 49, 175, 154, 78, 81, 146, 120, 171, 74, 88, 99, 196, 61, 124, 156, 35, 55, 39, 22, 189, 111, 108, 236, 3, 131})
	a[2] = new(big.Int).SetBytes([]byte{224, 15, 114, 83, 56, 148, 202, 7, 187, 99, 242, 4, 2, 168, 169, 168, 44, 174, 215, 111, 119, 162, 172, 44, 225, 97, 236, 240, 242, 233, 148, 49})

	res := GetCoefficient([]byte{0, 1, 1}, 3, 3, a, []byte{0, 1, 1})
	fmt.Printf("res: %v\n", res.Bytes())
}