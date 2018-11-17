package constantpos

import "github.com/ninjadotorg/constant/common"

type constantposLogger struct {
	log common.Logger
}

func (self *constantposLogger) Init(inst common.Logger) {
	self.log = inst
}

// Global instant to use
var Logger = constantposLogger{}
