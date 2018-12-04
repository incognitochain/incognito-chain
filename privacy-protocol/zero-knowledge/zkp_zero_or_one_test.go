package zkp

import (
	"fmt"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"math/big"
	"testing"
)

func TestPKComZeroOne(t *testing.T) {
	var res bool
	for i:=0;i<2;i++{
		// generate Openings
		valueRand := privacy.RandBytes(32)
		vInt := new(big.Int).SetBytes(valueRand)
		vInt.Mod(vInt, big.NewInt(2))
		rand := new(big.Int).SetBytes(privacy.RandBytes(32))

		// CommitAll
		cm := privacy.PedCom.CommitAtIndex(vInt, rand, privacy.VALUE)

		// create witness for proving
		var witness PKComZeroOneWitness
		witness.Set(vInt, rand, cm, privacy.VALUE)

		// Proving
		proof, _ := witness.Prove()
		fmt.Println(len(proof.Bytes()))
		//fmt.Printf("Proof: %+v\n", proof)

		// Set proof for verifying
		//Proof := new(PKComZeroOneProof)
		//Proof.Set(proof.ca, proof.cb, proof.f, proof.za, proof.zb, proof.commitment, proof.index)
		//
		//res = Proof.Verify()
		fmt.Println(res)
	}
}

