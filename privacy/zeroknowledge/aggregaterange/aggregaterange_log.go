package aggregaterange

import "github.com/incognitochain/incognito-chain/common"

type AggregaterangeLogger struct {
	Log common.Logger
}

func (logger *AggregaterangeLogger) Init(inst common.Logger) {
	logger.Log = inst
}

// Global instant to use
var Logger = AggregaterangeLogger{}
