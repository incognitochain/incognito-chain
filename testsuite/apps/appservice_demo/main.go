package main

import (
	"flag"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/config"
	devframework "github.com/incognitochain/incognito-chain/testsuite"
)

func main() {
	fullnode := flag.String("h", "http://51.222.153.212:9334/", "Fullnode Endpoint")
	flag.Parse()
	config.LoadConfig()
	config.LoadParam()
	app := devframework.NewAppService(*fullnode, true)
	//app.OnBeaconBlock(8664, func(blk types.BeaconBlock) {
	//	for sid, states := range blk.Body.ShardState {
	//		fmt.Println("Shard ", sid)
	//		for _, s := range states {
	//			fmt.Println(s.Height, s.Hash.String())
	//			fmt.Println(s.ValidationData, s.PreviousValidationData)
	//		}
	//	}
	//})

	app.OnShardBlock(4, 2313686, func(blk types.ShardBlock) {
		fmt.Println(blk.Body.Transactions)
		for _, tx := range blk.Body.Transactions {

		}
	})

	//for j := 0; j < 8; j++ {
	//	app.OnShardBlock(j, 2, func(blk types.ShardBlock) {
	//		shardID := blk.GetShardID()
	//		fmt.Println("blk", shardID, blk.GetHeight())
	//	})
	//}

	select {}
}
