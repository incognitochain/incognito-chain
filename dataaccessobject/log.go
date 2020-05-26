package dataaccessobject

import "github.com/incognitochain/incognito-chain/common"

type DAOLogger struct {
	Log common.Logger
}

func (dAOLogger *DAOLogger) Init(inst common.Logger) {
	dAOLogger.Log = inst
}

// Global instant to use
var Logger = DAOLogger{}
