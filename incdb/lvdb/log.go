package lvdb

import (
	"github.com/incognitochain/incognito-chain/common"
)

type LVDBLogger struct {
	log common.Logger
}

func (lvdbLogger *LVDBLogger) Init(inst common.Logger) {
	lvdbLogger.log = inst
}

// Global instant to use
var Logger = LVDBLogger{}
