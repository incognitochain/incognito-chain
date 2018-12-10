package utility

import (
	"fmt"
	"testing"
)

func TestGetNonceByTimestamp(t *testing.T) {
	res, err := GetNonceByTimestamp(0)
	if err != nil {
		t.Errorf("Error geting nonce: %s", err)
	}
	// if res != float64(1447291) {
	// 	t.Errorf("Error geting nonce 2: %s", err)
	// }
	fmt.Println(res)
	//{Name:BTC.main Height:378882 Hash:000000000000000005e5b78413032db5d54a6df5600549f5d3ec339dd7bac4c0 Time:2015-10-14 17:01:57.647896489 +0000 UTC PrevHash:0000000000000000111be2620123bcf2f8ca6209140a8b1f0d8c2412dd9067f2 PeerCount:896 HighFee:52463 MediumFee:27563 LowFee:24926 UnconfirmedCount:81047 LastForkHeight:378316 LastForkHash:00000000000000000806c49f6b53b439beec2a1434f15ae713b84b87a26bbb51}
}

func TestGetNonceByBlock(t *testing.T) {
	// curl https://api.blockcypher.com/v1/btc/test3/blocks/00000000001be2d75acc520630a117874316c07fd7a724afae1a5d99038f4f4a
	blockHash := "00000000001be2d75acc520630a117874316c07fd7a724afae1a5d99038f4f4a"
	res, err := getNonceByBlock(blockHash)
	if err != nil {
		t.Errorf("Error geting nonce: %s", err)
		return
	}
	if res != float64(3733494575) {
		t.Errorf("Error getting nonce in block, nonce should be 3733494575")
	}
}
