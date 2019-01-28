package zkp

import (
	"fmt"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

func TestSingleRangeProve(t *testing.T){
	wit := new(SingleRangeWitness)
	wit.value = big.NewInt(11)
	wit.rand = privacy.RandInt()

	start := time.Now()
	proof, err := wit.Prove()
	if err != nil{
		fmt.Printf("Err: %v\n", err)
	}
	end:= time.Since(start)
	fmt.Printf("Single range proving time: %v\n", end)

	//fmt.Printf("Proof size: \n")
	//fmt.Printf("cmValue: %v\n", len(proof.cmValue.Compress()))
	//fmt.Printf("A: %v\n", len(proof.A.Compress()))
	//fmt.Printf("S: %v\n", len(proof.S.Compress()))
	//fmt.Printf("T1: %v\n", len(proof.T1.Compress()))
	//fmt.Printf("T2: %v\n", len(proof.T2.Compress()))
	//fmt.Printf("TauX: %v\n", len(proof.tauX.Bytes()))
	//fmt.Printf("tHat: %v\n", len(proof.tHat.Bytes()))
	//fmt.Printf("mu: %v\n", len(proof.mu.Bytes()))
	//
	//fmt.Printf("IPA.len L: %v\n", len(proof.innerProductProof.L))
	//fmt.Printf("IPA.len R: %v\n", len(proof.innerProductProof.R))
	//
	//fmt.Printf("IPA.a: %v\n", len(proof.innerProductProof.a.Bytes()))
	//fmt.Printf("IPA.b: %v\n", len(proof.innerProductProof.b.Bytes()))
	//fmt.Printf("IPA.p: %v\n", len(proof.innerProductProof.p.Compress()))
	//fmt.Printf("n: %v\n", 1)


	start = time.Now()
	res := proof.Verify()
	end = time.Since(start)
	fmt.Printf("Single range verification time: %v\n", end)

	assert.Equal(t, true, res)
}

func BenchmarkSingleRangeProve(b *testing.B){
	for i:=0; i<b.N; i++ {
		wit := new(SingleRangeWitness)
		wit.value = big.NewInt(11)
		wit.rand = privacy.RandInt()

		//for n := 0; n < 1000; n++ {
		//	start := time.Now()
			proof, err := wit.Prove()
			//end := time.Since(start)
			//fmt.Printf("Single range proving time: %v\n", end)

			if err != nil {
				fmt.Printf("Err: %v\n", err)
			}
			//start = time.Now()
			res := proof.Verify()
			//end = time.Since(start)
			//fmt.Printf("Single range verification time: %v\n", end)

			assert.Equal(b, true, res)
		//}

	}
}
