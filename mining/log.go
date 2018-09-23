package mining

import "github.com/ninjadotorg/cash-prototype/common"

type MiningLogger struct {
	log common.Logger
}

func (self *MiningLogger) Init(inst common.Logger) {
	self.log = inst
}

// Global instant to use
var Logger = MiningLogger{}
