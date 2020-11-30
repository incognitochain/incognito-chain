package storage

import "github.com/incognitochain/incognito-chain/common"

type StorageLogger struct {
	log common.Logger
}

func (storageLogger *StorageLogger) Init(inst common.Logger) {
	storageLogger.log = inst
}

// Global instant to use
var Logger = StorageLogger{}
