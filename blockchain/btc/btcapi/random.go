package btcapi

func GenerateRandomNumber(timestamp int64, ch chan<- int64) {
	for {
		select {
		default:
			res, err := GetNonceByTimestamp(timestamp)
			if err == nil {
				ch <- res
				return
			}
		}
	}
}
