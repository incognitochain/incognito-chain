package consensus

import "github.com/incognitochain/incognito-chain/common"

type consensusLogger struct {
	log common.Logger
}

func (consensusLogger *consensusLogger) Init(inst common.Logger) {
	consensusLogger.log = inst
}

// Global instant to use
var Logger = consensusLogger{}
