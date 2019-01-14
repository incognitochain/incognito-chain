package privacy

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

func TestIsPowerOfTwo(t *testing.T) {

	//testcase1 := IsPowerOfTwo(64)
	//assert.Equal(t, true, testcase1)
	//
	//testcase2 := IsPowerOfTwo(124)
	//assert.Equal(t, false, testcase2)
	//
	//testcase3 := IsPowerOfTwo(0)
	//assert.Equal(t, false, testcase3)
	//
	//a := new(big.Int).SetBytes([]byte{})
	//fmt.Printf("a %v\n", a)
	//
	/********* Test multi exponentiation Algorithm *******/
	exponents := []*big.Int{big.NewInt(5), big.NewInt(10),big.NewInt(5),big.NewInt(7), big.NewInt(5)}
	//fmt.Printf("Values: %v\n", exponents[0])

	start1 := time.Now()
	expectedRes := PedCom.CommitAll(exponents)
	end1 := time.Since(start1)
	fmt.Printf("normal calculation time: %v\n", end1)
	//fmt.Printf("Res from normal calculation: %+v\n", expectedRes)


	start2 := time.Now()
	testcase4, err := multiScalarmult(PedCom.G, exponents)
	end2 := time.Since(start2)
	fmt.Printf("normal calculation time: %v\n", end2)


	if err != nil{
		Logger.Log.Errorf("Error of multi-exponentiation algorithm")
	}
	fmt.Printf("Res from multi exponentiation alg: %+v\n", testcase4)

	assert.Equal(t, expectedRes, testcase4)




}
