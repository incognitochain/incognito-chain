package btcrelaying

import (
	"github.com/btcsuite/btcutil"
	"testing"
)

func initBTCHeaderChain(t *testing.T) *BlockChain {
	networkName := "test3"
	genesisBlockHeight := int(1746520)

	chainParams, err := setGenesisBlockToChainParams(networkName, genesisBlockHeight)
	if err != nil {
		t.Errorf("Could not set genesis block to chain params with err: %v", err)
		return nil
	}
	dbName := "btc-blocks-test"
	btcChain, err := GetChainV2(dbName, chainParams, int32(genesisBlockHeight))
	if err != nil {
		t.Errorf("Could not get chain instance with err: %v", err)
		return nil
	}
	return btcChain
}

func TestDecodeInvalidBTCAddress(t *testing.T) {
	btcChain := initBTCHeaderChain(t)
	if btcChain == nil {
		t.Error("BTC chain instance should not be null")
		return
	}
	testAddrStr := "tbnb1v63crn5slveu50v8x590uwmqf7kk5xca74scwx"
	params := btcChain.GetChainParams()
	_, err := btcutil.DecodeAddress(testAddrStr, params)
	if err == nil {
		t.Error("Expected returned error is not null, but got null")
	}
}

func TestDecodeValidBTCAddress(t *testing.T) {
	btcChain := initBTCHeaderChain(t)
	if btcChain == nil {
		t.Errorf("BTC chain instance should not be null")
		return
	}
	testAddrStr := "mgLFmRTFRakf5zs23YHB4Pcd8JF7TWCy6E"
	params := btcChain.GetChainParams()
	_, err := btcutil.DecodeAddress(testAddrStr, params)
	if err != nil {
		t.Errorf("Expected returned error is null, but got %v\n", err)
	}
}
