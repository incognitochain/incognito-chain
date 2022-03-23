//nolint:revive // skip linter for this package name
package privacy_util

import (
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/operation"
)

func ScalarToBigInt(sc *operation.Scalar) *big.Int {
	rev := operation.Reverse(sc.ToBytes())
	bn := big.NewInt(0).SetBytes(rev[:])
	return bn
}

func BigIntToScalar(bn *big.Int) *operation.Scalar {
	bSlice := common.AddPaddingBigInt(bn, operation.Ed25519KeySize)
	var b [32]byte
	copy(b[:], bSlice)
	rev := operation.Reverse(b)
	sc := operation.NewScalar()
	sc.FromBytes(rev)
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

func SliceToArray(slice []byte) [operation.Ed25519KeySize]byte {
	var array [operation.Ed25519KeySize]byte
	copy(array[:], slice)
	return array
}

func ArrayToSlice(array [operation.Ed25519KeySize]byte) []byte {
	var slice []byte = array[:]
	return slice
}
