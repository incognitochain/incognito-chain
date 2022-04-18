package dataaccessobject

import "github.com/incognitochain/incognito-chain/common"

type STATEDBLogger struct {
	Log common.Logger
}

func (dAOLogger *STATEDBLogger) Init(inst common.Logger) {
	dAOLogger.Log = inst
}

// Global instant to use
var Logger = STATEDBLogger{}
