package lvdb

import (
	"bytes"
	"encoding/binary"
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/pkg/errors"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
)

/*
	Store new shard block
	Store 2 new record per new one block if success
	Record 1: Use to query all block in one shard
	- Key: s-{shardID}b-{blockHash}
	- Value: b-{blockHash}
	Record 2: Use to query a block by hash
	- Key: b-{blockHash}
	- Value: {block}
*/
func (db *db) StoreShardBlock(v interface{}, hash common.Hash, shardID byte, bd *[]database.BatchData) error {
	var (
		// key: b-{blockhash}:block
		keyBlockHash = addPrefixToKeyHash(string(blockKeyPrefix), hash)
		// key: s-{shardID}b-{[blockhash]}:b-{blockhash}
		keyShardBlock = append(append(shardIDPrefix, shardID), keyBlockHash...)
	)
	if ok, _ := db.HasValue(keyShardBlock); ok {
		return database.NewDatabaseError(database.BlockExisted, errors.Errorf("block %s already exists", hash.String()))
	}
	val, err := json.Marshal(v)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Marshal"))
	}

	if bd != nil {
		*bd = append(*bd, database.BatchData{keyShardBlock, keyBlockHash})
		*bd = append(*bd, database.BatchData{keyBlockHash, val})
		return nil
	}

	if err := db.Put(keyShardBlock, keyBlockHash); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.Put"))
	}
	if err := db.Put(keyBlockHash, val); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.Put"))
	}
	return nil
}

/*
	Query a block existence by hash. Return true if block exist otherwise false
*/
func (db *db) HasBlock(hash common.Hash) (bool, error) {
	exists, err := db.HasValue(addPrefixToKeyHash(string(blockKeyPrefix), hash))
	if err != nil {
		return false, err
	} else {
		return exists, nil
	}
}

/*
	Query a block by hash. Return block if existence
*/
func (db *db) FetchBlock(hash common.Hash) ([]byte, error) {
	block, err := db.lvdb.Get(addPrefixToKeyHash(string(blockKeyPrefix), hash), nil)
	if err != nil {
		if err == lvdberr.ErrNotFound {
			return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
		}
		return []byte{}, nil
	}
	ret := make([]byte, len(block))
	copy(ret, block)
	return ret, nil
}

/*
	Store index of block in shard
	Record 1: use hash to get block index ~= block height in a pariticular shard
	+ key: i-{hash}
	+ value: {index-shardID}
	Record 2: use block index to get block hash
	+ key: {index-shardID}
	+ value: {hash}
*/
func (db *db) StoreShardBlockIndex(hash common.Hash, idx uint64, shardID byte, bd *[]database.BatchData) error {
	buf := make([]byte, 9)
	binary.LittleEndian.PutUint64(buf, idx)
	buf[8] = shardID

	if bd != nil {
		*bd = append(*bd, database.BatchData{addPrefixToKeyHash(string(blockKeyIdxPrefix), hash), buf})
		*bd = append(*bd, database.BatchData{buf, hash[:]})
		return nil
	}
	//{i-[hash]}:index-shardID
	if err := db.Put(addPrefixToKeyHash(string(blockKeyIdxPrefix), hash), buf); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}
	//{index-shardID}:[hash]
	if err := db.Put(buf, hash[:]); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}

func (db *db) GetIndexOfBlock(hash common.Hash) (uint64, byte, error) {
	var idx uint64
	var shardID byte
	b, err := db.Get(addPrefixToKeyHash(string(blockKeyIdxPrefix), hash))
	//{i-[hash]}:index-shardID
	if err != nil {
		return 0, 0, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.get"))
	}
	if err := binary.Read(bytes.NewReader(b[:8]), binary.LittleEndian, &idx); err != nil {
		return 0, 0, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "binary.Read"))
	}
	if err = binary.Read(bytes.NewReader(b[8:]), binary.LittleEndian, &shardID); err != nil {
		return 0, 0, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "binary.Read"))
	}
	return idx, shardID, nil
}

/*
	Query a block by hash. Return block if existence otherwise error
	This function ONLY work when stored whole shard block,
	1. Delete record block by hash
*/
func (db *db) DeleteBlock(hash common.Hash, idx uint64, shardID byte) error {
	var (
		err error
		// key: b-{blockhash}:block
		keyBlockHash = addPrefixToKeyHash(string(blockKeyPrefix), hash)
		// key: s-{shardID}b-{[blockhash]}:b-{blockhash}
		keyShardBlock = append(append(shardIDPrefix, shardID), keyBlockHash...)
		//{i-[hash]}:index-shardID
		keyBlockIndex = addPrefixToKeyHash(string(blockKeyIdxPrefix), hash)
	)
	//{index-shardID}: hash
	var buf = make([]byte, 9)
	binary.LittleEndian.PutUint64(buf, idx)
	buf[8] = shardID
	// Delete block
	err = db.Delete(keyBlockHash)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Delete"))
	}
	err = db.Delete(keyShardBlock)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Delete"))
	}

	// Delete block index
	err = db.Delete(keyBlockIndex)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}

	err = db.Delete(buf)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	return nil
}

func (db *db) StoreShardViews(val []byte, shardID byte, bd *[]database.BatchData) error {
	key := append(shardViewsPrefix, shardID)
	if bd != nil {
		*bd = append(*bd, database.BatchData{key, val})
		return nil
	}
	if err := db.Put(key, val); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, err)
	}
	return nil
}

func (db *db) FetchShardViews(shardID byte) ([]byte, error) {
	key := append(shardViewsPrefix, shardID)
	block, err := db.Get(key)
	if err != nil {
		return nil, database.NewDatabaseError(database.UnexpectedError, err)
	}
	return block, nil
}

func (db *db) StoreShardBestState(v interface{}, shardID byte, bd *[]database.BatchData) error {
	val, err := json.Marshal(v)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Marshal"))
	}
	key := append(bestBlockKeyPrefix, shardID)

	if bd != nil {
		*bd = append(*bd, database.BatchData{key, val})
		return nil
	}
	if err := db.Put(key, val); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.put"))
	}
	return nil
}

func (db *db) FetchShardBestState(shardID byte) ([]byte, error) {
	key := append(bestBlockKeyPrefix, shardID)
	block, err := db.Get(key)
	if err != nil {
		return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.get"))
	}
	return block, nil
}

func (db *db) CleanShardBestState() error {
	for shardID := byte(0); shardID < common.MaxShardNumber; shardID++ {
		key := append(bestBlockKeyPrefix, shardID)
		err := db.Delete(key)
		if err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.delete"))
		}
	}
	return nil
}

func (db *db) GetBlockByIndex(idx uint64, shardID byte) (common.Hash, error) {
	buf := make([]byte, 9)
	binary.LittleEndian.PutUint64(buf, idx)
	buf[8] = shardID
	// {index-shardID}: {blockhash}
	b, err := db.Get(buf)
	if err != nil {
		return common.Hash{}, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.GetBlockByIndex"))
	}
	hash := new(common.Hash)
	err1 := hash.SetBytes(b[:])
	if err1 != nil {
		return common.Hash{}, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.GetBlockByIndex"))
	}
	return *hash, nil
}

func (db *db) StoreCommitteeFromShardBestState(shardID byte, shardHeight uint64, v interface{}) error {
	key := append(shardPrefix, shardID)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, shardHeight)
	key = append(key, buf[:]...)

	val, err := json.Marshal(v)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Marshal"))
	}

	if err := db.Put(key, val); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}

func (db *db) FetchCommitteeFromShardBestState(shardID byte, shardHeight uint64) ([]byte, error) {
	key := append(shardPrefix, shardID)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, shardHeight)
	key = append(key, buf[:]...)

	b, err := db.Get(key)
	if err != nil {
		return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.get"))
	}
	return b, nil
}

func (db *db) HasShardCommitteeByHeight(height uint64) (bool, error) {
	key := append(beaconPrefix, shardIDPrefix...)
	key = append(key, committeePrefix...)
	key = append(key, heightPrefix...)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, height)
	key = append(key, buf[:]...)

	exist, err := db.HasValue(key)
	if err != nil {
		return false, database.NewDatabaseError(database.HasShardCommitteeByHeightError, err)
	}
	return exist, nil
}
