package trie

import (
	"github.com/incognitochain/incognito-chain/common"
)

type TLogger struct {
	log common.Logger
}

func (tLogger *TLogger) Init(inst common.Logger) {
	tLogger.log = inst
}

// Global instant to use
var Logger = TLogger{}
