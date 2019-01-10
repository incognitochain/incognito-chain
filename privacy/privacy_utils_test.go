package privacy

import (
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
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
	///********* Test multi exponentiation Algorithm *******/
	//values := []*big.Int{big.NewInt(5), big.NewInt(10),big.NewInt(5),big.NewInt(7), big.NewInt(5)}
	////fmt.Printf("Values: %v\n", values[0])
	//
	//expectedRes := PedCom.CommitAll(values)
	//
	////expectedRes := new(EllipticPoint).Zero()
	//////fmt.Printf("Values: %v\n", values[0])
	////for i:=0; i<len(values); i++{
	////	expectedRes = expectedRes.Add(PedCom.G[i].ScalarMult(values[i]))
	////}
	//fmt.Printf("Res from normal calculation: %+v\n", expectedRes)
	//
	//testcase4, err := multiExp(PedCom.G, values)
	//if err != nil{
	//	Logger.Log.Errorf("Error of multi-exponentiation algorithm")
	//}
	//fmt.Printf("Res from multi exponentiation alg: %+v\n", testcase4)
	//
	//
	////testcase5 := multiExp()
	//
	//
	//
	//
	//
	////fmt.Printf("5* G[0]: %+v\n",PedCom.G[0].ScalarMult(big.NewInt(5)) )
	//
	//assert.Equal(t, expectedRes, testcase4)


	testcase5 := exp(PedCom.G[0], big.NewInt(10))
	expectedRes := PedCom.G[0].ScalarMult(big.NewInt(10))
	//fmt.Printf("5* G[0]: %+v\n",PedCom.G[0].ScalarMult(big.NewInt(5)) )
	assert.Equal(t, expectedRes, testcase5)



}
