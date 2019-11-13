package privacy

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

var _ = func() (_ struct{}) {
	fmt.Println("This runs before init()!")
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

func TestUtilsRandBytes(t *testing.T) {
	data := []int{
		0,
		10,
		45,
		100,
		1000,
	}

	for _, item := range data {
		res := RandBytes(item)
		fmt.Printf("Res: %v\n", res)
		assert.Equal(t, item, len(res))
	}
}


func TestUtilsConvertIntToBinary(t *testing.T) {
	data := []struct {
		number int
		size   int
		binary []byte
	}{
		{64, 8, []byte{0, 0, 0, 0, 0, 0, 1, 0}},
		{100, 10, []byte{0, 0, 1, 0, 0, 1, 1, 0, 0, 0}},
		{1, 8, []byte{1, 0, 0, 0, 0, 0, 0, 0}},
	}

	for _, item := range data {
		res := ConvertIntToBinary(item.number, item.size)
		assert.Equal(t, item.binary, res)
	}
}

//func TestUtilsConvertBigIntToBinary(t *testing.T) {
//	data := []struct {
//		number *big.Int
//		size   int
//		binary []*big.Int
//	}{
//		{new(big.Int).FromUint64(uint64(64)), 8, []*big.Int{new(big.Int).SetInt64(0), new(big.Int).SetInt64(0), new(big.Int).SetInt64(0), new(big.Int).SetInt64(0), new(big.Int).SetInt64(0), new(big.Int).SetInt64(0), new(big.Int).SetInt64(1), new(big.Int).SetInt64(0)}},
//		{new(big.Int).FromUint64(uint64(100)), 10, []*big.Int{new(big.Int).SetInt64(0), new(big.Int).SetInt64(0), new(big.Int).SetInt64(1), new(big.Int).SetInt64(0), new(big.Int).SetInt64(0), new(big.Int).SetInt64(1), new(big.Int).SetInt64(1), new(big.Int).SetInt64(0), new(big.Int).SetInt64(0), new(big.Int).SetInt64(0)}},
//	}
//
//	for _, item := range data {
//		res := ConvertBigIntToBinary(item.number, item.size)
//		assert.Equal(t, item.binary, res)
//	}
//}

func TestUtilsAddPaddingBigInt(t *testing.T) {
	data := []struct {
		number *big.Int
		size   int
	}{
		{new(big.Int).SetBytes(RandBytes(12)), common.BigIntSize},
		{new(big.Int).SetBytes(RandBytes(42)), 50},
		{new(big.Int).SetBytes(RandBytes(0)), 10},
	}

	for _, item := range data {
		res := common.AddPaddingBigInt(item.number, item.size)
		assert.Equal(t, item.size, len(res))
	}
}

func TestUtilsIntToByteArr(t *testing.T) {
	data := []struct {
		number int
		bytes  []byte
	}{
		{12345, []byte{48, 57}},
		{123, []byte{0, 123}},
		{0, []byte{0, 0}},
	}

	for _, item := range data {
		res := common.IntToBytes(item.number)
		assert.Equal(t, item.bytes, res)

		number := common.BytesToInt(res)
		assert.Equal(t, item.number, number)
	}
}

func TestInterface(t *testing.T) {
	a:= make(map[string]interface{})
	a["x"] = "10"

	value, ok := a["y"].(string)
	if !ok {
		fmt.Printf("Param is invalid\n")
	}

	value2, ok := a["y"]
	if !ok {
		fmt.Printf("Param is invalid\n")
	}

	value3, ok := a["x"].(string)
	if !ok {
		fmt.Printf("Param is invalid\n")
	}

	fmt.Printf("Value: %v\n", value)
	fmt.Printf("Value2: %v\n", value2)
	fmt.Printf("Value2: %v\n", value3)
}




