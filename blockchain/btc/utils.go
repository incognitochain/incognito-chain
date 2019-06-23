package btc

import (
	"errors"
	"math"
	"time"
)

// count in second
// use t.UnixNano() / int64(time.Millisecond) for milisecond
func makeTimestamp(t time.Time) int64 {
	return t.Unix()
}

// convert time.RFC3339 -> int64 value
// t,_ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
func makeTimestamp2(t string) (int64, error) {
	res, err := time.Parse(time.RFC3339, t)
	if err != nil {
		return -1, err
	}
	return makeTimestamp(res), nil
}

// assume that each block will be produced in 10 mins ~= 600s
// this function will based on the given #param1 timestamp and #param3 chainTimestamp
// to calculate blockheight with approximate timestamp with #param1
// blockHeight = chainHeight - (chainTimestamp - timestamp) / 600
func estimateBlockHeight(self RandomClient, timestamp int64, chainHeight int, chainTimestamp int64) (int, error) {
	var estimateBlockHeight int
	// fmt.Printf("EstimateBlockHeight timestamp %d, chainHeight %d, chainTimestamp %d\n", timestamp, chainHeight, chainTimestamp)
	offsetSeconds := timestamp - chainTimestamp
	if offsetSeconds > 0 {
		return chainHeight, nil
	} else {
		estimateBlockHeight = chainHeight
		// diff is negative
		for true {
			diff := int(offsetSeconds / 600)
			estimateBlockHeight = estimateBlockHeight + diff
			//fmt.Printf("Estimate blockHeight %d \n", estimateBlockHeight)
			if math.Abs(float64(diff)) < 3 {
				return estimateBlockHeight, nil
			}
			blockTimestamp, _, err := self.GetTimeStampAndNonceByBlockHeight(estimateBlockHeight)
			// fmt.Printf("Try to estimate block with timestamp %d \n", blockTimestamp)
			if err != nil {
				return -1, err
			}
			if blockTimestamp == MAX_TIMESTAMP {
				return -1, NewBTCAPIError(APIError, errors.New("Can't get result from API"))
			}
			offsetSeconds = timestamp - blockTimestamp
		}
	}
	return chainHeight, NewBTCAPIError(UnExpectedError, errors.New("Can't estimate block height"))
}

