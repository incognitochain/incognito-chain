package main

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/incdb"
	devframework "github.com/incognitochain/incognito-chain/testsuite"
	"log"
	"os"
	"strings"
	"time"
)

func GetShardBlockByHeight(db incdb.KeyValueReader, sid byte, height uint64) (*types.ShardBlock, error) {
	h, err := rawdbv2.GetFinalizedShardBlockHashByIndex(db, sid, height)
	if err != nil {
		return nil, err
	}
	b, err := rawdbv2.GetShardBlockByHash(db, *h)
	if err != nil {
		return nil, err
	}
	blk := &types.ShardBlock{}
	err = json.Unmarshal(b, blk)
	if err != nil {
		return nil, err
	}
	return blk, nil
}

func GetBeaconBlockByHeight(db incdb.KeyValueReader, height uint64) (*types.BeaconBlock, error) {
	h, err := rawdbv2.GetFinalizedBeaconBlockHashByIndex(db, height)
	if err != nil {
		return nil, err
	}
	b, err := rawdbv2.GetBeaconBlockByHash(db, *h)
	if err != nil {
		return nil, err
	}
	blk := &types.BeaconBlock{}
	err = json.Unmarshal(b, blk)
	if err != nil {
		return nil, err
	}
	return blk, nil
}

func main() {
	db, err := incdb.OpenMultipleDB("leveldb", "/data/mainnet/block")
	if err != nil {
		panic(err)
	}
	node := devframework.NewStandaloneSimulation("regression", devframework.Config{
		ChainParam: devframework.NewChainParam(devframework.ID_MAINNET),
		DataDir:    "/data/regression",
		ResetDB:    false,
		AppNode:    true,
	})

	beaconChain := node.GetBlockchain().BeaconChain
	//beacon insert process
	go func() {
		for {
			time1 := time.Now()
			nextHeight := node.GetBlockchain().GetChain(-1).GetBestView().GetHeight() + 1
			beaconBlock, err := GetBeaconBlockByHeight(db[-1], nextHeight)
			if err != nil {
				fmt.Println("Exit with beacon", nextHeight)
				os.Exit(0)
			}
			shouldWait := false
			for sid, shardStates := range beaconBlock.Body.ShardState {
				if len(shardStates) > 0 && shardStates[len(shardStates)-1].Height > node.GetBlockchain().GetChain(int(sid)).GetFinalView().GetHeight() {
					shouldWait = true
				}
			}
			//log.Printf("Get and Check beacon block %v - %vs\n", nextHeight, time.Since(time1).Seconds())
			time1 = time.Now()
			if !shouldWait {
				err = beaconChain.InsertBlock(beaconBlock, common.REGRESSION_TEST)
				log.Printf("Insert beacon block %v - %vs\n", nextHeight, time.Since(time1).Seconds())
				if err != nil {
					log.Println(err)
					continue
				}
			} else {
				time.Sleep(time.Millisecond * 5)
			}
		}
	}()

	for j := 0; j < 8; j++ {
		go func(shardID int) {
			for {
				time1 := time.Now()
				nextHeight := node.GetBlockchain().GetChain(shardID).GetBestView().GetHeight() + 1
				shardBlock, err := GetShardBlockByHeight(db[shardID], byte(shardID), nextHeight)
				if err != nil {
					fmt.Println("Exit with shard", shardID, nextHeight)
					os.Exit(0)
				}
				shouldWait := false
				beaconFinaView := beaconChain.GetFinalViewHeight()
				if shardBlock.Header.BeaconHeight > beaconChain.GetFinalViewHeight() {
					shouldWait = true
				}
				for sid, cross := range shardBlock.Body.CrossTransactions {
					if cross[len(cross)-1].BlockHeight > node.GetBlockchain().BeaconChain.GetFinalView().(*blockchain.BeaconBestState).BestShardHeight[sid] {
						shouldWait = true
					} else {
						for _, blk := range cross {
							fmt.Println("debug create crossshard block", int(sid), int(shardID), blk.BlockHeight, blk.BlockHash.String())
							crossBlk, err := GetShardBlockByHeight(db[int(sid)], sid, blk.BlockHeight)
							if err != nil {
								panic(err)
							}
							crossX, err := blockchain.CreateCrossShardBlock(crossBlk, byte(shardID))
							if err != nil {
								panic(err)
							}
							fmt.Println("debug insert cross shard block", int(sid), int(shardID), blk.BlockHeight, crossX.Hash().String())

							node.GetSyncker().InsertCrossShardBlock(crossX)
						}

					}
				}

				//log.Printf("Get and Check shard %v block %v - %vs\n", shardID, nextHeight, time.Since(time1).Seconds())
				time1 = time.Now()
				if !shouldWait {

					err = node.GetBlockchain().GetChain(shardID).(*blockchain.ShardChain).InsertBlock(shardBlock, common.REGRESSION_TEST)
					log.Printf("Insert shard %v block %v - %vs\n", shardID, nextHeight, time.Since(time1).Seconds())
					if err != nil {
						if strings.Index(err.Error(), "Fetch Beacon Blocks Error") == -1 {
							log.Println(err)
							panic(1)
						} else {
							log.Println("Wait for beacon", shardBlock.Header.BeaconHeight, beaconFinaView)
						}
						continue
					}
				} else {
					time.Sleep(time.Millisecond * 5)
				}
			}
		}(j)
	}
	select {}
}
