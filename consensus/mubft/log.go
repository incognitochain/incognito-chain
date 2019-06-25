package mubft

import "github.com/incognitochain/incognito-chain/common"

type mubftLogger struct {
	log common.Logger
}

func (mubftLogger *mubftLogger) Init(inst common.Logger) {
	mubftLogger.log = inst
}

// Global instant to use
var Logger = mubftLogger{}
