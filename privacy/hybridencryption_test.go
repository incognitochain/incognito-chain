package privacy

import (
	"crypto/rand"
	"github.com/stretchr/testify/assert"
	"testing"
)

/*
	Unit test for Hybrid encryption
*/
func TestHybridEncryption(t *testing.T) {
	for i := 0; i < 5000; i++ {
		// random message
		msg := randomMessage()

		// generate key pair for ElGamal
		privKey := new(elGamalPrivateKey)
		privKey.x = RandomScalar()

		// generate public key
		pubKey := new(elGamalPublicKey)
		pubKey.h = new(Point).ScalarMultBase(privKey.x)

		// encrypt message using public key
		ciphertext, err := HybridEncrypt(msg, pubKey.h)

		assert.Equal(t, nil, err)

		// convert HybridCipherText to bytes array
		ciphertextBytes := ciphertext.Bytes()

		// new HybridCipherText to set bytes array
		ciphertext2 := new(HybridCipherText)
		err2 := ciphertext2.SetBytes(ciphertextBytes)

		assert.Equal(t, nil, err2)
		assert.Equal(t, ciphertext, ciphertext2)

		// decrypt message using private key
		msg2, err := HybridDecrypt(ciphertext2, privKey.x)

		assert.Equal(t, nil, err)
		assert.Equal(t, msg, msg2)
	}
}

func randomMessage() []byte {
	msg := make([]byte, 128)
	rand.Read(msg)
	return msg
}
