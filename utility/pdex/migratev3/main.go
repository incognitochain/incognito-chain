package main

import (
	"log"
	"os"
	"time"

	"github.com/incognitochain/incognito-chain/common"
)

const (
	NextShardBlock  = 10 * time.Second
	NextBeaconBlock = 10 * time.Second
)

func main() {
	args := os.Args
	fullNodeHost := args[1]
	var err error
	// Load config full node

	// Submit key
	err = submitKey(fullNodeHost)
	if err != nil {
		log.Println(err)
	}
	log.Println("Finish submitKey")
	time.Sleep(3 * NextShardBlock)

	// Convert coin
	err = convertCoin(fullNodeHost)
	if err != nil {
		log.Println(err)
	}
	log.Println("Finish convertCoin")
	time.Sleep(3*NextShardBlock + 2*NextBeaconBlock)

	// Init pToken
	err = initToken(fullNodeHost)
	if err != nil {
		panic(err)
	}
	log.Println("Finish initToken")
	time.Sleep(3 * NextShardBlock)

	// Mint nft
	err = mintNft(fullNodeHost)
	if err != nil {
		panic(err)
	}
	log.Println("Finish mintNft")
	time.Sleep(4*NextShardBlock + 2*NextBeaconBlock)

	// Read state
	err = readState(fullNodeHost)
	if err != nil {
		panic(err)
	}

	// Add pool
	err = addLiquidity(fullNodeHost, true)
	if err != nil {
		panic(err)
	}
	log.Println("Finish add first contribution")
	time.Sleep(3 * NextShardBlock)
	err = addLiquidity(fullNodeHost, false)
	if err != nil {
		panic(err)
	}
	log.Println("Finish add second contribution")
	time.Sleep(3 * NextShardBlock)

	// Add staking pool liquidity
	err = staking(fullNodeHost, common.PRVCoinID)
	if err != nil {
		panic(err)
	}
	log.Println("Finish add liquidity to prv staking pool")
	time.Sleep(3 * NextShardBlock)
	err = staking(fullNodeHost, common.PDEXCoinID)
	if err != nil {
		panic(err)
	}
	log.Println("Finish add liquidity to pdex staking pool")
	time.Sleep(3 * NextShardBlock)

	// Read state
	err = readState(fullNodeHost)
	if err != nil {
		panic(err)
	}

	// Modify params
	err = modifyParam(fullNodeHost)
	if err != nil {
		panic(err)
	}
	time.Sleep(4*NextShardBlock + 2*NextBeaconBlock)

	// Trade
	err = trade(fullNodeHost, true)
	if err != nil {
		panic(err)
	}
	time.Sleep(3 * NextShardBlock)

	// Add order
	err = addOrder(fullNodeHost)
	if err != nil {
		panic(err)
	}
	time.Sleep(3 * NextShardBlock)

	// Trade second time
	err = trade(fullNodeHost, false)
	if err != nil {
		panic(err)
	}
	time.Sleep(4*NextShardBlock + 2*NextBeaconBlock)

	// Unstaking
	err = unstaking(fullNodeHost, common.PRVIDStr)
	if err != nil {
		panic(err)
	}
	time.Sleep(4*NextShardBlock + 3*NextBeaconBlock)
	err = unstaking(fullNodeHost, common.PDEXIDStr)
	if err != nil {
		panic(err)
	}
	time.Sleep(4*NextShardBlock + 3*NextBeaconBlock)

	// Withdraw liquidity
	err = withdrawLiquidity(fullNodeHost)
	if err != nil {
		panic(err)
	}
	time.Sleep(3*NextShardBlock + 2*NextBeaconBlock)
	log.Println("Done!!!")
}
