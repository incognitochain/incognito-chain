package privacy

import (
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/stretchr/testify/assert"
)

func TestSchnorrSignature(t *testing.T) {
	// generate Schnorr Private Key
	privKey := new(SchnorrPrivateKey)
	privKey.secretKey = RandScalar()
	privKey.r = RandScalar()

	// generate Schnorr Public Key
	privKey.publicKey = new(SchnorrPublicKey)

	// G is base generator on Curve
	privKey.publicKey.g = new(EllipticPoint)
	privKey.publicKey.g.Set(Curve.Params().Gx, Curve.Params().Gy)

	// H = alpha*G
	privKey.publicKey.h = privKey.publicKey.g.ScalarMult(RandScalar())

	// PK = G^SK * H^R
	privKey.publicKey.publicKey = privKey.publicKey.g.ScalarMult(privKey.secretKey).Add(privKey.publicKey.h.ScalarMult(privKey.r))

	// random message to sign
	data := RandScalar()

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
