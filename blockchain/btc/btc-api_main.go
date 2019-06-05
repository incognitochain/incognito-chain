package main

import (
	"fmt"
	"github.com/constant-money/constant-chain/blockchain/btc/btcapi"
)

func main() {
	//res, err := btcapi.GetCurrentChainTimeStamp()
	//fmt.Printf("res %+v \n error %+v", res, err)
	var btcClient = btcapi.NewBTCClient("admin","autonomous", "159.65.142.153","8332")
	res, err := btcClient.GetBlockchainInfo()
	fmt.Println(res,err)
}
