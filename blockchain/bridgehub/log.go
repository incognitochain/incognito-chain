package bridgehub

import "github.com/incognitochain/incognito-chain/common"

type BrideHubLogger struct {
	log common.Logger
}

func (bridgeHubLogger *BrideHubLogger) Init(logger common.Logger) {
	bridgeHubLogger.log = logger
}

// Global instant to use
var Logger = BrideHubLogger{}
