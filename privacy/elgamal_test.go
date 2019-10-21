package privacy

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

/*
	Unit test for elgamal encryption
*/

func TestElGamalCipherText_Bytes(t *testing.T) {
	privKey := new(elGamalPrivateKey)
	privKey.x = RandomScalar()

	// generate public key
	pubKey := new(elGamalPublicKey)
	pubKey.h = new(Point).ScalarMultBase(privKey.x)

	message := RandomPoint()

	// Encrypt message using public key
	c := pubKey.encrypt(message)
	cBytes := c.Bytes()
	fmt.Println(len(cBytes))
}

func TestElGamalPublicKey_Encryption(t *testing.T) {
	for i := 0; i < 5000; i++ {
		// generate private key
		privKey := new(elGamalPrivateKey)
		privKey.x = RandomScalar()

		// generate public key
		pubKey := new(elGamalPublicKey)
		pubKey.h = new(Point).ScalarMultBase(privKey.x)

		// random message (msg is an elliptic point)
		message := RandomPoint()

		// Encrypt message using public key
		ciphertext1 := pubKey.encrypt(message)

		// convert ciphertext1 to bytes array
		ciphertext1Bytes := ciphertext1.Bytes()

		// new ciphertext2
		ciphertext2 := new(elGamalCipherText)
		ciphertext2.SetBytes(ciphertext1Bytes)

		assert.Equal(t, ciphertext1, ciphertext2)

		// decrypt ciphertext using privateKey
		decryptedCiphertext, err := privKey.decrypt(ciphertext1)

		assert.Equal(t, nil, err)
		assert.Equal(t, message, decryptedCiphertext)
	}
}
