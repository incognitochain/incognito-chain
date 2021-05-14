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
	app.OnBeaconBlock(2, func(blk types.BeaconBlock) {
		fmt.Println("blk", blk.GetHeight())
	})

	for j := 0; j < 8; j++ {
		app.OnShardBlock(j, 2, func(blk types.ShardBlock) {
			shardID := blk.GetShardID()
			fmt.Println("blk", shardID, blk.GetHeight())
		})
	}

	select {}
}
