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
	"os"
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
	for {
		for i := -1; i < 8; i++ {
			nextHeight := node.GetBlockchain().GetChain(i).GetBestView().GetHeight() + 1
			fmt.Println("Insert", i, nextHeight)
			if i == -1 { //beacon
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
				if shouldWait == false {
					err = beaconChain.InsertBlock(beaconBlock, common.REGRESSION_TEST)
					if err != nil {
						fmt.Printf("%+v", beaconBlock)
						panic(fmt.Sprintf("Insert beacon fail! Block %v - %v. Error: %+v", beaconBlock.GetHeight(), beaconBlock.Hash().String(), err))
					}
				}
			} else {
				shardBlock, err := GetShardBlockByHeight(db[i], byte(i), nextHeight)
				if err != nil {
					fmt.Println("Exit with shard", i, nextHeight)
					os.Exit(0)
				}
				shouldWait := false
				if shardBlock.Header.BeaconHeight > beaconChain.GetFinalViewHeight() {
					shouldWait = true
				}
				for sid, cross := range shardBlock.Body.CrossTransactions {
					if cross[len(cross)-1].BlockHeight > node.GetBlockchain().GetChain(int(sid)).GetBestView().GetHeight() {
						shouldWait = true
					}
				}
				if !shouldWait {
					err = node.GetBlockchain().GetChain(i).(*blockchain.ShardChain).InsertBlock(shardBlock, common.REGRESSION_TEST)
					if err != nil {
						fmt.Printf("%+v", shardBlock)
						panic(fmt.Sprintf("Insert shard fail! Block %v - %v. Error: %+v", shardBlock.GetHeight(), shardBlock.Hash().String(), err))
					}
					crossX := blockchain.CreateAllCrossShardBlock(shardBlock, node.GetBlockchain().GetChainParams().ActiveShards)
					for _, blk := range crossX {
						node.GetSyncker().InsertCrossShardBlock(blk)
					}
				}

			}
		}
	}
}
