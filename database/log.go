package database

import (
	"github.com/constant-money/constant-chain/common"
)

type DbLogger struct {
	log common.Logger
}

func (dbLogger *DbLogger) Init(inst common.Logger) {
	dbLogger.log = inst
}

// Global instant to use
var Logger = DbLogger{}
