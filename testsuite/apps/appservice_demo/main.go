package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/config"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
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

	app.OnShardBlock(4, 2313570, func(blk types.ShardBlock) {
		fmt.Println(blk.Body.Transactions)
		for _, tx := range blk.Body.Transactions {
			if tx.GetMetadataType() == metadataCommon.Pdexv3AddOrderRequestMeta {
				md := tx.GetMetadata()
				req, ok := md.(*metadataPdexv3.AddOrderRequest)
				if !ok {
					panic(100)
				}
				tokenToSell := req.TokenToSell.String()
				pair := req.PoolPairID
				strs := strings.Split(pair, "-")
				if tokenToSell != strs[0] && tokenToSell != strs[1] {
					fmt.Println(tx.Hash().String())
				}
			}
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
