package portalprocess

import "github.com/incognitochain/incognito-chain/common"

type PortalInstLogger struct {
	log common.Logger
}

func (metricLogger *PortalInstLogger) Init(inst common.Logger) {
	metricLogger.log = inst
}

// Global instant to use
var Logger = PortalInstLogger{}
