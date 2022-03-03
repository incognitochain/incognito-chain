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

func GetFlatFileStateObjectIndexKey(blockHash, rootHash common.Hash) []byte {

	temp := make([]byte, len(flatFileStateObjectIndexPrefix))
	copy(temp, flatFileStateObjectIndexPrefix)

	temp = append(temp, blockHash[:]...)
	temp = append(temp, splitter...)
	temp = append(temp, rootHash[:]...)

	return temp
}

func GetFullSyncPivotBlockKey(shardID byte) []byte {

	temp := make([]byte, len(fullSyncPivotBlockKey))
	copy(temp, fullSyncPivotBlockKey)

	temp = append(temp, shardID)

	return temp
}

func CacheDirtyObjectForRepair(
	flatFile FlatFile,
	db incdb.KeyValueWriter,
	blockHash common.Hash,
	rootHash common.Hash,
	stateObjects map[common.Hash]StateObject,
) (int, error) {

	stateObjectsIndex, err := StoreStateObjectToFlatFile(flatFile, stateObjects)
	if err != nil {
		return 0, err
	}

	if err := StoreFlatFileStateObjectIndex(db, blockHash, rootHash, stateObjectsIndex); err != nil {
		return 0, err
	}

	return stateObjectsIndex, nil
}

func StoreStateObjectToFlatFile(
	flatFile FlatFile,
	stateObjects map[common.Hash]StateObject,
) (int, error) {

	res := MapByteSerialize(stateObjects)

	return flatFile.Append(res)
}

func StoreFlatFileStateObjectIndex(db incdb.KeyValueWriter, blockHash, rootHash common.Hash, index int) error {

	key := GetFlatFileStateObjectIndexKey(blockHash, rootHash)

	value := make([]byte, 4)
	binary.LittleEndian.PutUint32(value, uint32(index))

	if err := db.Put(key, value); err != nil {
		return err
	}

	return nil
}

func GetFlatFileStateObjectIndex(db incdb.KeyValueReader, blockHash, rootHash common.Hash) (int, error) {

	key := GetFlatFileStateObjectIndexKey(blockHash, rootHash)

	value, err := db.Get(key)
	if err != nil {
		return 0, err
	}

	return int(binary.LittleEndian.Uint32(value)), nil
}

func GetLatestPivotBlock(reader incdb.KeyValueReader, shardID byte) (common.Hash, error) {
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
