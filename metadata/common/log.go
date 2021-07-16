package common

import "github.com/incognitochain/incognito-chain/common"

type MetaDataLogger struct {
	Log common.Logger
}

func (metadataLogger *MetaDataLogger) Init(inst common.Logger) {
	metadataLogger.Log = inst
}

// Global instant to use
var Logger = MetaDataLogger{}
