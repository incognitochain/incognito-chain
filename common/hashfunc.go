package common

import "crypto/sha256"

// HashB calculates hash(b) and returns the resulting bytes.
func HashB(b []byte) []byte {
	hash := sha256.Sum256(b)
	return hash[:]
}

// HashH calculates hash(b) and returns the resulting bytes as a Hash.
func HashH(b []byte) Hash {
	return Hash(sha256.Sum256(b))
}

// DoubleHashB calculates hash(hash(b)) and returns the resulting bytes.
func DoubleHashB(b []byte) []byte {
	first := sha256.Sum256(b)
	second := sha256.Sum256(first[:])
	return second[:]
}

func DoubleHashH(b []byte) Hash {
	first := sha256.Sum256(b)
	return Hash(sha256.Sum256(first[:]))
}
