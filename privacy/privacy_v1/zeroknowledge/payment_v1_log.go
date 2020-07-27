package zkp

import (
	"github.com/incognitochain/incognito-chain/common"
)

type PaymentV1Logger struct {
	Log common.Logger
}

func (logger *PaymentV1Logger) Init(inst common.Logger) {
	logger.Log = inst
}

// Global instant to use
var Logger = PaymentV1Logger{}