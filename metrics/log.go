package metrics

import "github.com/incognitochain/incognito-chain/common"

type MetricLogger struct {
	log common.Logger
}

func (metricLogger *MetricLogger) Init(inst common.Logger) {
	metricLogger.log = inst
}

// Global instant to use
var Logger = MetricLogger{}
