package rpcserver

import "github.com/incognitochain/incognito-chain/common"

type RpcLogger struct {
	log common.Logger
}

func (rpcLogger *RpcLogger) Init(inst common.Logger) {
	rpcLogger.log = inst
}

type DeBridgeLogger struct {
	log common.Logger
}

func (self *DeBridgeLogger) Init(inst common.Logger) {
	self.log = inst
}

// Global instant to use
var Logger = RpcLogger{}
var BLogger = DeBridgeLogger{}
