package zkp

import (
	"fmt"
	"testing"
)

func TestEstimateProofSize(t *testing.T) {

	testcase1 := EstimateProofSize(10, 10)
	fmt.Printf("testcase 1: %v\n", testcase1)
	//assert.Equal(t, true, testcase1)

	//fmt.Printf("PedCom: %+v\n", len(privacy.PedCom.G))
	//fmt.Printf("AggParam G : %+v\n", len(AggParam.G))
	//fmt.Printf("AggParam H : %+v\n", len(AggParam.H))


}
