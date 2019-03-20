package addrmanager

import "github.com/constant-money/constant-chain/common"

type AddrManagerLogger struct {
	log common.Logger
}

func (addrManagerLogger *AddrManagerLogger) Init(inst common.Logger) {
	addrManagerLogger.log = inst
}

// Global instant to use
var Logger = AddrManagerLogger{}
