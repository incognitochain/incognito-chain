package data

import "github.com/incognitochain/incognito-chain/common"

type DataLogger struct {
	log common.Logger
}

func (dataLogger *DataLogger) Init(inst common.Logger) {
	dataLogger.log = inst
}

// Global instant to use
var Logger = DataLogger{}

