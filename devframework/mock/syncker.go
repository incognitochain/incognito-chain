package mock

import (
	"context"
	"encoding/json"
	"log"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type Syncker struct {
	Bc *blockchain.BlockChain
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
				return nil
			}
			var nextCrossShardInfo = new(NextCrossShardInfo)
			err = json.Unmarshal(b, nextCrossShardInfo)
			if err != nil {
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
	var result map[byte][]interface{}
	result = make(map[byte][]interface{})
	return result, nil
}

func (s *Syncker) SyncMissingBeaconBlock(ctx context.Context, peerID string, fromHash common.Hash) {
	return
}

func (s *Syncker) SyncMissingShardBlock(ctx context.Context, peerID string, sid byte, fromHash common.Hash) {
	return
}
