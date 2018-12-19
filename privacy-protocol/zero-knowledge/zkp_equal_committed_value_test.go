package zkp

import (
	"fmt"
	"testing"
)

func TestPKEqualityOfCommittedVal(t *testing.T) {
	res := true
	for res {
		witness := new(PKEqualityOfCommittedValWitness)
		witness.randValue()
		proof := new(PKEqualityOfCommittedValProof)
		proof = witness.Prove()
		fmt.Println(len(proof.Bytes()))
		res = proof.Verify()
	}
}
