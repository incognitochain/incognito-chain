package basemeta

import "github.com/incognitochain/incognito-chain/common"

type BaseMetaLogger struct {
	log common.Logger
}

func (baseMetaLogger *BaseMetaLogger) Init(inst common.Logger) {
	baseMetaLogger.log = inst
}

// Global instant to use
var Logger = BaseMetaLogger{}