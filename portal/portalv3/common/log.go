package common

import "github.com/incognitochain/incognito-chain/common"

type PortalCommonLoggerV3 struct {
	log common.Logger
}

func (portalCommonLogger *PortalCommonLoggerV3) Init(inst common.Logger) {
	portalCommonLogger.log = inst
}

// Global instant to use
var Logger = PortalCommonLoggerV3{}