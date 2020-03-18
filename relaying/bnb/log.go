package bnb

import "github.com/incognitochain/incognito-chain/common"

type RelayingLogger struct {
	log common.Logger
}

func (logger *RelayingLogger) Init(inst common.Logger) {
	logger.log = inst
}

// Global instant to use
var Logger = RelayingLogger{}
