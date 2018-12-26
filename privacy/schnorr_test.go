package privacy

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSchn(t *testing.T) {
	schnPrivKey := new(SchnPrivKey)
	schnPrivKey.KeyGen()

	hash := RandBytes(SpendingKeySize)
	fmt.Printf("Hash: %v\n", hash)

	signature, _ := schnPrivKey.Sign(hash)
	fmt.Printf("Signature: %+v\n", signature)

	res := schnPrivKey.PubKey.Verify(signature, hash)

	assert.Equal(t, true, res)
}
