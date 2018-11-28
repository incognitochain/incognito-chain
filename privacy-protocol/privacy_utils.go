package privacy

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// RandBytes generates random bytes
func RandBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("error:", err)
		return nil
	}
	return b
}

// RandInt generates a big int with value less than order of group of elliptic points
func RandInt() *big.Int {
	res, _ := rand.Int(rand.Reader, Curve.Params().N)
	return res
}

// IsPowerOfTwo checks whether n is power of two or not
func IsPowerOfTwo(n int) bool {
	if n < 2 {
		return false
	}
	for n > 2 {
		if n%2 == 0 {
			n = n / 2
		} else {
			return false
		}
	}
	return true
}

// ConvertIntToBinary represents a integer number in binary
func ConvertIntToBinary(inum int, n int) []byte {
	binary := make([]byte, n)

	for i := n - 1; i >= 0; i-- {
		binary[i] = byte(inum % 2)
		inum = inum / 2
	}

	return binary
}

//

// Cmp compare two hash
// hash = target : return 0
// hash > target : return 1
// hash < target : return -1
func (hash *Hash) Cmp(target *Hash) int {
	for i := 0; i < HashSize; i++ {
		if hash[i] > target[i] {
			return 1
		}
		if hash[i] < target[i] {
			return -1
		}
	}
	return 0
}
