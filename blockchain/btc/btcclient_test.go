package btc

import (
	"strconv"
	"testing"
	"time"
)

var (
	duration = 10 * time.Second
)

func TestBlockHashByHeight(t *testing.T) {
	var btcClient = NewBTCClient("admin", "autonomous", "159.65.142.153", "8332")
	hash, err := btcClient.GetBlockHashByHeight(2)
	if err != nil {
		t.Error("Failed to get hash by height")
	}
	if hash != "000000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd" {
		t.Error("Wrong Block Hash")
	}
}
func TestGetTimestampAndNonceByBlockHeight(t *testing.T) {
	var btcClient = NewBTCClient("admin", "autonomous", "159.65.142.153", "8332")
	timestamp, nonce, err := btcClient.GetTimeStampAndNonceByBlockHeight(2)
	if err != nil {
		t.Error("Fail to get timestamp and nonce")
	}
	if timestamp != 1231469744 {
		t.Error("Wrong Timestamp")
	}
	if nonce != 1639830024 {
		t.Error("Wrong Nonce")
	}
}
func TestGetTimestampAndNonceByBlockHash(t *testing.T) {
	var btcClient = NewBTCClient("admin", "autonomous", "159.65.142.153", "8332")
	timestamp, nonce, err := btcClient.GetTimeStampAndNonceByBlockHash("000000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd")
	if err != nil {
		t.Error("Fail to get timestamp and nonce")
	}
	if timestamp != 1231469744 {
		t.Error("Wrong Timestamp")
	}
	if nonce != 1639830024 {
		t.Error("Wrong Nonce")
	}
}
func TestGetChainTimeStampAndNonce(t *testing.T) {
	var btcClient = NewBTCClient("admin", "autonomous", "159.65.142.153", "8332")
	_, err := btcClient.GetCurrentChainTimeStamp()
	if err != nil {
		t.Error("Fail to get chain timestamp and nonce")
	}
}
func TestGetNonceByTimeStamp(t *testing.T) {
	startTime := time.Now()
	var btcClient = NewBTCClient("admin", "autonomous", "159.65.142.153", "8332")
	blockHeight, timestamp, nonce, err := btcClient.GetNonceByTimestamp(startTime, duration, 1373297940)
	t.Log(blockHeight, timestamp, nonce)
	if err != nil {
		t.Error("Fail to get chain timestamp and nonce")
	}
	if blockHeight != 245502 {
		t.Error("Wrong Block")
	}
	if timestamp != int64(1373298838) {
		t.Error("Wrong Timestamp")
	}
	if nonce != int64(3029573794) {
		t.Error("Wrong Nonce")
	}
}
func TestVerifyNonceByTimeStamp(t *testing.T) {
	startTime := time.Now()
	var btcClient = NewBTCClient("admin", "autonomous", "159.65.142.153", "8332")
	suite := []struct {
		timestamp int64
		nonce     int64
	}{
		{timestamp: 1373297940, nonce: 3029573794},
		{timestamp: 1569472740, nonce: 2208937036},
		{timestamp: 1572924720, nonce: 73996767},
		{timestamp: 1367725860, nonce: 3580525302},
		{timestamp: 1432484400, nonce: 2456900135},
	}
	for index, testSuite := range suite {
		t.Run(strconv.Itoa(index+1), func(t *testing.T) {
			isOk, err := btcClient.VerifyNonceWithTimestamp(startTime, duration, testSuite.timestamp, testSuite.nonce)
			if err != nil {
				t.Fatal("Fail to get chain timestamp and nonce")
			}
			if !isOk {
				t.Fatal("Fail to verify nonce by timestamp")
			}
		})
	}
}
