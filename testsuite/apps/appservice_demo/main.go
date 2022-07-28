package main

import (
	"flag"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	devframework "github.com/incognitochain/incognito-chain/testsuite"
)

func main() {
	fullnode := flag.String("h", "http://51.91.220.58:9335/", "Fullnode Endpoint")
	flag.Parse()

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

	app.OnShardBlock(0, 8650, func(blk types.ShardBlock) {
		shardID := blk.GetShardID()
		fmt.Println("blk", blk.GetHeight(), shardID, blk.GetVersion())
	})

	//for j := 0; j < 8; j++ {
	//	app.OnShardBlock(j, 2, func(blk types.ShardBlock) {
	//		shardID := blk.GetShardID()
	//		fmt.Println("blk", shardID, blk.GetHeight())
	//	})
	//}

	select {}
}
