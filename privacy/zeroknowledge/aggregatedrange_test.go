package zkp
//
//import (
//	"fmt"
//	"github.com/ninjadotorg/constant/privacy"
//	"github.com/stretchr/testify/assert"
//	"math/big"
//	"testing"
//)
//
//func TestPKComMultiRange(t *testing.T) {
//	numValues := 1
//	values := make([]*big.Int, numValues)
//	for i:=0; i<numValues; i++{
//		values[i] = new (big.Int).SetBytes(privacy.RandBytes(2))
//	}
//
//	witness := new (AggregatedRangeWitness)
//	witness.Set(values, privacy.MaxExp)
//
//	// Testing smallest number in range
//	proof, err := witness.Prove()
//	if err != nil{
//		fmt.Printf("error Aggregated range: %v\n", err)
//	}
//	proofBytes := proof.Bytes()
//	fmt.Printf("Aggregated range Proof size: %v\n", len(proofBytes))
//
//	proof2 := new(AggregatedRangeProof)
//	proof2.SetBytes(proofBytes)
//
//	res := proof2.Verify()
//
//	assert.Equal(t,true, res)
//}
