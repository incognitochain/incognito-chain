package btcrelaying

import "github.com/incognitochain/incognito-chain/common"

type BTCRelayingLogger struct {
	log common.Logger
}

func (self *BTCRelayingLogger) Init(inst common.Logger) {
	self.log = inst
}

// Global instant to use
var Logger = BTCRelayingLogger{}
