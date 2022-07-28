package mock

import (
	"time"
)

type BTCRandom struct{}

func (btc *BTCRandom) GetNonceByTimestamp(startTime time.Time, maxTime time.Duration, timestamp int64) (int, int64, int64, error) {
	return 0, 0, 0, nil
}
func (btc *BTCRandom) VerifyNonceWithTimestamp(startTime time.Time, maxTime time.Duration, timestamp int64, nonce int64) (bool, error) {
	return true, nil
}
func (btc *BTCRandom) GetCurrentChainTimeStamp() (int64, error) {
	return 0, nil
}
func (btc *BTCRandom) GetTimeStampAndNonceByBlockHeight(blockHeight int) (int64, int64, error) {
	return 0, 0, nil
}
