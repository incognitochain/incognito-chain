package mlsag

import (
	"github.com/incognitochain/incognito-chain/privacy"
)

type Ring struct {
	keys [][]privacy.Point
}

func createFakePublicKeyArray(length int) *[]privacy.Point {
	K := make([]privacy.Point, length)
	for i := 0; i < length; i += 1 {
		K[i] = *privacy.RandomPoint()
	}
	return &K
}

// Create a random ring with dimension: (numFake; len(privateKeys)) where we generate fake public keys inside
func NewRandomRing(privateKeys *[]privacy.Scalar, numFake, pi int) (K *Ring) {
	priv := *privateKeys
	m := len(priv)

	K = new(Ring)
	K.keys = make([][]privacy.Point, numFake)
	for i := 0; i < numFake; i += 1 {
		if i != pi {
			K.keys[i] = *createFakePublicKeyArray(m)
		} else {
			K.keys[pi] = make([]privacy.Point, m)
			for j := 0; j < m; j += 1 {
				K.keys[i][j] = *parsePublicKey(priv[j])
			}
		}
	}
	return
}

func (this *Ring) AppendToRow(row int, val *privacy.Point) {
	this.keys[row] = append(this.keys[row], *val)
}
