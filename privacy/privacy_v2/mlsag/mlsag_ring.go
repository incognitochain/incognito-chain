package mlsag

import (
	"github.com/incognitochain/incognito-chain/privacy/operation"
)

type Ring struct {
	keys [][]operation.Point
}

func createFakePublicKeyArray(length int) *[]operation.Point {
	K := make([]operation.Point, length)
	for i := 0; i < length; i += 1 {
		K[i] = *operation.RandomPoint()
	}
	return &K
}

// Create a random ring with dimension: (numFake; len(privateKeys)) where we generate fake public keys inside
func NewRandomRing(privateKeys *[]operation.Scalar, numFake, pi int) (K *Ring) {
	priv := *privateKeys
	m := len(priv)

	K = new(Ring)
	K.keys = make([][]operation.Point, numFake)
	for i := 0; i < numFake; i += 1 {
		if i != pi {
			K.keys[i] = *createFakePublicKeyArray(m)
		} else {
			K.keys[pi] = make([]operation.Point, m)
			for j := 0; j < m; j += 1 {
				K.keys[i][j] = *parsePublicKey(priv[j])
			}
		}
	}
	return
}

func (this *Ring) AppendToRow(row int, val *operation.Point) {
	this.keys[row] = append(this.keys[row], *val)
}
