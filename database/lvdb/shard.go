package lvdb

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/pkg/errors"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
)

func (db *db) StoreShardBlock(v interface{}, shardID byte) error {
	h, ok := v.(hasher)
	if !ok {
		return database.NewDatabaseError(database.NotImplHashMethod, errors.New("v must implement Hash() method"))
	}
	var (
		hash = h.Hash()
		key  = append(append(shardIDPrefix, shardID), append(blockKeyPrefix, hash[:]...)...)
		// PubKey should look like this s10{b-[blockhash]}:{b-[blockhash]}
		keyB = append(blockKeyPrefix, hash[:]...)
		// PubKey should look like this {b-blockhash}:block
	)
	if ok, _ := db.HasValue(key); ok {
		return database.NewDatabaseError(database.BlockExisted, errors.Errorf("block %s already exists", hash.String()))
	}
	val, err := json.Marshal(v)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Marshal"))
	}
	if err := db.Put(key, keyB); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.Put"))
	}
	if err := db.Put(keyB, val); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.Put"))
	}
	return nil
}

func (db *db) StoreShardBlockHeader(v interface{}, hash *common.Hash, shardID byte) error {
	//fmt.Println("Log in StoreShardBlockHeader", v, hash, shardID)
	var (
		key = append(append(shardIDPrefix, shardID), append(blockKeyPrefix, hash[:]...)...)
		// PubKey should look like this c10{bh-[blockhash]}:{bh-[blockhash]}
		keyB = append(blockKeyPrefix, hash[:]...)
		// PubKey should look like this {bh-blockhash}:block
	)
	if ok, _ := db.HasValue(key); ok {
		return database.NewDatabaseError(database.BlockExisted, errors.Errorf("block %s already exists", hash.String()))
	}
	val, err := json.Marshal(v)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Marshal"))
	}
	if err := db.Put(key, keyB); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.Put"))
	}
	if err := db.Put(keyB, val); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.Put"))
	}
	//fmt.Println("Test StoreShardBlockHeader keyB: ", string(keyB))
	return nil
}

func (db *db) HasBlock(hash *common.Hash) (bool, error) {
	exists, err := db.HasValue(db.GetKey(string(blockKeyPrefix), hash))
	if err != nil {
		return false, err
	} else {
		return exists, nil
	}
}

func (db *db) FetchBlock(hash *common.Hash) ([]byte, error) {
	block, err := db.lvdb.Get(db.GetKey(string(blockKeyPrefix), hash), nil)
	if err != nil {
		if err != lvdberr.ErrNotFound {
			return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
		}
		return []byte{}, nil
	}
	ret := make([]byte, len(block))
	copy(ret, block)
	return ret, nil
}

func (db *db) DeleteBlock(hash *common.Hash, idx uint64, shardID byte) error {
	// Delete block
	err := db.lvdb.Delete(db.GetKey(string(blockKeyPrefix), hash), nil)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}

	// Delete block index
	err = db.lvdb.Delete(db.GetKey(string(blockKeyIdxPrefix), hash), nil)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	buf := make([]byte, 9)
	binary.LittleEndian.PutUint64(buf, idx)
	buf[8] = shardID
	err = db.lvdb.Delete(buf, nil)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	return nil
}

func (db *db) StoreBestState(v interface{}, shardID byte) error {
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

func (db *db) FetchBestState(shardID byte) ([]byte, error) {
	key := append(bestBlockKey, shardID)
	block, err := db.Get(key)
	if err != nil {
		return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.get"))
	}
	return block, nil
}

func (db *db) CleanBestState() error {
	for shardID := byte(0); shardID < common.TotalValidators; shardID++ {
		key := append(bestBlockKey, shardID)
		err := db.lvdb.Delete(key, nil)
		if err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.delete"))
		}
	}
	return nil
}

func (db *db) StoreShardBlockIndex(h *common.Hash, idx uint64, shardID byte) error {
	buf := make([]byte, 9)
	binary.LittleEndian.PutUint64(buf, idx)
	buf[8] = shardID
	//{i-[hash]}:index-shardID
	if err := db.lvdb.Put(db.GetKey(string(blockKeyIdxPrefix), h), buf, nil); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}
	//{index-shardID}:[hash]
	if err := db.lvdb.Put(buf, h[:], nil); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}

func (db *db) GetIndexOfBlock(h *common.Hash) (uint64, byte, error) {
	b, err := db.lvdb.Get(db.GetKey(string(blockKeyIdxPrefix), h), nil)
	//{i-[hash]}:index-shardID
	if err != nil {
		return 0, 0, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.get"))
	}

	var idx uint64
	var shardID byte
	if err := binary.Read(bytes.NewReader(b[:8]), binary.LittleEndian, &idx); err != nil {
		return 0, 0, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "binary.Read"))
	}
	if err = binary.Read(bytes.NewReader(b[8:]), binary.LittleEndian, &shardID); err != nil {
		return 0, 0, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "binary.Read"))
	}
	return idx, shardID, nil
}

func (db *db) GetBlockByIndex(idx uint64, shardID byte) (*common.Hash, error) {
	buf := make([]byte, 9)
	binary.LittleEndian.PutUint64(buf, idx)
	buf[8] = shardID
	// {index-shardID}: {blockhash}

	b, err := db.lvdb.Get(buf, nil)
	if err != nil {
		return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	h := new(common.Hash)
	_ = h.SetBytes(b[:])
	return h, nil
}

//StoreIncomingCrossShard which store crossShardHash from which shard has been include in which block height
func (db *db) StoreIncomingCrossShard(shardID byte, crossShardID byte, blkHeight uint64, crossBlkHash *common.Hash) error {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, blkHeight)
	prefix := append([]byte{shardID}, append([]byte{crossShardID}, crossBlkHash[:]...)...)
	// csh-ShardID-CrossShardID-CrossShardBlockHash : ShardBlockHeight
	key := append(crossShardKeyPrefix, prefix...)
	if ok, _ := db.HasValue(key); ok {
		return database.NewDatabaseError(database.BlockExisted, errors.Errorf("block %s already exists"))
	}
	if err := db.lvdb.Put(key, buf, nil); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}

func (db *db) HasIncomingCrossShard(shardID byte, crossShardID byte, crossBlkHash *common.Hash) error {
	prefix := append([]byte{shardID}, append([]byte{crossShardID}, crossBlkHash[:]...)...)
	// csh-ShardID-CrossShardID-CrossShardBlockHash : ShardBlockHeight
	key := append(crossShardKeyPrefix, prefix...)
	if ok, _ := db.HasValue(key); ok {
		return nil
	}
	return database.NewDatabaseError(database.BlockExisted, errors.Errorf("Cross Shard Block doesn't exist"))
}

func (db *db) GetIncomingCrossShard(shardID byte, crossShardID byte, crossBlkHash *common.Hash) (uint64, error) {
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

//StoreOutgoingCrossShard which store crossShardBlk generated from the block's shard to other shard
func (db *db) StoreOutgoingCrossShard(shardID byte, crossShardID byte, blkHeight uint64, crossBlk interface{}) error {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, blkHeight)
	crossBlkHash, ok := crossBlk.(hasher)
	if !ok {
		return database.NewDatabaseError(database.NotImplHashMethod, errors.New("crossBlk must implement Hash() method"))
	}
	var (
		hash = crossBlkHash.Hash()
		key  = append(append(crossShardKeyPrefix, shardID), append(buf, crossShardID)...)
		// csh-ShardID-BlkIdx-CrossShardID : CrossShardBlk
	)
	if ok, _ := db.HasValue(key); ok {
		return database.NewDatabaseError(database.BlockExisted, errors.Errorf("block %s already exists", hash.String()))
	}
	val, err := json.Marshal(crossBlk)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Marshal"))
	}
	if err := db.Put(key, val); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.Put"))
	}
	return nil
}

func (db *db) HasOutgoingCrossShard(shardID byte, crossShardID byte, blkHeight uint64) error {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, blkHeight)
	key := append(append(crossShardKeyPrefix, shardID), append(buf, crossShardID)...)
	// csh-ShardID-BlkIdx-CrossShardID : CrossShardBlk
	if ok, _ := db.HasValue(key); ok {
		return nil
	}
	return database.NewDatabaseError(database.BlockExisted, errors.Errorf("Cross Shard Block doesn't exist"))
}

func (db *db) GetOutgoingCrossShard(shardID byte, crossShardID byte, blkHeight uint64) ([]byte, error) {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, blkHeight)
	key := append(append(crossShardKeyPrefix, shardID), append(buf, crossShardID)...)
	// csh-ShardID-BlkIdx-CrossShardID : CrossShardBlk
	block, err := db.Get(key)
	if err != nil {
		return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	ret := make([]byte, len(block))
	copy(ret, block)
	return ret, nil
}
