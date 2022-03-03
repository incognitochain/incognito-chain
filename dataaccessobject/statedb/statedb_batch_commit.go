package statedb

import (
	"encoding/binary"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
)

var (
	splitter                       = []byte("-[-]-")
	flatFileStateObjectIndexPrefix = []byte("ff-sob-" + string(splitter))
	fullSyncPivotBlockKey          = []byte("Full-Sync-Latest-Pivot-Block")
)

func GetFlatFileStateObjectIndexKey(hash common.Hash) []byte {

	temp := make([]byte, len(flatFileStateObjectIndexPrefix))
	copy(temp, flatFileStateObjectIndexPrefix)

	return append(temp, hash[:]...)
}

func GetFullSyncPivotBlockKey(shardID byte) []byte {

	temp := make([]byte, len(fullSyncPivotBlockKey))
	copy(temp, fullSyncPivotBlockKey)

	temp = append(temp, shardID)

	return temp
}

func CacheStateObjectForRepair(
	flatFile FlatFile,
	db incdb.KeyValueWriter,
	rootHash common.Hash,
	stateObjects map[common.Hash]StateObject,
) ([]int, error) {

	indexes := make([]int, 5)

	stateObjectsIndex, err := StoreStateObjectToFlatFile(flatFile, stateObjects)
	if err != nil {
		return []int{}, err
	}
	if err := StoreFlatFileStateObjectIndex(db, rootHash, stateObjectsIndex); err != nil {
		return []int{}, err
	}

	return indexes, nil
}

func StoreStateObjectToFlatFile(
	flatFile FlatFile,
	stateObjects map[common.Hash]StateObject,
) (int, error) {

	res := MapByteSerialize(stateObjects)

	return flatFile.Append(res)
}

func StoreFlatFileStateObjectIndex(db incdb.KeyValueWriter, hash common.Hash, index int) error {

	key := GetFlatFileStateObjectIndexKey(hash)

	value := make([]byte, 4)
	binary.LittleEndian.PutUint32(value, uint32(index))

	if err := db.Put(key, value); err != nil {
		return err
	}

	return nil
}

func GetStateObjectFromFlatFile(
	stateDB *StateDB,
	flatFile FlatFile,
	db incdb.Database,
	rootHash common.Hash,
) (map[common.Hash]StateObject, int, error) {

	stateObject := make(map[common.Hash]StateObject)

	index, err := GetFlatFileStateObjectIndex(db, rootHash)
	if err != nil {
		return stateObject, 0, err
	}

	data, err := flatFile.Read(index)
	if err != nil {
		return stateObject, 0, err
	}
	stateObjects, err := MapByteDeserialize(stateDB, data)
	if err != nil {
		return stateObject, 0, err
	}

	return stateObjects, index, nil
}

func GetFlatFileStateObjectIndex(db incdb.KeyValueReader, hash common.Hash) (int, error) {

	key := GetFlatFileStateObjectIndexKey(hash)

	value, err := db.Get(key)
	if err != nil {
		return 0, err
	}

	return int(binary.LittleEndian.Uint32(value)), nil
}

func GetPivotBlockHash(reader incdb.KeyValueReader, shardID byte) (common.Hash, error) {
	value, err := reader.Get(GetFullSyncPivotBlockKey(shardID))
	if err != nil {
		return common.Hash{}, err
	}

	h, err := common.Hash{}.NewHash(value)

	return *h, err
}

func StoreLatestPivotBlock(writer incdb.KeyValueWriter, shardID byte, hash common.Hash) error {
	return writer.Put(GetFullSyncPivotBlockKey(shardID), hash[:])
}
