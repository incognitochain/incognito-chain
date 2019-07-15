package privacy

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

/*
	Unit test for elgamal encryption
 */
func TestElGamalEncryption(t *testing.T) {
	// generate private key
	privKey := new(ElGamalPrivKey)
	privKey.X = RandScalar()

	// generate public key
	pubKey := new(ElGamalPubKey)
	pubKey.H = new(EllipticPoint)
	pubKey.H.Set(Curve.Params().Gx, Curve.Params().Gy)
	pubKey.H = pubKey.H.ScalarMult(privKey.X)

	// random message (msg is an elliptic point)
	message := new(EllipticPoint)
	message.Randomize()

	// Encrypt message using public key
	ciphertext1 := pubKey.Encrypt(message)

	// convert ciphertext1 to bytes array
	ciphertext1Bytes := ciphertext1.Bytes()

	// new ciphertext2
	ciphertext2 := new(ElGamalCiphertext)
	ciphertext2.SetBytes(ciphertext1Bytes)

	assert.Equal(t, ciphertext1, ciphertext2)

	// decrypt ciphertext using privateKey
	decryptedCiphertext, err := privKey.Decrypt(ciphertext2)

	assert.Equal(t, nil, err)
	assert.Equal(t, message, decryptedCiphertext)
}
