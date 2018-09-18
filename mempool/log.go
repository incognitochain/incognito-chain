package mempool

import "github.com/ninjadotorg/cash-prototype/common"

type MempoolLogger struct {
	log common.Logger
}

func (self *MempoolLogger) Init(inst common.Logger) {
	self.log = inst
}

// Global instant to use
var Logger = MempoolLogger{}
