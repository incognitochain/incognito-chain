package random

import "github.com/incognitochain/incognito-chain/common"

type RandomLogger struct {
	log common.Logger
}

func (self *RandomLogger) Init(inst common.Logger) {
	self.log = inst
}

// Global instant to use
var Logger = RandomLogger{}
