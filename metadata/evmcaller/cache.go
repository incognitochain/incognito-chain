package evmcaller

import (
	"fmt"
	"time"

	cache "github.com/patrickmn/go-cache"
)

var evmCallerCacher cache.Cache

const CacheLiveTime = 5 * time.Minute

func InitCacher() {
	evmCallerCacher = *cache.New(CacheLiveTime, CacheLiveTime)
}

// define EVM Caller prefix
const evmheaderPrefix = "evm-header"

func getEVMHeaderKey(networkPrefix string, evmBlockHash string) string {
	return fmt.Sprintf("%s-%s-%s", evmheaderPrefix, networkPrefix, evmBlockHash)
}

func handleCacheEVMHeaderResult(networkPrefix string, evmBlockHash string, evmHeaderResult EVMHeaderResult, isReplace bool) error {
	key := getEVMHeaderKey(networkPrefix, evmBlockHash)
	if isReplace {
		return evmCallerCacher.Replace(key, evmHeaderResult, CacheLiveTime)
	}
	return evmCallerCacher.Add(key, evmHeaderResult, CacheLiveTime)
}

func handleGetEVMHeaderResult(networkPrefix string, evmBlockHash string) (*EVMHeaderResult, bool) {
	key := getEVMHeaderKey(networkPrefix, evmBlockHash)
	res, ok := evmCallerCacher.Get(key)
	if !ok {
		return nil, false
	}
	evmHeaderRes := res.(EVMHeaderResult)
	return &evmHeaderRes, true
}
