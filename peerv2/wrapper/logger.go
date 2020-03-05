package wrapper

import "github.com/incognitochain/incognito-chain/common"

type WrapperLogger struct {
	common.Logger
}

func (self *WrapperLogger) Init(inst common.Logger) {
	self.Logger = inst
}

// Global instant to use
var Logger = WrapperLogger{}
