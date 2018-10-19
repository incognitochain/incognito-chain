package netsync

import "github.com/ninjadotorg/cash/common"

type NetSyncLogger struct {
	log common.Logger
}

func (self *NetSyncLogger) Init(inst common.Logger) {
	self.log = inst
}

// Global instant to use
var Logger = NetSyncLogger{}
