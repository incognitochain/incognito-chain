package mempool

import "github.com/ninjadotorg/constant/common"

type MempoolLogger struct {
	log common.Logger
}

func (self *MempoolLogger) Init(inst common.Logger) {
	self.log = inst
}

// Global instant to use
var Logger = MempoolLogger{}
