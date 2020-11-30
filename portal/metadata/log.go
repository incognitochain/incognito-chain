package metadata

import "github.com/incognitochain/incognito-chain/common"

type PortalMetaDataLogger struct {
	log common.Logger
}

func (metricLogger *PortalMetaDataLogger) Init(inst common.Logger) {
	metricLogger.log = inst
}

// Global instant to use
var Logger = PortalMetaDataLogger{}
