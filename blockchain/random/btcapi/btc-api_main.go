//+build !test

package main

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/btc"
)

func main() {
	//res, err := btcapi.GetCurrentChainTimeStamp()
	//fmt.Printf("res %+v \n error %+v", res, err)
	var btcClient = btc.NewBTCClient("admin", "autonomous", "159.65.142.153", "8332")
	res, err := btcClient.GetBlockchainInfo()
	fmt.Println(res, err)
	blockHeight, err := btcClient.GetBestBlockHeight()
	fmt.Println(blockHeight, err)
	chainHeight, timeStamp, nonce, err := btcClient.GetChainTimeStampAndNonce()
	fmt.Println(chainHeight, timeStamp, nonce, err)
	fmt.Println(btcClient.GetTimeStampAndNonceByBlockHeight(579358))
}
