package mlsag

import (
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/stretchr/testify/assert"
	"crypto/rand"
)

// TEST DURATION NOTE : 1000 iterations = 190sec

var (
	maxPrivateKeys = 15
	minPrivateKeys = 1
	maxTotalLayers  = 10
	minTotalLayers  = 1
	numOfLoops     = 1000
)

func TestWorkflowMlsag(t *testing.T) {
	message := make([]byte,32)
	
	for loopCount:=0;loopCount<=numOfLoops;loopCount++{
		keyInputs := []*operation.Scalar{}

		// ring params : #private keys, #fake layers, pi
		// are picked randomly in their domain
		numOfPrivateKeys := common.RandInt() % (maxPrivateKeys-minPrivateKeys+1) + minPrivateKeys
		for i := 0; i < numOfPrivateKeys; i += 1 {
			privateKey := operation.RandomScalar()
			keyInputs = append(keyInputs, privateKey)
		}
		numOfLayers := common.RandInt() % (maxTotalLayers-minTotalLayers+1) + minTotalLayers
		pi := common.RandInt() % numOfLayers
		ring := NewRandomRing(keyInputs, numOfLayers, pi)
		signer := NewMlsag(keyInputs, ring, pi)

		// take a random 20-byte message
		rand.Read(message)


		s := common.HashH(message)
		signature, err := signer.Sign(s[:])
		assert.Equal(t, nil, err, "There should not be any error when sign")

		s2 := common.HashH(message)
		check, err := Verify(signature, ring, s2[:])
		assert.Equal(t, nil, err, "There should not be any error when verify")
		assert.Equal(t, true, check, "It should verify correctly")
	}
}
