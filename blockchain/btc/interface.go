package btc

//Random Client Using Bitcoin
type RandomClient interface {
	//Get Nonce compatible with a given Timestamp
	GetNonceByTimestamp(timestamp int64) (int, int64, int64, error) // return blockHeight, timestamp, nonce, int
	//Verify a given Nonce with a given Timestamp
	VerifyNonceWithTimestamp(timestamp int64, nonce int64) (bool, error)
	// Get timestamp of current block in bitcoin blockchain
	GetCurrentChainTimeStamp() (int64, error)
	// Get timestamp and nonce of a given block height in bitcoin blockchain
	GetTimeStampAndNonceByBlockHeight(blockHeight int) (int64, int64, error)
}
