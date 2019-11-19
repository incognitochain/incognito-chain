package btc

import (
	"time"
)

//Random Client Using Bitcoin
type RandomClient interface {
	//Get Nonce compatible with a given Timestamp
	GetNonceByTimestamp(startTime time.Time, maxTime time.Duration, timestamp int64) (int, int64, int64, error) // return blockHeight, timestamp, nonce, int
	//Verify a given Nonce with a given Timestamp
	VerifyNonceWithTimestamp(startTime time.Time, maxTime time.Duration, timestamp int64, nonce int64) (bool, error)
	// Get timestamp of current block in bitcoin blockchain
	GetCurrentChainTimeStamp() (int64, error)
	// Get timestamp and nonce of a given block height in bitcoin blockchain
	GetTimeStampAndNonceByBlockHeight(blockHeight int) (int64, int64, error)
}
