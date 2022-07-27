package pruner

import "github.com/incognitochain/incognito-chain/common"

type PrunerLogger struct {
	log common.Logger
}

func (prunerLogger *PrunerLogger) Init(logger common.Logger) {
	prunerLogger.log = logger
}

// Global instant to use
var Logger = PrunerLogger{}
