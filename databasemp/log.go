package databasemp



import (
	"github.com/constant-money/constant-chain/common"
)

type DBMemmpoolLogger struct {
	log common.Logger
}

func (dbLogger *DBMemmpoolLogger) Init(inst common.Logger) {
	dbLogger.log = inst
}

// Global instant to use
var Logger = DBMemmpoolLogger{}
