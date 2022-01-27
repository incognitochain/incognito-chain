package blockchain

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/flatfile"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
)

const (
	REPAIR_STATE_CONSENSUS   = 0
	REPAIR_STATE_TRANSACTION = 1
	REPAIR_STATE_FEATURE     = 2
	REPAIR_STATE_REWARD      = 3
	REPAIR_STATE_SLASH       = 4
)

func StoreTransactionStateObjectForRepair(
	flatFileManager *flatfile.FlatFileManager,
	db incdb.Batch,
	hash common.Hash,
	consensusStateObjects map[common.Hash]statedb.StateObject,
	transactionStateObjects map[common.Hash]statedb.StateObject,
	featureStateObjects map[common.Hash]statedb.StateObject,
	rewardStateObjects map[common.Hash]statedb.StateObject,
	slashStateObjects map[common.Hash]statedb.StateObject,
) ([][]int, error) {

	indexes := make([][]int, 5)

	consensusStateObjectIndex, err := StoreStateObjectToFlatFile(flatFileManager, consensusStateObjects)
	if err != nil {
		return [][]int{}, err
	}
	indexes[REPAIR_STATE_CONSENSUS] = consensusStateObjectIndex

	transactionStateObjectIndex, err := StoreStateObjectToFlatFile(flatFileManager, transactionStateObjects)
	if err != nil {
		return [][]int{}, err
	}
	indexes[REPAIR_STATE_TRANSACTION] = transactionStateObjectIndex

	featureStateObjectIndex, err := StoreStateObjectToFlatFile(flatFileManager, featureStateObjects)
	if err != nil {
		return [][]int{}, err
	}
	indexes[REPAIR_STATE_FEATURE] = featureStateObjectIndex

	rewardStateObjectIndex, err := StoreStateObjectToFlatFile(flatFileManager, rewardStateObjects)
	if err != nil {
		return [][]int{}, err
	}
	indexes[REPAIR_STATE_REWARD] = rewardStateObjectIndex

	slashStateObjectIndex, err := StoreStateObjectToFlatFile(flatFileManager, slashStateObjects)
	if err != nil {
		return [][]int{}, err
	}
	indexes[REPAIR_STATE_SLASH] = slashStateObjectIndex

	if len(transactionStateObjectIndex) > 0 {
		Logger.log.Infof("State Object Store Flatfiles \n", transactionStateObjectIndex)
	}

	if err := StoreFlatFileStateObjectIndex(db, hash, indexes); err != nil {
		return [][]int{}, err
	}

	return indexes, nil
}

func StoreStateObjectToFlatFile(
	flatFileManager *flatfile.FlatFileManager,
	stateObjects map[common.Hash]statedb.StateObject,
) ([]int, error) {

	newIndexes := []int{}

	for _, stateObject := range stateObjects {
		data := statedb.ByteSerialize(stateObject)
		newIndex, err := flatFileManager.Append(data)
		if err != nil {
			return newIndexes, err
		}

		newIndexes = append(newIndexes, newIndex)
	}

	return newIndexes, nil
}

func StoreFlatFileStateObjectIndex(db incdb.Batch, hash common.Hash, indexes [][]int) error {
	return rawdbv2.StoreFlatFileStateObjectIndex(db, hash, indexes)
}

func GetStateObjectFromFlatFile(
	stateDBs []*statedb.StateDB,
	flatFileManager *flatfile.FlatFileManager,
	db incdb.Database,
	hash common.Hash,
) ([]map[common.Hash]statedb.StateObject, error) {

	allStateObjects := make([]map[common.Hash]statedb.StateObject, 5)

	indexes, err := rawdbv2.GetFlatFileStateObjectIndex(db, hash)
	if err != nil {
		return allStateObjects, err
	}

	if len(indexes[1]) > 0 {
		Logger.log.Infof("GetStateObjectFromFlatFile, %+v", indexes[3])
	}

	for i := range indexes {
		stateDB := stateDBs[i]
		stateObjects := make(map[common.Hash]statedb.StateObject)

		for _, index := range indexes[i] {
			data, err := flatFileManager.Read(index)
			if err != nil {
				return allStateObjects, err
			}

			stateObject, err := statedb.ByteDeSerialize(stateDB, data)
			if err != nil {
				return allStateObjects, err
			}

			stateObjects[stateObject.GetHash()] = stateObject
		}

		allStateObjects[i] = stateObjects
	}

	return allStateObjects, nil
}
