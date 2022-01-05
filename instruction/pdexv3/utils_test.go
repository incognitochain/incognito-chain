package pdexv3

import (
	"testing"

	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/stretchr/testify/assert"
)

const (
	validOTAReceiver0 = "15sXoyo8kCZCHjurNC69b8WV2jMCvf5tVrcQ5mT1eH9Nm351XRjE1BH4WHHLGYPZy9dxTSLiKQd6KdfoGq4yb4gP1AU2oaJTeoGymsEzonyi1XSW2J2U7LeAVjS1S2gjbNDk1t3f9QUg2gk4"
	validOTAReceiver1 = "15ujixNQY1Qc5wyX9UYQW3s6cbcecFPNhrWjWiFCggeN5HukPVdjbKyRE3goUpFgZhawtBtRUK3ZSZb5LtH7bevhGzz3UTh1muzLHG3pvsE6RNB81y8xNGhyHdpHZfjwmSWDdwDe74Tg2CUP"
	validAccessOTA    = "5xbO6s+gO8pn/Irevhdy6l7S3A64oKGKkAENpRTI5MA="
)

var (
	accessOTA metadataPdexv3.AccessOTA
)

func initTestParam(t *testing.T) {
	err := accessOTA.FromString(validAccessOTA)
	assert.Nil(t, err)
}
