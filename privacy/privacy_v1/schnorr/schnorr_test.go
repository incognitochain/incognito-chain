package schnorr

import (
	"testing"

	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/stretchr/testify/assert"
)

func TestSchnorrSignature(t *testing.T) {
	for i := 0; i < 100; i++ {
		// generate Schnorr Private Key

		privKey := new(SchnorrPrivateKey)
		privKey.Set(operation.RandomScalar(), operation.RandomScalar())

		// random message to sign
		data := operation.RandomScalar()
		// sign on message
		signature, err := privKey.Sign(data.ToBytesS())
		assert.Equal(t, nil, err)

		// convert signature to bytes array
		signatureBytes := signature.Bytes()

		// revert bytes array to signature
		signature2 := new(SchnSignature)
		signature2.SetBytes(signatureBytes)
		assert.Equal(t, signature, signature2)

		// verify the signature with private key
		res := privKey.publicKey.Verify(signature2, data.ToBytesS())
		assert.Equal(t, true, res)
	}
}

func TestSchnorrSignatureWithoutZ2(t *testing.T) {
	for i := 0; i < 100; i++ {
		// generate Schnorr Private Key

		privKey := new(SchnorrPrivateKey)
		privKey.Set(operation.RandomScalar(), new(Scalar).FromUint64(0))

		// random message to sign
		data := operation.RandomScalar()
		// sign on message
		signature, err := privKey.Sign(data.ToBytesS())
		assert.Equal(t, nil, err)

		// convert signature to bytes array
		signatureBytes := signature.Bytes()

		// revert bytes array to signature
		signature2 := new(SchnSignature)
		signature2.SetBytes(signatureBytes)
		assert.Equal(t, signature, signature2)

		// verify the signature with private key
		res := privKey.publicKey.Verify(signature2, data.ToBytesS())
		assert.Equal(t, true, res)
	}
}
