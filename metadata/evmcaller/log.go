package evmcaller

import "github.com/incognitochain/incognito-chain/common"

type EVMCallerLogger struct {
	log common.Logger
}

func (metricLogger *EVMCallerLogger) Init(inst common.Logger) {
	metricLogger.log = inst
}

// Global instant to use
var Logger = EVMCallerLogger{}
