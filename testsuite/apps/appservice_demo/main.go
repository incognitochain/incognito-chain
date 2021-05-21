package main

import (
	"flag"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	devframework "github.com/incognitochain/incognito-chain/testsuite"
)

func main() {
	fullnode := flag.String("h", "http://139.162.54.236:38934/", "Fullnode Endpoint")
	flag.Parse()

	app := devframework.NewAppService(*fullnode, true)
	cnt := 0
	app.OnBeaconBlock(1215433, func(blk types.BeaconBlock) {
		cnt++
		fmt.Println("produce", blk.GetHeight(), blk.GetProduceTime(), "propose", blk.GetProposeTime())
		//fmt.Printf("%+v", blk.Body.ShardState)
		if cnt == 10 {
			panic(1)
		}
	})

	//app.OnShardBlock(4, 1110926, func(blk types.ShardBlock) {
	//	cnt++
	//	fmt.Println("produce", blk.GetHeight(), blk.GetProduceTime(), "propose", blk.GetProposeTime(), blk.Header.BeaconHeight, blk.Header.Epoch)
	//	if cnt == 10 {
	//		panic(1)
	//	}
	//})
	select {}
}
