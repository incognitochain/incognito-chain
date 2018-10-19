package wallet

import "github.com/ninjadotorg/cash/common"

type WalletLogger struct {
	log common.Logger
}

func (self *WalletLogger) Init(inst common.Logger) {
	self.log = inst
}

// Global instant to use
var Logger = WalletLogger{}
