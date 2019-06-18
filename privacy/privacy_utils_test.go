package privacy

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

func TestIsPowerOfTwo(t *testing.T) {
	testcase1 := IsPowerOfTwo(64)
	assert.Equal(t, true, testcase1)

	testcase2 := IsPowerOfTwo(124)
	assert.Equal(t, false, testcase2)

	testcase3 := IsPowerOfTwo(0)
	assert.Equal(t, false, testcase3)

}

func TestConvertBigIntToBinary(t *testing.T) {
	num := big.NewInt(100)
	binary := ConvertBigIntToBinary(num, 10)

	for i:=0; i<len(binary); i++{
		fmt.Printf("%v: %v ",i,  binary[i])
	}
}

func TestMultiExponentiation(t *testing.T){
	exponents := []*big.Int{big.NewInt(5), big.NewInt(10),big.NewInt(5),big.NewInt(7), big.NewInt(5)}

	//exponents := make([]*big.Int, 64)
	//fmt.Printf("Values: %v\n", exponents[0])

	start1 := time.Now()
	expectedRes := PedCom.CommitAll(exponents)
	end1 := time.Since(start1)
	fmt.Printf("normal calculation time: %v\n", end1)
	fmt.Printf("Res from normal calculation: %+v\n", expectedRes)


	start2 := time.Now()
	testcase4, err := MultiScalarmult(PedCom.G, exponents)
	if err != nil{
		Logger.Log.Errorf("Error of multi-exponentiation algorithm")
	}
	end2 := time.Since(start2)
	fmt.Printf("multi scalarmult: %v\n", end2)
	fmt.Printf("Res from multi exponentiation alg: %+v\n", testcase4)

	start3 := time.Now()
	testcase5, err := MultiScalar2(PedCom.G, exponents)
	if err != nil{
		Logger.Log.Errorf("Error of multi-exponentiation algorithm")
	}
	end3 := time.Since(start3)
	fmt.Printf("multi scalarmult 2: %v\n", end3)
	fmt.Printf("Res from multi exponentiation 2 alg: %+v\n", testcase5)
	assert.Equal(t, expectedRes, testcase4)
}


func TestHashEC(t *testing.T){
	res := PedCom.G[0].Hash(100)
	fmt.Printf("Res: %v\n", res.Compress())

	res = PedCom.G[0].Hash(1000)
	fmt.Printf("Res: %v\n", res.Compress())

	res = PedCom.G[0].Hash(10000)
	fmt.Printf("Res: %v\n", res.Compress())

	res = PedCom.G[0].Hash(100000)
	fmt.Printf("Res: %v\n", res.Compress())

	res = PedCom.G[0].Hash(1000000)
	fmt.Printf("Res: %v\n", res.Compress())

	res = PedCom.G[0].Hash(10000000)
	fmt.Printf("Res: %v\n", res.Compress())

	res = PedCom.G[0].Hash(100000000)
	//fmt.Printf("Byte: %v\n", byte(100000000))
	fmt.Printf("Res: %v\n", res.Compress())

	res = PedCom.G[0].Hash(1000000000)
	fmt.Printf("Res: %v\n", res.Compress())

	res = PedCom.G[0].Hash(10000000000)
	fmt.Printf("Res: %v\n", res.Compress())
}

func TestMap(t *testing.T){
	m := map[int][]int{
		0: {0,2,4,6,8},
		1: {1,3,5,7},
	}

	m[2] = append(m[2], 10)

	fmt.Printf("m[false]: %v\n", m[0])
	fmt.Printf("m[true]: %v\n", m[1])
	fmt.Printf("m[2]: %v\n", m[2])

}
