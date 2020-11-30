package common

import "github.com/incognitochain/incognito-chain/common"

type PortalCommonLogger struct {
	log common.Logger
}

func (metricLogger *PortalCommonLogger) Init(inst common.Logger) {
	metricLogger.log = inst
}

// Global instant to use
var Logger = PortalCommonLogger{}
