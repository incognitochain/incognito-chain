package privacy

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

/*
	Unit test for Hybrid encryption
*/
func TestHybridEncryption(t *testing.T) {
	// random message
	msg := RandBytes(100)

	// generate key pair for ElGamal
	privKey := RandScalar()
	publicKey := PedCom.G[0].ScalarMult(privKey)

	// encrypt message using public key
	ciphertext, err := hybridEncrypt(msg, publicKey)

	assert.Equal(t, nil, err)
	assert.Equal(t, elGamalCiphertextSize, len(ciphertext.symKeyEncrypted))
	assert.Greater(t, len(ciphertext.msgEncrypted), 0)

	// convert hybridCiphertext to bytes array
	ciphertextBytes := ciphertext.Bytes()

	assert.Greater(t, len(ciphertextBytes), elGamalCiphertextSize)

	// new hybridCiphertext to set bytes array
	ciphertext2 := new(hybridCiphertext)
	err2 := ciphertext2.SetBytes(ciphertextBytes)

	assert.Equal(t, nil, err2)
	assert.Equal(t, ciphertext, ciphertext2)

	// decrypt message using private key
	msg2, err := hybridDecrypt(ciphertext2, privKey)

	assert.Equal(t, nil, err)
	assert.Equal(t, msg, msg2)
}
