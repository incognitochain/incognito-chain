package portaltokens

import "github.com/incognitochain/incognito-chain/common"

type PortalTokenLogger struct {
	log common.Logger
}

func (metricLogger *PortalTokenLogger) Init(inst common.Logger) {
	metricLogger.log = inst
}

// Global instant to use
var Logger = PortalTokenLogger{}