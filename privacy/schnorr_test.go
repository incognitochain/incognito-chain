package privacy

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSchn(t *testing.T) {
	schnPrivKey := new(SchnPrivKey)
	schnPrivKey.GenKey()

	data := RandBytes(SpendingKeySize)

	signature, _ := schnPrivKey.Sign(data)

	res := schnPrivKey.PubKey.Verify(signature, data)

	assert.Equal(t, true, res)
}
