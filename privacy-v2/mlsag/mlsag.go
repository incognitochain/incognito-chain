// MAIN IMPLEMENTATION OF MLSAG

package mlsag

import (
	"github.com/incognitochain/incognito-chain/common"
	C25519 "github.com/incognitochain/incognito-chain/privacy/curve25519"
)

type Ring struct {
	keys [][]C25519.Key
}

type Mlsag struct {
	K           *Ring
	pi          int
	keyImages   []C25519.Key
	privateKeys []C25519.Key
}

func (this *Ring) isEmpty() bool {
	return len(this.keys) == 0
}

func (this *Mlsag) isEmpty() bool {
	return this.K.isEmpty() || len(this.privateKeys) == 0
}

func createFakePublicKey() C25519.Key {
	privateKey := C25519.RandomScalar()
	publicKey := C25519.ScalarmultBase(privateKey)
	return publicKey.ToBytes()
}

func createFakePublicKeyArray(length int) (K []C25519.Key) {
	K = make([]C25519.Key, length)
	for i := 0; i < length; i += 1 {
		K[i] = createFakePublicKey()
	}
	return
}

// Create a random ring with dimension: (numFake; len(privateKeys)) where we generate fake public keys inside
func NewRing(privateKeys []C25519.Key, numFake, pi int) (K *Ring) {
	var m int = len(privateKeys)

	if K == nil {
		K = new(Ring)
	}

	K.keys = make([][]C25519.Key, numFake)
	for i := 0; i < numFake; i += 1 {
		if i != pi {
			K.keys[i] = createFakePublicKeyArray(m)
		} else {
			K.keys[pi] = make([]C25519.Key, m)
			for j := 0; j < m; j += 1 {
				K.keys[i][j] = parsePublicKey(privateKeys[j])
			}
		}
	}

	return
}

func createRandomChallenges(n, m, pi int) (alpha []C25519.Key, r [][]C25519.Key) {
	alpha = make([]C25519.Key, m)
	for i := 0; i < m; i += 1 {
		alpha[i] = *C25519.RandomScalar()
	}
	r = make([][]C25519.Key, n)
	for i := 0; i < n; i += 1 {
		r[i] = make([]C25519.Key, m)
		if i == pi {
			continue
		}
		for j := 0; j < m; j += 1 {
			r[i][j] = *C25519.RandomScalar()
		}
	}
	return
}

// func createCHash()

func NewMlsagWithDefinedRing(privateKeys []C25519.Key, K Ring, numFake int) (mlsag *Mlsag) {
	if mlsag == nil {
		mlsag = new(Mlsag)
	}

	mlsag.privateKeys = privateKeys
	mlsag.pi = common.RandInt() % numFake
	mlsag.keyImages = parseKeyImages(privateKeys) // 1st step in monero paper
	mlsag.K = NewRing(privateKeys, numFake, mlsag.pi)

	return
}

func NewMlsagWithRandomRing(privateKeys []C25519.Key, numFake int) (mlsag *Mlsag) {
	if mlsag == nil {
		mlsag = new(Mlsag)
	}

	mlsag.privateKeys = privateKeys
	mlsag.pi = common.RandInt() % numFake
	mlsag.keyImages = parseKeyImages(privateKeys) // 1st step in monero paper
	mlsag.K = NewRing(privateKeys, numFake, mlsag.pi)

	return
}

func (mlsag *Mlsag) Sign(message string) {
	digest := hashToNum([]byte(message))

}

func Sign(mlsag Mlsag, message string) {
	return mlsag.Sign(message)
}

// SignWithRandomRing will generate random ring and use MLSAG to sign the message.
func SignWithRandomRing(privateKeys []C25519.Key, message string, numFake int) []byte {
	mlsag := NewMlsagWithRandomRing(privateKeys, numFake)
	return mlsag.Sign(message)
}
