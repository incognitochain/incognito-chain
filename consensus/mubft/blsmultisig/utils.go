package blsmultisig

import (
	"fmt"
	"math/big"
	"reflect"
	"runtime"
)

func printBit(bn *big.Int) {
	for i := 255; i >= 0; i-- {
		fmt.Printf("%+v", bn.Bit(i))
	}
	fmt.Println("")
}

// I2Bytes take an integer and return bytes arrays of it with fixed length
func I2Bytes(bn *big.Int, length int) []byte {
	res := bn.Bytes()
	for ; len(res) < length; res = append([]byte{0}, res...) {
	}
	return res
}

// func Bytes2I(bytes []byte) *big.Int{
// 	res := bn.Bytes()
// 	for ; len(res) < length; res = append([]byte{0}, res...) {
// 	}
// 	return res
// }

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
