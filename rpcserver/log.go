package rpcserver

import "github.com/incognitochain/incognito-chain/common"

type RpcLogger struct {
	log common.Logger
}

func (rpcLogger *RpcLogger) Init(inst common.Logger) {
	rpcLogger.log = inst
}

// Global instant to use
var Logger = RpcLogger{}
