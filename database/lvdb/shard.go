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
func (db *db) StoreShardBlock(v interface{}, hash common.Hash, shardID byte) error {
	var (
		// key: b-{blockhash}:block
		keyBlockHash = db.GetKey(string(blockKeyPrefix), hash)
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
	exists, err := db.HasValue(db.GetKey(string(blockKeyPrefix), hash))
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
	block, err := db.Get(db.GetKey(string(blockKeyPrefix), hash))
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
func (db *db) StoreShardBlockIndex(hash common.Hash, idx uint64, shardID byte) error {
	buf := make([]byte, 9)
	binary.LittleEndian.PutUint64(buf, idx)
	buf[8] = shardID
	//{i-[hash]}:index-shardID
	if err := db.Put(db.GetKey(string(blockKeyIdxPrefix), hash), buf); err != nil {
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
	b, err := db.Get(db.GetKey(string(blockKeyIdxPrefix), hash))
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
		keyBlockHash = db.GetKey(string(blockKeyPrefix), hash)
		// key: s-{shardID}b-{[blockhash]}:b-{blockhash}
		keyShardBlock = append(append(shardIDPrefix, shardID), keyBlockHash...)
		//{i-[hash]}:index-shardID
		keyBlockIndex = db.GetKey(string(blockKeyIdxPrefix), hash)
	)
	//{index-shardID}: hash
	var buf = make([]byte, 9)
	binary.LittleEndian.PutUint64(buf, idx)
	buf[8] = shardID
	// Delete block
	err = db.lvdb.Delete(keyBlockHash, nil)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Delete"))
	}
	err = db.lvdb.Delete(keyShardBlock, nil)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Delete"))
	}

	// Delete block index
	err = db.lvdb.Delete(keyBlockIndex, nil)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}

	err = db.lvdb.Delete(buf, nil)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	return nil
}

func (db *db) StoreShardBestState(v interface{}, shardID byte) error {
	val, err := json.Marshal(v)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Marshal"))
	}
	key := append(bestBlockKey, shardID)
	if err := db.Put(key, val); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.put"))
	}
	return nil
}

func (db *db) FetchShardBestState(shardID byte) ([]byte, error) {
	key := append(bestBlockKey, shardID)
	block, err := db.Get(key)
	if err != nil {
		return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.get"))
	}
	return block, nil
}

func (db *db) CleanShardBestState() error {
	for shardID := byte(0); shardID < common.MAX_SHARD_NUMBER; shardID++ {
		key := append(bestBlockKey, shardID)
		err := db.lvdb.Delete(key, nil)
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

//StoreIncomingCrossShard which store crossShardHash from which shard has been include in which block height
func (db *db) StoreIncomingCrossShard(shardID byte, crossShardID byte, blkHeight uint64, crossBlkHash common.Hash) error {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, blkHeight)
	prefix := append([]byte{shardID}, append([]byte{crossShardID}, crossBlkHash[:]...)...)
	// csh-ShardID-CrossShardID-CrossShardBlockHash : ShardBlockHeight
	key := append(crossShardKeyPrefix, prefix...)
	if ok, _ := db.HasValue(key); ok {
		return database.NewDatabaseError(database.BlockExisted, errors.Errorf("block %d already exists", blkHeight))
	}
	if err := db.Put(key, buf); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}

func (db *db) HasIncomingCrossShard(shardID byte, crossShardID byte, crossBlkHash common.Hash) error {
	prefix := append([]byte{shardID}, append([]byte{crossShardID}, crossBlkHash[:]...)...)
	// csh-ShardID-CrossShardID-CrossShardBlockHash : ShardBlockHeight
	key := append(crossShardKeyPrefix, prefix...)
	if ok, _ := db.HasValue(key); ok {
		return nil
	}
	return database.NewDatabaseError(database.BlockExisted, errors.Errorf("Cross Shard Block doesn't exist"))
}

func (db *db) GetIncomingCrossShard(shardID byte, crossShardID byte, crossBlkHash common.Hash) (uint64, error) {
	prefix := append([]byte{shardID}, append([]byte{crossShardID}, crossBlkHash[:]...)...)
	// csh-ShardID-CrossShardID-CrossShardBlockHash : ShardBlockHeight
	key := append(crossShardKeyPrefix, prefix...)
	b, err := db.Get(key)
	if err != nil {
		return 0, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	var idx uint64
	if err := binary.Read(bytes.NewReader(b[:]), binary.LittleEndian, &idx); err != nil {
		return 0, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "binary.Read"))
	}
	return idx, nil
}
