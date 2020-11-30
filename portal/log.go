package portal

import "github.com/incognitochain/incognito-chain/common"

type PortalLogger struct {
	log common.Logger
}

func (metricLogger *PortalLogger) Init(inst common.Logger) {
	metricLogger.log = inst
}

// Global instant to use
var Logger = PortalLogger{}