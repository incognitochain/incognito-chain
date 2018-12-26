package privacy

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestElGamalEncryption(t *testing.T) {
	privKey := new(ElGamalPrivKey)
	privKey.Curve = &Curve
	privKey.X = RandInt()

	pubKey := privKey.GenPubKey()

	message := new(EllipticPoint)
	message.Randomize()

	ciphertext1 := pubKey.Encrypt(message)

	ciphertext2 := new(ElGamalCiphertext)
	ciphertext2.SetBytes(ciphertext1.Bytes())

	decryptedCiphertext := privKey.Decrypt(ciphertext2)

	assert.Equal(t, message, decryptedCiphertext)
}
