package main

import (
	"fmt"

	"github.com/ninjadotorg/constant/utility/btc/btcapi"
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

	// blockHeightList := []string{"1447348", "1447349", "1447350", "1447351"}
	// for _, blockHeight := range blockHeightList {
	// 	nonce, time, err := btcapi.GetNonceOrTimeStampByBlock(blockHeight, false)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// 	fmt.Println("Nonce of blockHeight", blockHeight, " is:", nonce)
	// 	fmt.Println("Nonce of blockHeight", blockHeight, " is:", time)
	// }
	timestamp := int64(1544500800)
	res, err := btcapi.GetNonceByTimestamp(timestamp)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Result of timestamp ", timestamp, "is: ", res)
}
