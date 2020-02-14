// INTERPRET/PARSE FROM SOME DEFINED VALUE INTO ANOTHER VALUE

package mlsag

import (
	"crypto/sha256"

	"github.com/incognitochain/incognito-chain/privacy"
)

func hashToNum(b []byte) [sha256.Size]byte {
	return sha256.Sum256(b)
}

// Parse public key from private key
func parsePublicKey(privateKey privacy.Scalar) *privacy.Point {
	return new(privacy.Point).ScalarMultBase(&privateKey)
}

func parseKeyImages(privateKey []privacy.Scalar) (result []privacy.Point) {
	var m int = len(privateKey)

	result = make([]privacy.Point, m)
	for i := 0; i < m; i += 1 {
		publicKey := parsePublicKey(privateKey[i])
		hashPoint := privacy.HashToPoint(publicKey.ToBytesS())
		result[i] = *new(privacy.Point).ScalarMult(hashPoint, &privateKey[i])
	}
	return
}
