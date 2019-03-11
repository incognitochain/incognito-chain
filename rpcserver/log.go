package rpcserver

import "github.com/big0t/constant-chain/common"

type RpcLogger struct {
	log common.Logger
}

func (rpcLogger *RpcLogger) Init(inst common.Logger) {
	rpcLogger.log = inst
}

// Global instant to use
var Logger = RpcLogger{}
