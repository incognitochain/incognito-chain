package aggregatedrange

import "github.com/incognitochain/incognito-chain/common"

type aggregatedrangeLogger struct {
	Log common.Logger
}

func (logger *aggregatedrangeLogger) Init(inst common.Logger) {
	logger.Log = inst
}

// Global instant to use
var Logger = aggregatedrangeLogger{}

var _ = func() (_ struct{}) {
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()
