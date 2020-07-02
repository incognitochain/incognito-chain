package oneoutofmany

import "github.com/incognitochain/incognito-chain/common"

type OneoutofmanyLogger struct {
	Log common.Logger
}

func (logger *OneoutofmanyLogger) Init(inst common.Logger) {
	logger.Log = inst
}

// Global instant to use
var Logger = OneoutofmanyLogger{}

var _ = func() (_ struct{}) {
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()
