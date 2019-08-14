package metadata

import "github.com/incognitochain/incognito-chain/common"

type MetaDataLogger struct {
	log common.Logger
}

func (metricLogger *MetaDataLogger) Init(inst common.Logger) {
	metricLogger.log = inst
}

// Global instant to use
var Logger = MetaDataLogger{}
