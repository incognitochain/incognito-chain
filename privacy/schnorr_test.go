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
	privKey.publicKey = new(SchnorrPubKey)

	// G is base generator on Curve
	privKey.publicKey.G = new(EllipticPoint)
	privKey.publicKey.G.Set(Curve.Params().Gx, Curve.Params().Gy)

	// H = alpha*G
	privKey.publicKey.H = privKey.publicKey.G.ScalarMult(RandScalar())

	// PK = G^SK * H^R
	privKey.publicKey.PK = privKey.publicKey.G.ScalarMult(privKey.secretKey).Add(privKey.publicKey.H.ScalarMult(privKey.r))

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
