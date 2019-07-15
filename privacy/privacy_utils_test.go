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
		assert.Equal(t, item, len(res))
	}
}

func TestUtilsRandScalar(t *testing.T) {
	for i:= 0;i<100; i++{
		scalar := RandScalar()
		isLessThanN := scalar.Cmp(Curve.Params().N)
		assert.Equal(t, -1, isLessThanN)
		assert.Equal(t, BigIntSize, len(scalar.Bytes()))
	}
}

func TestUtilsIsPowerOfTwo(t *testing.T) {
	data := []struct{
		number int
		isPowerOf2 bool
	}{
		{64, true},
		{124, false},
		{0, false},
		{1, false},
	}

	for _, item := range data {
		res := IsPowerOfTwo(item.number)
		assert.Equal(t, item.isPowerOf2, res)
	}
}

func TestUtilsConvertIntToBinary(t *testing.T) {
	data := []struct{
		number int
		size int
		binary []byte
	}{
		{64, 8, []byte{0,0,0,0,0,0,1,0}},
		{100, 10, []byte{0,0,1,0,0,1,1,0,0,0}},
	}

	for _, item := range data {
		res := ConvertIntToBinary(item.number, item.size)
		assert.Equal(t, item.binary, res)
	}
}

func TestUtilsConvertBigIntToBinary(t *testing.T) {
	data := []struct{
		number *big.Int
		size int
		binary []*big.Int
	}{
		{new(big.Int).SetUint64(uint64(64)), 8, []*big.Int{new(big.Int).SetInt64(0),new(big.Int).SetInt64(0),new(big.Int).SetInt64(0),new(big.Int).SetInt64(0),new(big.Int).SetInt64(0),new(big.Int).SetInt64(0),new(big.Int).SetInt64(1),new(big.Int).SetInt64(0)}},
		{new(big.Int).SetUint64(uint64(100)), 10, []*big.Int{new(big.Int).SetInt64(0),new(big.Int).SetInt64(0),new(big.Int).SetInt64(1),new(big.Int).SetInt64(0),new(big.Int).SetInt64(0),new(big.Int).SetInt64(1),new(big.Int).SetInt64(1),new(big.Int).SetInt64(0),new(big.Int).SetInt64(0),new(big.Int).SetInt64(0)}},
	}

	for _, item := range data {
		res := ConvertBigIntToBinary(item.number, item.size)
		assert.Equal(t, item.binary, res)
	}
}

func TestUtilsAddPaddingBigInt(t *testing.T) {
	data := []struct{
		number *big.Int
		size int
	}{
		{new(big.Int).SetBytes(RandBytes(12)), BigIntSize},
		{new(big.Int).SetBytes(RandBytes(42)), 50},
		{new(big.Int).SetBytes(RandBytes(0)), 10},
	}

	for _, item := range data {
		res := AddPaddingBigInt(item.number, item.size)
		assert.Equal(t, item.size, len(res))
	}
}

func TestUtilsIntToByteArr(t *testing.T) {
	data := []struct{
		number int
		bytes []byte
	}{
		{12345, []byte{48, 57}},
		{123, []byte{0, 123}},
		{0, []byte{0,0}},
	}

	for _, item := range data {
		res := IntToByteArr(item.number)
		assert.Equal(t, item.bytes, res)

		number := ByteArrToInt(res)
		assert.Equal(t, item.number, number)
	}
}