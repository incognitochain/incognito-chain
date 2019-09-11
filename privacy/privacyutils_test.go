package privacy

import (
	"crypto/rand"
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

func TestUtilsRandScalar(t *testing.T) {
	var r = rand.Reader
	for i := 0; i < 100; i++ {
		scalar := RandScalar(r)
		isLessThanN := scalar.Cmp(Curve.Params().N)
		assert.Equal(t, -1, isLessThanN)
		assert.GreaterOrEqual(t, common.BigIntSize, len(scalar.Bytes()))
	}
}

func TestUtilsRandScalar2(t *testing.T) {
	//var r io.Reader
	var r = rand.Reader
	for i := 0; i < 100; i++ {
		scalar := RandScalar(r)
		isLessThanN := scalar.Cmp(Curve.Params().N)
		fmt.Printf("Scalar: %v\n", scalar.Bytes())
		assert.Equal(t, -1, isLessThanN)
		assert.GreaterOrEqual(t, common.BigIntSize, len(scalar.Bytes()))
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
	}

	for _, item := range data {
		res := ConvertIntToBinary(item.number, item.size)
		assert.Equal(t, item.binary, res)
	}
}

func TestUtilsConvertBigIntToBinary(t *testing.T) {
	data := []struct {
		number *big.Int
		size   int
		binary []*big.Int
	}{
		{new(big.Int).SetUint64(uint64(64)), 8, []*big.Int{new(big.Int).SetInt64(0), new(big.Int).SetInt64(0), new(big.Int).SetInt64(0), new(big.Int).SetInt64(0), new(big.Int).SetInt64(0), new(big.Int).SetInt64(0), new(big.Int).SetInt64(1), new(big.Int).SetInt64(0)}},
		{new(big.Int).SetUint64(uint64(100)), 10, []*big.Int{new(big.Int).SetInt64(0), new(big.Int).SetInt64(0), new(big.Int).SetInt64(1), new(big.Int).SetInt64(0), new(big.Int).SetInt64(0), new(big.Int).SetInt64(1), new(big.Int).SetInt64(1), new(big.Int).SetInt64(0), new(big.Int).SetInt64(0), new(big.Int).SetInt64(0)}},
	}

	for _, item := range data {
		res := ConvertBigIntToBinary(item.number, item.size)
		assert.Equal(t, item.binary, res)
	}
}

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
