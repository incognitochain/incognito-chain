package main

import (
	"flag"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/config"
	devframework "github.com/incognitochain/incognito-chain/testsuite"
)

func main() {
	fullnode := flag.String("h", "http://127.0.0.1:20000/", "Fullnode Endpoint")
	flag.Parse()
	config.LoadConfig()
	config.LoadParam()
	app := devframework.NewAppService(*fullnode, true)
	app.OnBeaconBlock(1, func(blk types.BeaconBlock) {
		//if blk.Header.ProcessBridgeFromBlock != nil {
		//	fmt.Println("blk", blk.GetVersion(), blk.GetHeight(), blk.GetFinalityHeight(), *blk.Header.ProcessBridgeFromBlock)
		//} else {
		//	fmt.Println("blk", blk.GetVersion(), blk.GetHeight(), blk.GetFinalityHeight())
		//}
		fmt.Println(blk.GetVersion(), blk.Body.Instructions)
	})

	//app.OnShardBlock(0, 1, func(blk types.ShardBlock) {
	//	fmt.Println("blk", blk.GetHeight(), blk.GetFinalityHeight(), blk.GetProposeTime(), blk.GetProduceTime())
	//})

	//for j := 0; j < 8; j++ {
	//	app.OnShardBlock(j, 2, func(blk types.ShardBlock) {
	//		shardID := blk.GetShardID()
	//		fmt.Println("blk", shardID, blk.GetHeight())
	//	})
	//}

	select {}
}
