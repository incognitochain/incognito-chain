package privacy

import (
	"crypto/rand"
	"io"
	"math/big"
)

// RandBytes generates random bytes with length
func RandBytes(length int) []byte {
	rbytes := make([]byte, length)
	rand.Read(rbytes)
	return rbytes
}

// RandScalar generates a big int with value less than order of group of elliptic points
func RandScalar(r io.Reader) *big.Int {
	var k *big.Int
	var err error
	for {
		k, err = rand.Int(r, Curve.Params().N)
		if err != nil {
			continue
		}
		if k.Sign() > 0 {
			break
		}
	}
	return k
}

// ConvertIntToBinary represents a integer number in binary array with little endian with size n
func ConvertIntToBinary(inum int, n int) []byte {
	binary := make([]byte, n)

	for i := 0; i < n; i++ {
		binary[i] = byte(inum % 2)
		inum = inum / 2
	}

	return binary
}

// ConvertIntToBinary represents a integer number in binary
func ConvertBigIntToBinary(number *big.Int, n int) []*big.Int {
	if number.Cmp(big.NewInt(0)) == 0 {
		res := make([]*big.Int, n)
		for i := 0; i < n; i++ {
			res[i] = big.NewInt(0)
		}
		return res
	}

	binary := make([]*big.Int, n)
	numberClone := new(big.Int)
	numberClone.Set(number)

	zeroNumber := big.NewInt(0)
	twoNumber := big.NewInt(2)

	for i := 0; i < n; i++ {
		binary[i] = new(big.Int)
		binary[i] = new(big.Int).Mod(numberClone, twoNumber)
		numberClone.Div(numberClone, twoNumber)

		if numberClone.Cmp(zeroNumber) == 0 && i != n-1 {
			for j := i + 1; j < n; j++ {
				binary[j] = zeroNumber
			}
			break
		}
	}
	return binary
}

// isOdd check a big int is odd or not
func isOdd(a *big.Int) bool {
	return a.Bit(0) == 1
}

// padd1Div4 computes (p + 1) / 4
func padd1Div4(p *big.Int) (res *big.Int) {
	res = new(big.Int).Add(p, big.NewInt(1))
	res.Div(res, big.NewInt(4))
	return
}

// paddedAppend appends the src byte slice to dst, returning the new slice.
// If the length of the source is smaller than the passed size, leading zero
// bytes are appended to the dst slice before appending src.
func paddedAppend(size uint, dst, src []byte) []byte {
	for i := 0; i < int(size)-len(src); i++ {
		dst = append(dst, 0)
	}
	return append(dst, src...)
}
