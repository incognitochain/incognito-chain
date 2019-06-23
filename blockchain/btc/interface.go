package btc

type RandomClient interface {
	GetNonceByTimestamp(timestamp int64) (int, int64, int64, error) // return blockHeight, timestamp, nonce, int
	VerifyNonceWithTimestamp(timestamp int64, nonce int64) (bool, error)
	GetCurrentChainTimeStamp() (int64, error)
	GetTimeStampAndNonceByBlockHeight(blockHeight int) (int64, int64, error)
}
