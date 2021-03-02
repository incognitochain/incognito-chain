package portaltokens

import "github.com/incognitochain/incognito-chain/common"

type PortalTokenLoggerV4 struct {
	log common.Logger
}

func (portalTokenLogger *PortalTokenLoggerV4) Init(inst common.Logger) {
	portalTokenLogger.log = inst
}

// Global instant to use
var Logger = PortalTokenLoggerV4{}