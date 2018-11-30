package temp

import (
	"fmt"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"math/big"
	"testing"
)

func TestPKComMultiRange(t *testing.T) {
	testNum:= 14
	values := make([]*big.Int, testNum)
	for i:=0;i< testNum;i++{
		values[i] = new(big.Int)
		x:=new(big.Int).SetBytes(privacy.RandBytes(4))
		values[i] = x
	}
	var witness PKComMultiRangeWitness
	witness.Set(values,64)

	// Testing smallest number in range
	proof,_:= witness.Prove()
	if proof.Verify() {
		fmt.Println("Multi Range Proof Verification works")
	} else {
		fmt.Println("***** Multi Range Proof FAILURE")
	}
	proof_sum,_:= witness.ProveSum()
	if proof.VerifySum(proof_sum) {
		fmt.Println("Sum Proof Verification works")
	} else {
		fmt.Println("***** Sum Proof FAILURE")
	}
}

