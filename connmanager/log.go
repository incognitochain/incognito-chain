package connmanager

import (
	"github.com/ninjadotorg/cash-prototype/common"
)

type ConnManagerLogger struct {
	logger common.Logger
}

func (self *ConnManagerLogger) Init(inst common.Logger) {
	self.logger = inst
}

// Global instant to use
var Logger = ConnManagerLogger{}
