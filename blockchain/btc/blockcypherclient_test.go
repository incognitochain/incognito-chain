//+build !test

package btc

import (
	"testing"
)

func TestGetChainTimeStampAndNonceBlockCypher(t *testing.T) {
	var btcClient = BlockCypherClient{}
	_, err := btcClient.GetCurrentChainTimeStamp()
	if err != nil {
		t.Error("Fail to get chain timestamp and nonce")
	}
}
func TestGetTimestampAndNonceByBlockHeightBlockCypher(t *testing.T) {
	var btcClient = BlockCypherClient{}
	timestamp, nonce, err := btcClient.GetTimeStampAndNonceByBlockHeight(2)
	t.Log(timestamp, nonce)
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

func TestGetNonceByTimeStampBlockCypher(t *testing.T) {
	var btcClient = BlockCypherClient{}
	blockHeight, timestamp, nonce, err := btcClient.GetNonceByTimestamp(1373297940)
	t.Log(blockHeight, timestamp, nonce)
	if err != nil {
		t.Error("Fail to get chain timestamp and nonce", err)
		t.Fatal()
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
func TestVerifyNonceByTimeStampBlockCypher(t *testing.T) {
	var btcClient = BlockCypherClient{}
	isOk, err := btcClient.VerifyNonceWithTimestamp(1373297940, 3029573794)
	if err != nil {
		t.Error("Fail to get chain timestamp and nonce")
		t.FailNow()
	}
	if !isOk {
		t.Error("Fail to verify nonce by timestamp")
	}
}
