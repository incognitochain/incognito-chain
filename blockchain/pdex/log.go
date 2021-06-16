package pdex

import "github.com/incognitochain/incognito-chain/common"

type PDEXLogger struct {
	common.Logger
}

func (pDEXLogger *PDEXLogger) Init(logger common.Logger) {
	pDEXLogger.Logger = logger
}

// Global instant to use
var Logger = PDEXLogger{}
