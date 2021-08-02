//nolint:revive // skip linter for this package name
package privacy_util

import (
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/operation/curve25519"
)

func ScalarToBigInt(sc *operation.Scalar) *big.Int {
	keyR := operation.Reverse(sc.GetKey())
	keyRByte := keyR.ToBytes()
	bi := new(big.Int).SetBytes(keyRByte[:])
	return bi
}

func BigIntToScalar(bi *big.Int) *operation.Scalar {
	biByte := common.AddPaddingBigInt(bi, operation.Ed25519KeySize)
	var key curve25519.Key
	key.FromBytes(SliceToArray(biByte))
	keyR := operation.Reverse(key)
	sc, err := new(operation.Scalar).SetKey(&keyR)
	if err != nil {
		return nil
	}
	return sc
}

// ConvertIntToBinary represents a integer number in binary array with little endian with size n
func ConvertIntToBinary(inum int, n int) []byte {
	binary := make([]byte, n)

	for i := 0; i < n; i++ {
		binary[i] = byte(inum % 2)
		inum /= 2
	}

	return binary
}

// ConvertIntToBinary represents a integer number in binary
func ConvertUint64ToBinary(number uint64, n int) []*operation.Scalar {
	if number == 0 {
		res := make([]*operation.Scalar, n)
		for i := 0; i < n; i++ {
			res[i] = new(operation.Scalar).FromUint64(0)
		}
		return res
	}

	binary := make([]*operation.Scalar, n)

	for i := 0; i < n; i++ {
		binary[i] = new(operation.Scalar).FromUint64(number % 2)
		number /= 2
	}
	return binary
}

func ConvertScalarArrayToBigIntArray(scalarArr []*operation.Scalar) []*big.Int {
	res := make([]*big.Int, len(scalarArr))

	for i := 0; i < len(res); i++ {
		tmp := operation.Reverse(scalarArr[i].GetKey())
		res[i] = new(big.Int).SetBytes(ArrayToSlice(tmp.ToBytes()))
	}

	return res
}

func SliceToArray(slice []byte) [operation.Ed25519KeySize]byte {
	var array [operation.Ed25519KeySize]byte
	copy(array[:], slice)
	return array
}

func ArrayToSlice(array [operation.Ed25519KeySize]byte) []byte {
	var slice []byte = array[:]
	return slice
}
