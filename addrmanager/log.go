package addrmanager

import "github.com/ninjadotorg/cash-prototype/common"

type AddrManagerLogger struct {
	logger common.Logger
}

func (self *AddrManagerLogger) Init(inst common.Logger) {
	self.logger = inst
}

// Global instant to use
var Logger = AddrManagerLogger{}
