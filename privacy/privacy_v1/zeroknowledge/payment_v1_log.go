package zkp

import (
	"fmt"

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

var _ = func() (_ struct{}) {
	fmt.Println("This runs before init()!")
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()
