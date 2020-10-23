package mock

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
)

type Syncker struct {
	Bc                  *blockchain.BlockChain
	LastCrossShardState map[byte]map[byte]uint64
}

type NextCrossShardInfo struct {
	NextCrossShardHeight uint64
	NextCrossShardHash   string
	ConfirmBeaconHeight  uint64
	ConfirmBeaconHash    string
}

const MAX_CROSSX_BLOCK = 10

func (s *Syncker) GetCrossShardBlocksForShardProducer(toShard byte, limit map[byte][]uint64) map[byte][]interface{} {
	var result map[byte][]interface{}
	result = make(map[byte][]interface{})
	lastRequestCrossShard := s.Bc.ShardChain[toShard].GetCrossShardState()
	for i := 0; i < s.Bc.GetConfig().ChainParams.ActiveShards; i++ {
		for {
			if i == int(toShard) {
				break
			}

			//if limit has 0 length, we should break now
			if limit != nil && len(result[byte(i)]) >= len(limit[byte(i)]) {
				break
			}
			requestHeight := lastRequestCrossShard[byte(i)]
			b, err := rawdbv2.GetCrossShardNextHeight(s.Bc.GetConfig().DataBase[common.BeaconChainDataBaseID], byte(i), byte(toShard), uint64(requestHeight))
			if err != nil {
				log.Println(err)
				return result
			}
			var nextCrossShardInfo = new(NextCrossShardInfo)
			err = json.Unmarshal(b, nextCrossShardInfo)
			if err != nil {
				log.Fatalln(err)
				return nil
			}
			beaconHash, _ := common.Hash{}.NewHashFromStr(nextCrossShardInfo.ConfirmBeaconHash)
			beaconBlockBytes, err := rawdbv2.GetBeaconBlockByHash(s.Bc.GetBeaconChainDatabase(), *beaconHash)
			if err != nil {
				break
			}

			beaconBlock := new(blockchain.BeaconBlock)
			json.Unmarshal(beaconBlockBytes, beaconBlock)
			for _, shardState := range beaconBlock.Body.ShardState[byte(i)] {
				if shardState.Height == nextCrossShardInfo.NextCrossShardHeight {
					block, _, err := s.Bc.GetShardBlockByHash(shardState.Hash)
					if err != nil {
						log.Fatalln(err)
						return nil
					}
					blkXShard, err := block.CreateCrossShardBlock(toShard)
					if err != nil {
						log.Fatalln(err)
						return nil
					}
					beaconConsensusRootHash, err := s.Bc.GetBeaconConsensusRootHash(s.Bc.GetBeaconBestState(), beaconBlock.GetHeight()-1)
					if err != nil {
						log.Fatalln(err)
						return nil
					}
					beaconConsensusStateDB, err := statedb.NewWithPrefixTrie(beaconConsensusRootHash, statedb.NewDatabaseAccessWarper(s.Bc.GetBeaconChainDatabase()))
					committee := statedb.GetOneShardCommittee(beaconConsensusStateDB, byte(i))
					err = s.Bc.ShardChain[byte(i)].ValidateBlockSignatures(blkXShard, committee)
					if err != nil {
						log.Fatalln(err)
						return nil
					}
					//add to result list
					result[byte(i)] = append(result[byte(i)], blkXShard)
					//has block in pool, update request pointer
					lastRequestCrossShard[byte(i)] = nextCrossShardInfo.NextCrossShardHeight
					break
				}
			}

			//cannot append crossshard for a shard (no block in pool, validate error) => break process for this shard
			if requestHeight == lastRequestCrossShard[byte(i)] {
				break
			}

			if len(result[byte(i)]) >= MAX_CROSSX_BLOCK {
				break
			}
		}
	}

	return result
}

func (s *Syncker) GetCrossShardBlocksForShardValidator(toShard byte, list map[byte][]uint64) (map[byte][]interface{}, error) {
	return s.GetCrossShardBlocksForShardProducer(toShard, list), nil
}

func (s *Syncker) SyncMissingBeaconBlock(ctx context.Context, peerID string, fromHash common.Hash) {
	return
}

func (s *Syncker) SyncMissingShardBlock(ctx context.Context, peerID string, sid byte, fromHash common.Hash) {
	return
}

type LastCrossShardBeaconProcess struct {
	BeaconHeight        uint64
	LastCrossShardState map[byte]map[byte]uint64
}

//watching confirm beacon block and update cross shard info (which beacon confirm crossshard block N of shard X)
func (s *Syncker) UpdateConfirmCrossShard() {
	state := rawdbv2.GetLastBeaconStateConfirmCrossShard(s.Bc.GetBeaconChainDatabase())
	lastBeaconStateConfirmCrossX := new(LastCrossShardBeaconProcess)
	_ = json.Unmarshal(state, &lastBeaconStateConfirmCrossX)
	lastBeaconHeightConfirmCrossX := uint64(1)
	if lastBeaconStateConfirmCrossX.BeaconHeight != 0 {
		s.LastCrossShardState = lastBeaconStateConfirmCrossX.LastCrossShardState
		lastBeaconHeightConfirmCrossX = lastBeaconStateConfirmCrossX.BeaconHeight
	}
	fmt.Println("lastBeaconHeightConfirmCrossX", lastBeaconHeightConfirmCrossX)
	for {
		if lastBeaconHeightConfirmCrossX > s.Bc.BeaconChain.GetFinalViewHeight() {
			//fmt.Println("DEBUG:larger than final view", s.chain.GetFinalViewHeight())
			time.Sleep(time.Second * 5)
			continue
		}
		blkhash, err := rawdbv2.GetFinalizedBeaconBlockHashByIndex(s.Bc.GetBeaconChainDatabase(), lastBeaconHeightConfirmCrossX)
		if err != nil {
			fmt.Println(err)
			return
		}
		beaconBlock, _, err := s.Bc.GetBeaconBlockByHash(*blkhash)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = processBeaconForConfirmmingCrossShard(s.Bc.GetBeaconChainDatabase(), beaconBlock, s.LastCrossShardState)
		if err == nil {
			lastBeaconHeightConfirmCrossX++
			if lastBeaconHeightConfirmCrossX%1000 == 0 {
				fmt.Println("store lastBeaconHeightConfirmCrossX", lastBeaconHeightConfirmCrossX)
				rawdbv2.StoreLastBeaconStateConfirmCrossShard(s.Bc.GetBeaconChainDatabase(), LastCrossShardBeaconProcess{lastBeaconHeightConfirmCrossX, s.LastCrossShardState})
			}
		} else {
			fmt.Println(err)
			time.Sleep(time.Second * 5)
		}

	}
}

func processBeaconForConfirmmingCrossShard(database incdb.Database, beaconBlock *blockchain.BeaconBlock, lastCrossShardState map[byte]map[byte]uint64) error {
	if beaconBlock != nil && beaconBlock.Body.ShardState != nil {
		for fromShard, shardBlocks := range beaconBlock.Body.ShardState {
			for _, shardBlock := range shardBlocks {
				for _, toShard := range shardBlock.CrossShard {

					if fromShard == toShard {
						continue
					}
					if lastCrossShardState[fromShard] == nil {
						lastCrossShardState[fromShard] = make(map[byte]uint64)
					}
					lastHeight := lastCrossShardState[fromShard][toShard] // get last cross shard height from shardID  to crossShardShardID
					waitHeight := shardBlock.Height

					info := NextCrossShardInfo{
						waitHeight,
						shardBlock.Hash.String(),
						beaconBlock.GetHeight(),
						beaconBlock.Hash().String(),
					}
					fmt.Println("DEBUG: processBeaconForConfirmmingCrossShard ", fromShard, toShard, info)
					b, _ := json.Marshal(info)
					fmt.Println("debug StoreCrossShardNextHeight", fromShard, toShard, lastHeight, string(b))
					err := rawdbv2.StoreCrossShardNextHeight(database, fromShard, toShard, lastHeight, b)
					if err != nil {
						return err
					}
					if lastCrossShardState[fromShard] == nil {
						lastCrossShardState[fromShard] = make(map[byte]uint64)
					}
					lastCrossShardState[fromShard][toShard] = waitHeight //update lastHeight to waitHeight
				}
			}
		}
	}
	return nil
}
