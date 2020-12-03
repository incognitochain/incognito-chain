package common

import "github.com/incognitochain/incognito-chain/common"

type PortalCommonLogger struct {
	log common.Logger
}

func (portalCommonLogger *PortalCommonLogger) Init(inst common.Logger) {
	portalCommonLogger.log = inst
}

// Global instant to use
var Logger = PortalCommonLogger{}
