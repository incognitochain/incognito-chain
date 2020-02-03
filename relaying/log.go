package relaying

import "github.com/incognitochain/incognito-chain/common"

type RelayingLogger struct {
	Log common.Logger
}

func (logger *RelayingLogger) Init(inst common.Logger) {
	logger.Log = inst
}

// Global instant to use
var Logger = RelayingLogger{}
