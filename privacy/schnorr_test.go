package privacy

import (
	"crypto/rand"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/stretchr/testify/assert"
)

func TestSchnorrSignature(t *testing.T) {
	// generate Schnorr Private Key
	var r = rand.Reader
	privKey := new(SchnorrPrivateKey)
	privKey.privateKey = RandScalar(r)
	privKey.randomness = RandScalar(r)

	// generate Schnorr Public Key
	privKey.publicKey = new(SchnorrPublicKey)

	// G is base generator on Curve
	privKey.publicKey.g = new(EllipticPoint)
	privKey.publicKey.g.Set(Curve.Params().Gx, Curve.Params().Gy)

	// H = alpha*G
	privKey.publicKey.h = privKey.publicKey.g.ScalarMult(RandScalar(r))

	// PK = G^SK * H^R
	privKey.publicKey.publicKey = privKey.publicKey.g.ScalarMult(privKey.privateKey).Add(privKey.publicKey.h.ScalarMult(privKey.randomness))

	// random message to sign
	data := RandScalar(r)

	// sign on message
	signature, err := privKey.Sign(data.Bytes())
	assert.Equal(t, nil, err)

	// convert signature to bytes array
	signatureBytes := signature.Bytes()
	assert.Equal(t, common.SigPrivacySize, len(signatureBytes))

	// revert bytes array to signature
	signature2 := new(SchnSignature)
	signature2.SetBytes(signatureBytes)
	assert.Equal(t, signature, signature2)

	// verify the signature with private key
	res := privKey.publicKey.Verify(signature2, data.Bytes())
	assert.Equal(t, true, res)
}
