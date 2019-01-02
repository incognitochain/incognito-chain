package privacy

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSchn(t *testing.T) {
	schnPrivKey := new(SchnPrivKey)
	schnPrivKey.GenKey()
	fmt.Printf("Private key SK: %v\n", schnPrivKey.SK.Bytes())
	fmt.Printf("Private key R: %v\n", schnPrivKey.R.Bytes())
	fmt.Printf("Public key G: %v\n", schnPrivKey.PubKey.G.Compress())
	fmt.Printf("Public key H: %v\n", schnPrivKey.PubKey.H.Compress())

	data := RandBytes(SpendingKeySize)
	fmt.Printf("Data: %v\n", data)

	signature, _ := schnPrivKey.Sign(data)
	fmt.Printf("Signature: %+v\n", signature.Bytes())

	res := schnPrivKey.PubKey.Verify(signature, data)

	assert.Equal(t, true, res)
}
