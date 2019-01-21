package transaction

import "github.com/ninjadotorg/constant/common"

type TransactionLogger struct {
	log common.Logger
}

func (transactionLogger *TransactionLogger) Init(inst common.Logger) {
	transactionLogger.log = inst
}

// Global instant to use
var Logger = TransactionLogger{}
