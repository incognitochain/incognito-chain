package portal

import "github.com/incognitochain/incognito-chain/common"

type PortalLogger struct {
	log common.Logger
}

func (p *PortalLogger) Init(inst common.Logger) {
	p.log = inst
}

// Global instant to use
var Logger = PortalLogger{}