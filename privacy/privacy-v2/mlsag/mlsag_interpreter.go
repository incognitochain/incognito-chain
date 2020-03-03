// INTERPRET/PARSE FROM SOME DEFINED VALUE INTO ANOTHER VALUE

package mlsag

import (
	"crypto/sha256"

	"github.com/incognitochain/incognito-chain/privacy/operation"
)

func hashToNum(b []byte) [sha256.Size]byte {
	return sha256.Sum256(b)
}

// Parse public key from private key
func parsePublicKey(privateKey operation.Scalar) *operation.Point {
	return new(operation.Point).ScalarMultBase(&privateKey)
}

func parseKeyImages(privateKey *[]operation.Scalar) *[]operation.Point {
	priv := *privateKey
	var m int = len(priv)

	result := make([]operation.Point, m)
	for i := 0; i < m; i += 1 {
		publicKey := parsePublicKey(priv[i])
		hashPoint := operation.HashToPoint(publicKey.ToBytesS())
		result[i] = *new(operation.Point).ScalarMult(hashPoint, &priv[i])
	}
	return &result
}
