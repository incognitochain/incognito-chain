package metrics

import "github.com/incognitochain/incognito-chain/common"

type MetricLogger struct {
	Log common.Logger
}

func (metricLogger *MetricLogger) Init(inst common.Logger) {
	metricLogger.Log = inst
}

// Global instant to use
var IncLogger = MetricLogger{}
