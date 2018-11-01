package rpcserver

import "github.com/ninjadotorg/constant/common"

type RpcLogger struct {
	log common.Logger
}

func (self *RpcLogger) Init(inst common.Logger) {
	self.log = inst
}

// Global instant to use
var Logger = RpcLogger{}
