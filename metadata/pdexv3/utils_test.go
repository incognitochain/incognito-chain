package pdexv3

import (
	"testing"

	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/stretchr/testify/assert"
)

const (
	validOTAReceiver0Str = "15sXoyo8kCZCHjurNC69b8WV2jMCvf5tVrcQ5mT1eH9Nm351XRjE1BH4WHHLGYPZy9dxTSLiKQd6KdfoGq4yb4gP1AU2oaJTeoGymsEzonyi1XSW2J2U7LeAVjS1S2gjbNDk1t3f9QUg2gk4"
	validOTAReceiver1Str = "15ujixNQY1Qc5wyX9UYQW3s6cbcecFPNhrWjWiFCggeN5HukPVdjbKyRE3goUpFgZhawtBtRUK3ZSZb5LtH7bevhGzz3UTh1muzLHG3pvsE6RNB81y8xNGhyHdpHZfjwmSWDdwDe74Tg2CUP"
)

var (
	validOTAReceiver0 = privacy.OTAReceiver{}
	validOTAReceiver1 = privacy.OTAReceiver{}
)

func initTestParam(t *testing.T) {
	err := validOTAReceiver0.FromString(validOTAReceiver0Str)
	assert.Nil(t, err)
	err = validOTAReceiver1.FromString(validOTAReceiver1Str)
	assert.Nil(t, err)
}
