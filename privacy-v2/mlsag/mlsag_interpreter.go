// INTERPRET/PARSE FROM SOME DEFINED VALUE INTO ANOTHER VALUE

package mlsag

import (
	"crypto/sha256"

	C25519 "github.com/incognitochain/incognito-chain/privacy/curve25519"
)

func hashToPoint(b []byte) *C25519.Key {
	keyHash := C25519.Key(C25519.Keccak256(b))
	return keyHash.HashToPoint()
}

func hashToNum(b []byte) [sha256.Size]byte {
	return sha256.Sum256(b)
}

// Parse public key from private key of C25519
func parsePublicKey(privateKey C25519.Key) [C25519.KeyLength]byte {
	publicKey := C25519.ScalarmultBase(&privateKey)
	return publicKey.ToBytes()
}

func parseKeyImages(privateKey []C25519.Key) (result []C25519.Key) {
	var m int = len(privateKey)

	result = make([]C25519.Key, m)
	for i := 0; i < m; i += 1 {
		publicKey := parsePublicKey(privateKey[i])
		hashPoint := hashToPoint(publicKey[:])
		image := C25519.ScalarMultKey(&privateKey[i], hashPoint)
		result[i] = *image
	}
	return
}
