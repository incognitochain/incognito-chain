package ppos

import "github.com/ninjadotorg/cash/common"

type PoSLogger struct {
	log common.Logger
}

func (self *PoSLogger) Init(inst common.Logger) {
	self.log = inst
}

// Global instant to use
var Logger = PoSLogger{}
