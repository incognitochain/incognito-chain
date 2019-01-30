package btcapi

import "strconv"

func GenerateRandomNumber(timestamp int64, ch chan<- string) {
	for {
		select {
		default:
			height, timestamp, nonce, err := GetNonceByTimestamp(timestamp)
			if err == nil {
				ch <- strconv.Itoa(height) + "," + strconv.Itoa(int(timestamp)) + "," + strconv.Itoa(int(nonce))
				return
			}
		}
	}
}
