package btcrelaying

import (
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
)

func setGenesisBlockToChainParamsByNetwork(
	networkName string,
	genesisBlkHeight int,
	chainParams *chaincfg.Params,
) (*chaincfg.Params, error) {
	blk, err := buildBTCBlockFromCypher(networkName, genesisBlkHeight)
	if err != nil {
		return nil, err
	}

	// chainParams := chaincfg.MainNetParams
	// chainParams := chaincfg.TestNet3Params
	chainParams.GenesisBlock = blk.MsgBlock()
	chainParams.GenesisHash = blk.Hash()
	return chainParams, nil
}

func initBTCHeaderTestNetChain(t *testing.T) *BlockChain {
	networkName := "test3"
	genesisBlockHeight := int(1746520)

	chainParams, err := setGenesisBlockToChainParamsByNetwork(networkName, genesisBlockHeight, &chaincfg.TestNet3Params)
	if err != nil {
		t.Errorf("Could not set genesis block to chain params with err: %v", err)
		return nil
	}
	dbName := "btc-blocks-testnet"
	btcChain, err := GetChainV2(dbName, chainParams, int32(genesisBlockHeight))
	if err != nil {
		t.Errorf("Could not get chain instance with err: %v", err)
		return nil
	}
	return btcChain
}

func initBTCHeaderMainNetChain(t *testing.T) *BlockChain {
	networkName := "main"
	genesisBlockHeight := int(632061)

	chainParams, err := setGenesisBlockToChainParamsByNetwork(networkName, genesisBlockHeight, &chaincfg.MainNetParams)
	if err != nil {
		t.Errorf("Could not set genesis block to chain params with err: %v", err)
		return nil
	}
	dbName := "btc-blocks-mainnet"
	btcChain, err := GetChainV2(dbName, chainParams, int32(genesisBlockHeight))
	if err != nil {
		t.Errorf("Could not get chain instance with err: %v", err)
		return nil
	}
	return btcChain
}

func TestDecodeInvalidBTCTestNetAddress(t *testing.T) {
	btcChain := initBTCHeaderTestNetChain(t)
	if btcChain == nil {
		t.Error("BTC chain instance should not be null")
		return
	}
	// an address on mainnet
	testAddrStr := "1A7tWftaGHohhGcJMVkkm4zAYnF53KjRnU"
	params := btcChain.GetChainParams()
	_, err := btcutil.DecodeAddress(testAddrStr, params)
	if err == nil {
		t.Error("Expected returned error is not null, but got null")
	}
}

func TestDecodeValidBTCTestNetAddress(t *testing.T) {
	btcChain := initBTCHeaderTestNetChain(t)
	if btcChain == nil {
		t.Errorf("BTC chain instance should not be null")
		return
	}
	// an address on testnet
	testAddrStr := "mgLFmRTFRakf5zs23YHB4Pcd8JF7TWCy6E"
	params := btcChain.GetChainParams()
	_, err := btcutil.DecodeAddress(testAddrStr, params)
	if err != nil {
		t.Errorf("Expected returned error is null, but got %v\n", err)
	}
}

func TestDecodeInvalidBTCMainNetAddress(t *testing.T) {
	btcChain := initBTCHeaderMainNetChain(t)
	if btcChain == nil {
		t.Error("BTC chain instance should not be null")
		return
	}
	// an address on testnet
	testAddrStr := "mgLFmRTFRakf5zs23YHB4Pcd8JF7TWCy6E"
	params := btcChain.GetChainParams()
	_, err := btcutil.DecodeAddress(testAddrStr, params)
	if err == nil {
		t.Error("Expected returned error is not null, but got null")
	}
}

func TestDecodeValidBTCMainNetAddress(t *testing.T) {
	btcChain := initBTCHeaderMainNetChain(t)
	if btcChain == nil {
		t.Error("BTC chain instance should not be null")
		return
	}
	// an address on mainnet
	testAddrStr := "bc1qq7ndvtvyzcea44ps6d4nt3plk02ghpsha0t55y"
	params := btcChain.GetChainParams()
	_, err := btcutil.DecodeAddress(testAddrStr, params)
	if err != nil {
		t.Errorf("Expected returned error is null, but got %v\n", err)
	}
}

