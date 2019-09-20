package privacy

import (
	"crypto/rand"
	"github.com/stretchr/testify/assert"
	"testing"
)

/*
	Unit test for elgamal encryption
*/
func TestElGamalEncryption(t *testing.T) {
	for i := 0; i < 1000; i ++ {
		// generate private key
		privKey := new(elGamalPrivateKeyOld)
		var r= rand.Reader
		privKey.x = RandScalar(r)

		// generate public key
		pubKey := new(elGamalPublicKeyOld)
		pubKey.h = new(EllipticPoint)
		pubKey.h.Set(Curve.Params().Gx, Curve.Params().Gy)
		pubKey.h = pubKey.h.ScalarMult(privKey.x)

		// random message (msg is an elliptic point)
		message := new(EllipticPoint)
		message.Randomize()

		// Encrypt message using public key
		ciphertext1 := pubKey.encrypt(message)

		// convert ciphertext1 to bytes array
		ciphertext1Bytes := ciphertext1.Bytes()

		// new ciphertext2
		ciphertext2 := new(elGamalCipherTextOld)
		ciphertext2.SetBytes(ciphertext1Bytes)

		assert.Equal(t, ciphertext1, ciphertext2)

		// decrypt ciphertext using privateKey
		decryptedCiphertext, err := privKey.decrypt(ciphertext2)

		assert.Equal(t, nil, err)
		assert.Equal(t, message, decryptedCiphertext)
	}
}
