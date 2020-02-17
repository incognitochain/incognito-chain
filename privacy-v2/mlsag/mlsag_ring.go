package mlsag

import (
	"github.com/incognitochain/incognito-chain/privacy"
)

type Ring struct {
	keys [][]privacy.Point
}

func createFakePublicKeyArray(length int) (K []privacy.Point) {
	K = make([]privacy.Point, length)
	for i := 0; i < length; i += 1 {
		K[i] = *privacy.RandomPoint()
	}
	return
}

// Create a random ring with dimension: (numFake; len(privateKeys)) where we generate fake public keys inside
func NewRandomRing(privateKeys []privacy.Scalar, numFake, pi int) (K *Ring) {
	m := len(privateKeys)
	if K == nil {
		K = new(Ring)
	}
	K.keys = make([][]privacy.Point, numFake)
	for i := 0; i < numFake; i += 1 {
		if i != pi {
			K.keys[i] = createFakePublicKeyArray(m)
		} else {
			K.keys[pi] = make([]privacy.Point, m)
			for j := 0; j < m; j += 1 {
				K.keys[i][j] = *parsePublicKey(privateKeys[j])
			}
		}
	}
	return
}
