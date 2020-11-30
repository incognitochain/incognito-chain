package appservices


import "github.com/incognitochain/incognito-chain/common"

type AppServiceLogger struct {
	log common.Logger
}

func (appServiceLogger *AppServiceLogger) Init(inst common.Logger) {
	appServiceLogger.log = inst
}

// Global instant to use
var Logger = AppServiceLogger{}