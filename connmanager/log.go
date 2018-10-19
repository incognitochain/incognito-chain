package connmanager

import (
	"github.com/ninjadotorg/cash/common"
)

type ConnManagerLogger struct {
	log common.Logger
}

func (self *ConnManagerLogger) Init(inst common.Logger) {
	self.log = inst
}

// Global instant to use
var Logger = ConnManagerLogger{}
