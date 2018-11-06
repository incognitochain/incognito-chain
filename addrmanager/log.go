package addrmanager

import "github.com/ninjadotorg/constant/common"

type AddrManagerLogger struct {
	log common.Logger
}

func (self *AddrManagerLogger) Init(inst common.Logger) {
	self.log = inst
}

// Global instant to use
var Logger = AddrManagerLogger{}
