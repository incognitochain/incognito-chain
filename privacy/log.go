package privacy

import "github.com/ninjadotorg/constant/common"

type PrivacyLogger struct {
	Log common.Logger
}

func (self *PrivacyLogger) Init(inst common.Logger) {
	self.Log = inst
}

// Global instant to use
var Logger = PrivacyLogger{}
