package portal

import "github.com/incognitochain/incognito-chain/common"

type PortalLogger struct {
	log common.Logger
}

func (portalLogger *PortalLogger) Init(inst common.Logger) {
	portalLogger.log = inst
}

// Global instant to use
var Logger = PortalLogger{}