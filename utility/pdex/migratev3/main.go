package main

import (
	"log"
	"time"

	"github.com/incognitochain/incognito-chain/common"
)

const (
	NextShardBlock  = 10 * time.Second
	NextBeaconBlock = 10 * time.Second
)

func main() {
	var err error
	// Load config full node
	fullNodeHost := "http://127.0.0.1:9334"

	// Submit key
	err = submitKey(fullNodeHost)
	if err != nil {
		panic(err)
	}
	log.Println("Finish submitKey")
	time.Sleep(2 * NextShardBlock)

	// Convert coin
	err = convertCoin(fullNodeHost)
	if err != nil {
		panic(err)
	}
	log.Println("Finish convertCoin")
	time.Sleep(2*NextShardBlock + NextBeaconBlock)

	// Init pToken
	err = initToken(fullNodeHost)
	if err != nil {
		panic(err)
	}
	log.Println("Finish initToken")
	time.Sleep(2 * NextShardBlock)

	// Mint nft
	err = mintNft(fullNodeHost)
	if err != nil {
		panic(err)
	}
	log.Println("Finish mintNft")
	time.Sleep(3*NextShardBlock + 2*NextBeaconBlock)

	// Read nftID
	newNftID, err := readNftID(fullNodeHost)
	if err != nil {
		panic(err)
	}
	nftID = newNftID

	// Add pool
	err = addLiquidity(fullNodeHost, true)
	if err != nil {
		panic(err)
	}
	log.Println("Finish add first contribution")
	time.Sleep(2 * NextShardBlock)
	err = addLiquidity(fullNodeHost, false)
	if err != nil {
		panic(err)
	}
	log.Println("Finish add second contribution")
	time.Sleep(2 * NextShardBlock)

	// Add staking pool liquidity
	err = addStakingPoolLiquidity(fullNodeHost, common.PRVCoinID)
	if err != nil {
		panic(err)
	}
	log.Println("Finish add liquidity to prv staking pool")
	time.Sleep(2 * NextShardBlock)
	err = addStakingPoolLiquidity(fullNodeHost, common.PDEXCoinID)
	if err != nil {
		panic(err)
	}
	log.Println("Finish add liquidity to pdex staking pool")
	time.Sleep(2 * NextShardBlock)

	// Modify params
	err = modifyParam(fullNodeHost)
	if err != nil {
		panic(err)
	}
	time.Sleep(3*NextShardBlock + NextBeaconBlock)

	// Trade

	// Add order

	// Withdraw liquidity

	// Unstaking
}
