package connmanager

import (
	"github.com/big0t/constant-chain/common"
)

type ConnManagerLogger struct {
	log common.Logger
}

func (connManagerLogger *ConnManagerLogger) Init(inst common.Logger) {
	connManagerLogger.log = inst
}

// Global instant to use
var Logger = ConnManagerLogger{}
