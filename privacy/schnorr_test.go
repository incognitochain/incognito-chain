package privacy

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSchn(t *testing.T) {
	schnPrivKey := new(SchnPrivKey)
	schnPrivKey.GenKey()

	data := RandInt()

	signature, _ := schnPrivKey.Sign(data.Bytes())
	signature.SetBytes(signature.Bytes())

	res := schnPrivKey.PubKey.Verify(signature, data.Bytes())

	assert.Equal(t, true, res)
}
