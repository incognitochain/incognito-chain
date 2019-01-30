package zkp

import (
	"fmt"
	"testing"
)

func TestEstimateProofSize(t *testing.T) {
	testcase1 := EstimateProofSize(4, 2, false)
	fmt.Printf("testcase 1: %v\n", testcase1)
}
