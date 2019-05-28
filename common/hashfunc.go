package common

import (
	"golang.org/x/crypto/ripemd160"
	"golang.org/x/crypto/sha3"
	"io"
)

// HashB calculates hash(b) and returns the resulting bytes.
func HashB(b []byte) []byte {
	hash := sha3.Sum256(b)
	return hash[:]
}

// HashH calculates hash(b) and returns the resulting bytes as a Hash.
func HashH(b []byte) Hash {
	return Hash(sha3.Sum256(b))
}

//
// Hashes Rip 160
//
func HashRipeMD160(data []byte) ([]byte, error) {
	hasher := ripemd160.New()
	_, err := io.WriteString(hasher, string(data))
	if err != nil {
		return nil, err
	}
	return hasher.Sum(nil), nil
}

/*
Double hash or HASH 160
*/
func Hash160(data []byte) ([]byte, error) {
	hash1 := HashB(data)
	hash2, err := HashRipeMD160(hash1)
	if err != nil {
		return nil, err
	}

	return hash2, nil
}
