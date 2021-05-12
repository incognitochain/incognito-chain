package config

import "github.com/incognitochain/incognito-chain/common"

type ConfigLogger struct {
	log common.Logger
}

func (configLogger *ConfigLogger) Init(inst common.Logger) {
	configLogger.log = inst
}

// Global instant to use
var Logger = ConfigLogger{}
