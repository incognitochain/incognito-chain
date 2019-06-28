package btc

import (
	"testing"
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
	var btcClient = NewBTCClient("admin", "autonomous", "159.65.142.153", "8332")
	blockHeight, timestamp, nonce, err := btcClient.GetNonceByTimestamp(1373297940)
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
	var btcClient = NewBTCClient("admin", "autonomous", "159.65.142.153", "8332")
	isOk, err := btcClient.VerifyNonceWithTimestamp(1373297940, 3029573794)
	if err != nil {
		t.Error("Fail to get chain timestamp and nonce")
	}
	if !isOk {
		t.Error("Fail to verify nonce by timestamp")
	}
}
