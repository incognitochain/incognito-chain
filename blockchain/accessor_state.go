package blockchain

import (
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/flatfile"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
)

const (
	SHARD_CONSENSUS_STATEDB   = 0
	SHARD_TRANSACTION_STATEDB = 1
	SHARD_FEATURE_STATEDB     = 2
	SHARD_REWARD_STATEDB      = 3
	SHARD_SLASH_STATEDB       = 4
)

func GetStateObjectFromFlatFile(
	stateDBs []*statedb.StateDB,
	flatFileManager *flatfile.FlatFileManager,
	db incdb.Database,
	blockHash common.Hash,
	sRH *ShardRootHash,
) ([]map[common.Hash]statedb.StateObject, []int, error) {

	allStateObjects := make([]map[common.Hash]statedb.StateObject, 5)
	indexes := make([]int, 5)
	var err error

	indexes[SHARD_CONSENSUS_STATEDB], err = statedb.GetFlatFileStateObjectIndex(db, blockHash, sRH.ConsensusStateDBRootHash)
	if err != nil {
		return allStateObjects, nil, err
	}
	indexes[SHARD_TRANSACTION_STATEDB], err = statedb.GetFlatFileStateObjectIndex(db, blockHash, sRH.TransactionStateDBRootHash)
	if err != nil {
		return allStateObjects, nil, err
	}
	indexes[SHARD_FEATURE_STATEDB], err = statedb.GetFlatFileStateObjectIndex(db, blockHash, sRH.FeatureStateDBRootHash)
	if err != nil {
		return allStateObjects, nil, err
	}
	indexes[SHARD_REWARD_STATEDB], err = statedb.GetFlatFileStateObjectIndex(db, blockHash, sRH.RewardStateDBRootHash)
	if err != nil {
		return allStateObjects, nil, err
	}
	indexes[SHARD_SLASH_STATEDB], err = statedb.GetFlatFileStateObjectIndex(db, blockHash, sRH.SlashStateDBRootHash)
	if err != nil {
		return allStateObjects, nil, err
	}

	for i := range indexes {

		stateDB := stateDBs[i]

		data, err := flatFileManager.Read(indexes[i])
		if err != nil {
			return allStateObjects, nil, err
		}
		stateObjects, err := statedb.MapByteDeserialize(stateDB, data)
		if err != nil {
			return allStateObjects, nil, err
		}

		allStateObjects[i] = stateObjects
	}

	return allStateObjects, indexes, nil
}

func (bc *BlockChain) GetPivotBlock(shardID byte) (*types.ShardBlock, error) {

	db := bc.GetShardChainDatabase(shardID)

	hash, err := statedb.GetLatestPivotBlock(db, shardID)
	if err != nil {
		return nil, err
	}

	res, _, err := bc.GetShardBlockByHash(hash)
	if err != nil {
		return nil, err
	}

	return res, nil
}
