package transaction

import "github.com/ninjadotorg/constant/common"

type TransactionLogger struct {
	log common.Logger
}

func (self *TransactionLogger) Init(inst common.Logger) {
	self.log = inst
}

// Global instant to use
var Logger = TransactionLogger{}
