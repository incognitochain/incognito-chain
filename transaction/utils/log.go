package utils

import (
	"github.com/incognitochain/incognito-chain/common"
)

type TransactionLogger struct {
	Log common.Logger
}

func (transactionLogger *TransactionLogger) Init(inst common.Logger) {
	transactionLogger.Log = inst
}

// Global instant to use
var Logger = TransactionLogger{}
