package portalprocess

import "github.com/incognitochain/incognito-chain/common"

type PortalProcessLoggerV4 struct {
	log common.Logger
}

func (p *PortalProcessLoggerV4) Init(inst common.Logger) {
	p.log = inst
}

// Global instant to use
var Logger = PortalProcessLoggerV4{}