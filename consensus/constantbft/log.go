package constantbft

import "github.com/ninjadotorg/constant/common"

type constantbftLogger struct {
	log common.Logger
}

func (self *constantbftLogger) Init(inst common.Logger) {
	self.log = inst
}

// Global instant to use
var Logger = constantbftLogger{}
