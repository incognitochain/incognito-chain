package btcapi

import "github.com/ninjadotorg/constant/common"

type RandomLogger struct {
	log common.Logger
}

func (self *RandomLogger) Init(inst common.Logger) {
	self.log = inst
}

// Global instant to use
var Logger = RandomLogger{}
