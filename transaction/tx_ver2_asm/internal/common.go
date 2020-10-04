package internal

import (
	"crypto/rand"
	"math/big"

	"golang.org/x/crypto/sha3"
)

const(
	HashSize = 32
	MaxShardNumber = 1
)
var(
	PRVCoinID = Hash{4}
)
type Hash [HashSize]byte


// HashB calculates SHA3-256 hashing of input b
// and returns the result in bytes array.
func HashB(b []byte) []byte {
	hash := sha3.Sum256(b)
	return hash[:]
}

// HashB calculates SHA3-256 hashing of input b
// and returns the result in Hash.
func HashH(b []byte) Hash {
	return Hash(sha3.Sum256(b))
}

func (hashObj *Hash) SetBytes(newHash []byte) error {
	nhlen := len(newHash)
	if nhlen != HashSize {
		return genericError
	}
	copy(hashObj[:], newHash)

	return nil
}

func RandBigIntMaxRange(max *big.Int) (*big.Int, error) {
	return rand.Int(rand.Reader, max)
}

func GetShardIDFromLastByte(b byte) byte {
	return byte(int(b) % MaxShardNumber)
}