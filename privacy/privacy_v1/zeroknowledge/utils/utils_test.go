package utils

import (
	"fmt"
	"testing"
)

func TestEstimateProofSize(t *testing.T) {
	testcase1 := EstimateProofSize(4, 1, true)
	fmt.Printf("testcase 1: %v\n", testcase1)
}
