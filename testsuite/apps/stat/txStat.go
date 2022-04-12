package main

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/portal"
	devframework "github.com/incognitochain/incognito-chain/testsuite"
	"time"
)

const MONGODB = "mongodb://51.91.72.45:38118"
const INCOGNITO_NODE = "http://135.125.97.141:9339/"

func main() {
	config.LoadConfig()
	config.LoadParam()
	portal.SetupParam()

	statDB, err := NewStatDB(MONGODB, "netmonitor", "stattx")
	if err != nil {
		panic(err)
	}
	node := devframework.NewAppService(INCOGNITO_NODE, true)

	for i := 0; i < config.Param().ActiveShards; i++ {
		fromBlock := statDB.lastBlock(i)
		node.OnShardBlock(i, uint64(int64(fromBlock)), func(shardBlk types.ShardBlock) {
			shardID := shardBlk.GetShardID()
			fmt.Println("shardBlk", shardID, shardBlk.GetHeight())
			for _, tx := range shardBlk.Body.Transactions {
				txHash := tx.Hash().String()
				inputCoin := 0
				outCoin := 0
				if tx.GetProof() != nil {
					inputCoin = len(tx.GetProof().GetInputCoins())
					outCoin = len(tx.GetProof().GetOutputCoins())
				}

				txType := tx.GetType()
				metaType := tx.GetMetadataType()
				statDB.set(StatInfo{
					ChainID:     shardID,
					BlockHeight: int(shardBlk.GetHeight()),
					BlockHash:   shardBlk.Hash().String(),
					BlockTime:   time.Unix(shardBlk.GetProduceTime(), 0),
					TxHash:      txHash,
					InputCoin:   inputCoin,
					OutputCoin:  outCoin,
					Type:        txType,
					MetaType:    metaType,
				})

			}

		})
	}
	select {}
}
