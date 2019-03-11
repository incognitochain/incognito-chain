package mempool

import "github.com/big0t/constant-chain/common"

type MempoolLogger struct {
	log common.Logger
}

func (mempoolLogger *MempoolLogger) Init(inst common.Logger) {
	mempoolLogger.log = inst
}

// Global instant to use
var Logger = MempoolLogger{}
