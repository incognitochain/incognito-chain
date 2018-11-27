package zkp

import (
	"fmt"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"math/big"
	"testing"
)

func TestPKComMultiRangeValues(t *testing.T) {
	test:=8
	values := make([]*big.Int,test)
	for i:=0;i<test;i++{
		values[i] = new(big.Int)
		x:=new(big.Int).SetBytes(privacy.RandBytes(8))
		fmt.Println(x)
		values[i]=x
	}
	//values := []*big.Int{big.NewInt(5136325419070411678), big.NewInt()}
	var witness PKComMultiRangeWitness
	witness.Values = values
	var zk PKComMultiRangeProtocol
	zk.SetWitness(witness)
	RangeProofParams = NewECPrimeGroupKey(64 * len(values))
	// Testing smallest number in range
	proof:=zk.Prove()
	proofStr:=fmt.Sprintf("%s", proof)
	fmt.Printf("Proof Size: %d KB\n",len(proofStr)*1.0/1024)
	if MRPVerify(proof) {
		fmt.Println("Multi Range Proof Verification works")
	} else {
		t.Errorf("***** Multi Range Proof FAILURE")
	}
}
