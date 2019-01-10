package zkp

import (
	"fmt"
	"testing"
)

func TestEstimateProofSize(t *testing.T) {

	testcase1 := EstimateProofSize(256, 256, true)
	fmt.Printf("testcase 1: %v\n", testcase1)
	//assert.Equal(t, true, testcase1)

	//testcase2 := estimateMultiRangeProof(10)

	testcase2 := estimateMultiRangeProof(256)
	fmt.Printf("testcase 2: %v\n", testcase2)

}
