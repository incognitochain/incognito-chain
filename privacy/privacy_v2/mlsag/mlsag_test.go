package mlsag

import (
	"bytes"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/stretchr/testify/assert"
)

func TestWorkflowMlsag(t *testing.T) {
	keyInputs := []*operation.Scalar{}
	for i := 0; i < 8; i += 1 {
		privateKey := operation.RandomScalar()
		keyInputs = append(keyInputs, privateKey)
	}
	numFake := 5
	pi := common.RandInt() % numFake
	ring := NewRandomRing(keyInputs, numFake, pi)
	signer := NewMlsag(keyInputs, ring, pi)

	signature, err := signer.Sign([]byte("Hello"))
	assert.Equal(t, nil, err, "There should not be any error when sign")

	check, err := Verify(signature, ring, []byte("Hello"))
	assert.Equal(t, nil, err, "There should not be any error when verify")
	assert.Equal(t, true, check, "It should verify correctly")

	b, err := signature.ToHex()
	assert.Equal(t, nil, err, "There should not be any error when to hex the signature")
	sig, err := new(MlsagSig).FromHex(b)
	assert.Equal(t, nil, err, "There should not be any error when from hex the signature")

	b1, err := sig.ToBytes()
	assert.Equal(t, nil, err, "There should not be any error when to bytes the signature")
	b2, _ := signature.ToBytes()
	assert.Equal(t, nil, err, "There should not be any error when to bytes the signature")
	assert.Equal(t, true, bytes.Equal(b1, b2), "There should not be any error when to bytes the signature")
}
