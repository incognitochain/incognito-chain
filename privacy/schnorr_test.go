package privacy

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSchn(t *testing.T) {
	privKey := new(SchnPrivKey)

	privKey.SK = RandBigInt()
	privKey.R = RandBigInt()

	privKey.PubKey = new(SchnPubKey)

	privKey.PubKey.G = new(EllipticPoint)
	privKey.PubKey.G.Set(Curve.Params().Gx, Curve.Params().Gy)

	privKey.PubKey.H = privKey.PubKey.G.ScalarMult(RandBigInt())
	rH := privKey.PubKey.H.ScalarMult(privKey.R)

	privKey.PubKey.PK = privKey.PubKey.G.ScalarMult(privKey.SK).Add(rH)

	data := RandBigInt()

	signature, _ := privKey.Sign(data.Bytes())
	signature.SetBytes(signature.Bytes())

	res := privKey.PubKey.Verify(signature, data.Bytes())

	assert.Equal(t, true, res)
}
