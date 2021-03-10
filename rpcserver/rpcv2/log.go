package rpcv2

import (
	"github.com/incognitochain/incognito-chain/common"
)

type RpcV2Logger struct {
	Log common.Logger
}

func (transactionLogger *RpcV2Logger) Init(inst common.Logger) {
	transactionLogger.Log = inst
}

// Global instant to use
var Logger = RpcV2Logger{}
