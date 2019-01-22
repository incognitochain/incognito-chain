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

	wit.n = privacy.MaxExp

	start := time.Now()
	proof, err := wit.Prove()
	end:= time.Since(start)
	fmt.Printf("Single range proving time: %v\n", end)

	if err != nil{
		fmt.Printf("Err: %v\n", err)
	}
	start = time.Now()
	res := proof.Verify()
	end = time.Since(start)
	fmt.Printf("Single range verification time: %v\n", end)

	assert.Equal(t, true, res)
}
