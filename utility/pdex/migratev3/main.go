package main

import (
	"log"
	"time"
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
	time.Sleep(2 * NextShardBlock)

	// Add pool

	// Add staking pool liquidity

	// Modify params

	// Trade

	// Add order

	// Withdraw liquidity

}
