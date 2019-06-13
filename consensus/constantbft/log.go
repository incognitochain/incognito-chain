package constantbft

import "github.com/incognitochain/incognito-chain/common"

type constantbftLogger struct {
	log common.Logger
}

func (constantbftLogger *constantbftLogger) Init(inst common.Logger) {
	constantbftLogger.log = inst
}

// Global instant to use
var Logger = constantbftLogger{}
