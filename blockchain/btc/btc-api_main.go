package main

import (
	"fmt"

	"github.com/ninjadotorg/constant/blockchain/btc/btcapi"
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
	timestamp := int64(1544500800)
	// res, err := btcapi.GetNonceByTimestamp(timestamp)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println("Result of timestamp ", timestamp, "is: ", res)

	msg := make(chan int64)

	go btcapi.GenerateRandomNumber(timestamp, msg)
	res := <-msg
	fmt.Println("Result of timestamp ", timestamp, "is: ", res)
}
