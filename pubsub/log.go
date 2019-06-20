package pubsub

import (
	"github.com/incognitochain/incognito-chain/common"
)

type PubsubLogger struct {
	log common.Logger
}

func (metricLogger *PubsubLogger) Init(inst common.Logger) {
	metricLogger.log = inst
}

// Global instant to use
var Logger = PubsubLogger{}

