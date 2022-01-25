package blockchain

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/flatfile"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
)

func StoreStateObjectForRepair(
	flatfile *flatfile.FlatFileManager,
	db incdb.Batch,
	hash common.Hash,
	stateObjects map[common.Hash]statedb.StateObject) error {

	indexes, err := StoreStateObjectToFlatFile(flatfile, stateObjects)
	if err != nil {
		return err
	}

	if err := StoreFlatFileTransactionIndex(db, hash, indexes); err != nil {
		return err
	}

	return nil
}

func StoreStateObjectToFlatFile(
	flatfile *flatfile.FlatFileManager,
	stateObjects map[common.Hash]statedb.StateObject) ([]int, error) {

	newIndexes := []int{}

	for _, stateObject := range stateObjects {
		data := statedb.ByteSerialize(stateObject)
		newIndex, err := flatfile.Append(data)
		if err != nil {
			return newIndexes, err
		}
		newIndexes = append(newIndexes, newIndex)
	}

	return newIndexes, nil
}

func StoreFlatFileTransactionIndex(db incdb.Batch, hash common.Hash, indexes []int) error {
	return rawdbv2.StoreFlatFileTransactionIndex(db, hash, indexes)
}

func GetTransactionStateObjectFromFlatFile(
	stateDB *statedb.StateDB,
	flatfile *flatfile.FlatFileManager,
	db incdb.Database,
	hash common.Hash,
) (map[common.Hash]statedb.StateObject, error) {

	stateObjects := make(map[common.Hash]statedb.StateObject)

	indexes, err := rawdbv2.GetFlatFileTransactionIndex(db, hash)
	if err != nil {
		return stateObjects, err
	}

	for _, index := range indexes {
		data, err := flatfile.Read(index)
		if err != nil {
			return stateObjects, err
		}

		stateObject, err := statedb.ByteDeSerialize(stateDB, data)
		if err != nil {
			return stateObjects, err
		}

		stateObjects[stateObject.GetHash()] = stateObject
	}

	return stateObjects, nil
}
