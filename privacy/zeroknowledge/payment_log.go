package zkp

import "github.com/incognitochain/incognito-chain/common"

type PaymentLogger struct {
	Log common.Logger
}

func (logger *PaymentLogger) Init(inst common.Logger) {
	logger.Log = inst
}

// Global instant to use
var Logger = PaymentLogger{}
