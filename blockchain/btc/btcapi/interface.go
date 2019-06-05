package btcapi

type RandomClient interface {
	GetNonceByTimestamp(timestamp int64) (int, int64, int64, error)
	VerifyNonceWithTimestamp(timestamp int64, nonce int64) (bool, error)
	GetCurrentChainTimeStamp() (int64, error)
	GetNonceOrTimeStampByBlock(blockHeight string, nonceOrTime bool) (int64, int64, error)
}
