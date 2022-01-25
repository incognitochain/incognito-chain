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
	flatfile *flatfile.FlatFileManager,
	db incdb.Batch,
	hash common.Hash,
	consensusStateObjects map[common.Hash]statedb.StateObject,
	transactionStateObjects map[common.Hash]statedb.StateObject,
	featureStateObjects map[common.Hash]statedb.StateObject,
	rewardStateObjects map[common.Hash]statedb.StateObject,
	slashStateObjects map[common.Hash]statedb.StateObject,
) error {

	indexes := make([][]int, 5)

	consensusStateObjectIndex, err := StoreStateObjectToFlatFile(flatfile, consensusStateObjects)
	if err != nil {
		return err
	}
	indexes[REPAIR_STATE_CONSENSUS] = consensusStateObjectIndex

	transactionStateObjectIndex, err := StoreStateObjectToFlatFile(flatfile, transactionStateObjects)
	if err != nil {
		return err
	}
	indexes[REPAIR_STATE_TRANSACTION] = transactionStateObjectIndex

	featureStateObjectIndex, err := StoreStateObjectToFlatFile(flatfile, featureStateObjects)
	if err != nil {
		return err
	}
	indexes[REPAIR_STATE_FEATURE] = featureStateObjectIndex

	rewardStateObjectIndex, err := StoreStateObjectToFlatFile(flatfile, rewardStateObjects)
	if err != nil {
		return err
	}
	indexes[REPAIR_STATE_REWARD] = rewardStateObjectIndex

	slashStateObjectIndex, err := StoreStateObjectToFlatFile(flatfile, slashStateObjects)
	if err != nil {
		return err
	}
	indexes[REPAIR_STATE_SLASH] = slashStateObjectIndex

	if err := StoreFlatFileStateObjectIndex(db, hash, indexes); err != nil {
		return err
	}

	return nil
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

func GetTransactionStateObjectFromFlatFile(
	consensusStateDB *statedb.StateDB,
	transactionStateDB *statedb.StateDB,
	featureStateDB *statedb.StateDB,
	rewardStateDB *statedb.StateDB,
	slashStateDB *statedb.StateDB,
	flatFileManager *flatfile.FlatFileManager,
	db incdb.Database,
	hash common.Hash,
) ([]map[common.Hash]statedb.StateObject, error) {

	allStateObjects := make([]map[common.Hash]statedb.StateObject, 5)

	indexes, err := rawdbv2.GetFlatFileStateObjectIndex(db, hash)
	if err != nil {
		return allStateObjects, err
	}

	for i := range indexes {
		stateDB := &statedb.StateDB{}
		switch i {
		case REPAIR_STATE_CONSENSUS:
			stateDB = consensusStateDB
		case REPAIR_STATE_TRANSACTION:
			stateDB = transactionStateDB
		case REPAIR_STATE_FEATURE:
			stateDB = featureStateDB
		case REPAIR_STATE_REWARD:
			stateDB = rewardStateDB
		case REPAIR_STATE_SLASH:
			stateDB = slashStateDB
		}
		stateObjects := make(map[common.Hash]statedb.StateObject)

		for index := range indexes[i] {
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
