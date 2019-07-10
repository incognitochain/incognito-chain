package privacy

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSchnorrSignature(t *testing.T) {
	// generate Schnorr Private Key
	privKey := new(SchnPrivKey)
	privKey.SK = RandScalar()
	privKey.R = RandScalar()

	// generate Schnorr Public Key
	privKey.PubKey = new(SchnPubKey)

	// G is base generator on Curve
	privKey.PubKey.G = new(EllipticPoint)
	privKey.PubKey.G.Set(Curve.Params().Gx, Curve.Params().Gy)

	// H = alpha*G
	privKey.PubKey.H = privKey.PubKey.G.ScalarMult(RandScalar())

	// PK = G^SK * H^R
	privKey.PubKey.PK = privKey.PubKey.G.ScalarMult(privKey.SK).Add(privKey.PubKey.H.ScalarMult(privKey.R))

	// random message to sign
	data := RandScalar()

	// sign on message
	signature, err := privKey.Sign(data.Bytes())
	assert.Equal(t, nil, err)

	// convert signature to bytes array
	signatureBytes := signature.Bytes()
	assert.Equal(t, SigPrivacySize, len(signatureBytes))

	// revert bytes array to signature
	signature2 := new(SchnSignature)
	signature2.SetBytes(signatureBytes)
	assert.Equal(t, signature, signature2)

	// verify the signature with private key
	res := privKey.PubKey.Verify(signature2, data.Bytes())
	assert.Equal(t, true, res)
}
