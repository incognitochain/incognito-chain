package main

import (
	"fmt"

	"github.com/big0t/constant-chain/blockchain/btc/btcapi"
)

func main() {
	// blockHeight := "1447349"
	// flag := true
	// nonce, time, err := btcapi.GetNonceOrTimeStampByBlock(blockHeight, flag)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println("Nonce:", nonce)
	// fmt.Println("Time:", time)

	//
	// blockHeightList := []string{"1447348", "1447349", "1447350", "1447351"}
	// blockHeightList := []string{"577265", "577266", "577267", "577268"}
	// for _, blockHeight := range blockHeightList {
	// 	// false for timestamp
	// 	// true for nonce
	// 	nonce, time, err := btcapi.GetNonceOrTimeStampByBlock(blockHeight, false)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// 	fmt.Println("Nonce of blockHeight", blockHeight, " is:", nonce)
	// 	fmt.Println("Timestamp of blockHeight", blockHeight, " is:", time)
	// }

	//1444500800, 1544500800
	// timestamp := int64(1547449828)
	// height, blockTimestamp, nonce, err := btcapi.GetNonceByTimestamp(timestamp)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Printf("Result of height %+v, blockTimestamp %+v, timestamp %+v, nonce %+v \n", height, blockTimestamp, timestamp, nonce)
	res, err := btcapi.GetCurrentChainTimeStamp()
	fmt.Printf("res %+v \n error %+v", res, err)
}
