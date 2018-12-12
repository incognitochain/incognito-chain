package zkp

import (
	"fmt"
	"testing"
)

//TestProofIsZero test prove and verify function
func TestProofIsZero(t* testing.T){
	res := true
	for res {
		witness := new(PKComZeroWitness)
		witness.randValue(true)
		proof, _ := witness.Prove()
		fmt.Println(len(proof.Bytes()))
		res = proof.Verify()
	}
}
