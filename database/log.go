package database

import (
	"github.com/ninjadotorg/constant/common"
)

type DbLogger struct {
	log common.Logger
}

func (self *DbLogger) Init(inst common.Logger) {
	self.log = inst
}

// Global instant to use
var Logger = DbLogger{}
