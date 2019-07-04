package common

import (
	"golang.org/x/crypto/sha3"
)

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