package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/incdb"
	devframework "github.com/incognitochain/incognito-chain/testsuite"
	"log"
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

func streamblock(endpoint string, node *devframework.NodeEngine) (chan types.BeaconBlock, map[int]chan types.ShardBlock) {
	beaconCh := make(chan types.BeaconBlock, 500)
	shardCh := make(map[int]chan types.ShardBlock)

	fullnodeRPC := devframework.RemoteRPCClient{endpoint}

	go func() {
		fromBlk := node.GetBlockchain().BeaconChain.GetBestViewHeight() + 1

		for {

			data, err := fullnodeRPC.GetBlocksFromHeight(-1, uint64(fromBlk), 50)
			//fmt.Println("len", len(beaconCh))
			if err != nil || len(data.([]types.BeaconBlock)) == 0 {
				fmt.Println(err)
				time.Sleep(time.Minute)
				continue
			}
			for _, blk := range data.([]types.BeaconBlock) {
				beaconCh <- blk
				fromBlk = blk.GetHeight() + 1
				if fromBlk%10000 == 0 {

				}
			}
		}
	}()

	for i := 0; i < 8; i++ {
		shardCh[i] = make(chan types.ShardBlock, 500)
		go func(sid int) {
			fromBlk := node.GetBlockchain().GetChain(sid).GetBestView().GetHeight() + 1
			for {
				data, err := fullnodeRPC.GetBlocksFromHeight(sid, uint64(fromBlk), 50)
				//fmt.Println("len", len(shardCh[sid]))
				if err != nil || len(data.([]types.ShardBlock)) == 0 {
					fmt.Println(err)
					time.Sleep(time.Minute)
					continue
				}
				for _, blk := range data.([]types.ShardBlock) {
					shardCh[sid] <- blk
					fromBlk = blk.GetHeight() + 1
				}
			}
		}(i)
	}

	return beaconCh, shardCh

}

func main() {
	dir := flag.String("d", "./regression", "Datadir")
	fullnode := flag.String("h", "http://139.162.54.236:38934/", "Fullnode Endpoint")
	flag.Parse()

	node := devframework.NewStandaloneSimulation("regression", devframework.Config{
		ChainParam: devframework.NewChainParam(devframework.ID_MAINNET),
		DataDir:    *dir,
		ResetDB:    false,
		AppNode:    true,
	})
	beaconCh, shardCh := streamblock(*fullnode, node)

	beaconChain := node.GetBlockchain().BeaconChain
	//beacon insert process
	go func() {
		for {
			time1 := time.Now()
			nextHeight := node.GetBlockchain().GetChain(-1).GetBestView().GetHeight() + 1
			beaconBlock := <-beaconCh
			if nextHeight != beaconBlock.GetHeight() {
				fmt.Println("Something wrong", nextHeight, beaconBlock.GetHeight())
				panic(1)
			}

		BEACON_WAIT:
			shouldWait := false
			for sid, shardStates := range beaconBlock.Body.ShardState {
				if len(shardStates) > 0 && shardStates[len(shardStates)-1].Height > node.GetBlockchain().GetChain(int(sid)).GetFinalView().GetHeight() {
					shouldWait = true
				}
			}
			//log.Printf("Get and Check beacon block %v - %vs\n", nextHeight, time.Since(time1).Seconds())
			time1 = time.Now()
			if !shouldWait {
				err := beaconChain.InsertBlock(&beaconBlock, common.REGRESSION_TEST)
				log.Printf("Insert beacon block %v - %vs\n", nextHeight, time.Since(time1).Seconds())
				if err != nil {
					log.Println(err)
					goto BEACON_WAIT
				}
			} else {
				time.Sleep(time.Millisecond * 5)
				goto BEACON_WAIT
			}
		}
	}()

	for j := 0; j < 8; j++ {
		go func(shardID int) {
			for {
				time1 := time.Now()
				nextHeight := node.GetBlockchain().GetChain(shardID).GetBestView().GetHeight() + 1
				shardBlock := <-shardCh[shardID]
				if nextHeight != shardBlock.GetHeight() {
					fmt.Println("Something wrong", nextHeight, shardBlock.GetHeight())
					panic(1)
				}
			SHARD_WAIT:
				//fmt.Println("insert shard", shardID, nextHeight)
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
							//fmt.Println("debug create crossshard block", int(sid), int(shardID), blk.BlockHeight, blk.BlockHash.String())
							crossBlk, err := node.GetBlockchain().GetShardBlockByHeightV1(blk.BlockHeight, sid)
							if err != nil {
								panic(err)
							}
							crossX, err := blockchain.CreateCrossShardBlock(crossBlk, byte(shardID))
							if err != nil {
								panic(err)
							}
							//fmt.Println("debug insert cross shard block", int(sid), int(shardID), blk.BlockHeight, crossX.Hash().String())
							node.GetSyncker().InsertCrossShardBlock(crossX)
						}

					}
				}

				//log.Printf("Get and Check shard %v block %v - %vs\n", shardID, nextHeight, time.Since(time1).Seconds())
				time1 = time.Now()
				if !shouldWait {
					err := node.GetBlockchain().GetChain(shardID).(*blockchain.ShardChain).InsertBlock(&shardBlock, common.REGRESSION_TEST)
					log.Printf("Insert shard %v block %v - %vs\n", shardID, nextHeight, time.Since(time1).Seconds())
					if err != nil {
						if strings.Index(err.Error(), "Fetch Beacon Blocks Error") == -1 {
							log.Println(err)
							panic(1)
						} else {
							log.Println("Wait for beacon", shardBlock.Header.BeaconHeight, beaconFinaView)
						}
						goto SHARD_WAIT
					}
				} else {
					time.Sleep(time.Millisecond * 5)
					goto SHARD_WAIT
				}
			}
		}(j)
	}
	select {}
}
