package privacy_v2

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
)

type PaymentV2Logger struct {
	Log common.Logger
}

func (logger *PaymentV2Logger) Init(inst common.Logger) {
	logger.Log = inst
}

const (
	ConversionProofVersion = 255
)

// Global instant to use
var Logger = PaymentV2Logger{}

var _ = func() (_ struct{}) {
	fmt.Println("This runs before init()!")
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()
