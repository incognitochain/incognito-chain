package lvdb

import (
	"bytes"
	"encoding/binary"
	"encoding/json"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func (db *db) StoreCrossShardNextHeight(fromShard, toShard byte, curHeight uint64, nextHeight uint64) error {
	//ncsh-{fromShard}-{toShard}-{curHeight} = {nextHeight, nextHash}
	key := append(nextCrossShardKeyPrefix, fromShard)
	key = append(key, []byte("-")...)
	key = append(key, toShard)
	key = append(key, []byte("-")...)
	curHeightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(curHeightBytes, curHeight)
	key = append(key, curHeightBytes...)

	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, nextHeight)

	if err := db.Put(key, buf); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "Cannot store cross shard next height"))
	}

	return nil
}

func (db *db) HasCrossShardNextHeight(key []byte) (bool, error) {
	exist, err := db.HasValue(key)
	if err != nil {
		return false, err
	} else {
		return exist, nil
	}
}

func (db *db) FetchCrossShardNextHeight(fromShard, toShard byte, curHeight uint64) (uint64, error) {
	//ncsh-{fromShard}-{toShard}-{curHeight} = {nextHeight, nextHash}
	key := append(nextCrossShardKeyPrefix, fromShard)
	key = append(key, []byte("-")...)
	key = append(key, toShard)
	key = append(key, []byte("-")...)
	curHeightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(curHeightBytes, curHeight)
	key = append(key, curHeightBytes...)

	if _, err := db.HasCrossShardNextHeight(key); err != nil {
		return 0, err
	}
	info, err := db.Get(key)
	if err != nil {
		return 0, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	var nextHeight uint64
	binary.Read(bytes.NewReader(info[:8]), binary.LittleEndian, &nextHeight)
	return nextHeight, nil
}

func (db *db) StoreBeaconBlock(v interface{}) error {
	h, ok := v.(hasher)
	if !ok {
		return database.NewDatabaseError(database.NotImplHashMethod, errors.New("v must implement Hash() method"))
	}
	var (
		hash = h.Hash()
		// bea-b-{hash}
		key = append(append(beaconPrefix, blockKeyPrefix...), hash[:]...)
		// b-{hash}
		keyB = append(blockKeyPrefix, hash[:]...)
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

// for lightmode only
func (db *db) StoreBeaconBlockHeader(v interface{}, hash *common.Hash) error {
	var (
		key = append(append(beaconPrefix, blockKeyPrefix...), hash[:]...)

		keyB = append(blockKeyPrefix, hash[:]...)
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

func (db *db) HasBeaconBlock(hash *common.Hash) (bool, error) {
	key := append(append(beaconPrefix, blockKeyPrefix...), hash[:]...)
	_, err := db.HasValue(key)
	if err != nil {
		return false, err
	} else {
		keyB := append(blockKeyPrefix, hash[:]...)
		existsB, err := db.HasValue(keyB)
		if err != nil {
			return false, err
		} else {
			return existsB, nil
		}
	}
}

func (db *db) FetchBeaconBlock(hash *common.Hash) ([]byte, error) {
	if _, err := db.HasBeaconBlock(hash); err != nil {
		return []byte{}, err
	}
	key := append(blockKeyPrefix, hash[:]...)
	block, err := db.Get(key)
	if err != nil {
		return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	ret := make([]byte, len(block))
	copy(ret, block)
	return ret, nil
}

func (db *db) DeleteBeaconBlock(hash *common.Hash, idx uint64) error {
	// Delete block
	// bea-b-{hash}
	key := append(append(beaconPrefix, blockKeyPrefix...), hash[:]...)
	err := db.lvdb.Delete(key, nil)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Delete"))
	}
	// b-{hash}
	keyB := append(blockKeyPrefix, hash[:]...)
	err = db.lvdb.Delete(keyB, nil)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Delete"))
	}

	// delete by index
	// bea-i-{hash} -> index
	keyIndex := append(append(beaconPrefix, blockKeyIdxPrefix...), hash[:]...)
	err = db.lvdb.Delete(keyIndex, nil)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}

	// index -> {hash}
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, idx)
	err = db.lvdb.Delete(buf, nil)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}

	err = db.lvdb.Delete(key, nil)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Delete"))
	}
	return nil
}

func (db *db) StoreBeaconBestState(v interface{}) error {
	val, err := json.Marshal(v)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Marshal"))
	}
	key := beaconBestBlockkey
	if err := db.Put(key, val); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.put"))
	}
	return nil
}

func (db *db) FetchBeaconBestState() ([]byte, error) {
	key := beaconBestBlockkey
	block, err := db.Get(key)
	if err != nil {
		return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.get"))
	}
	return block, nil
}

func (db *db) CleanBeaconBestState() error {
	key := beaconBestBlockkey
	err := db.lvdb.Delete(key, nil)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.delete"))
	}
	return nil
}

func (db *db) StoreBeaconBlockIndex(h *common.Hash, idx uint64) error {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, idx)
	key := append(append(beaconPrefix, blockKeyIdxPrefix...), h[:]...)
	//{bea-i-{hash}}:index
	if err := db.lvdb.Put(key, buf, nil); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}
	//bea-i-{index}:[hash]
	beaconBuf := append(append(beaconPrefix, blockKeyIdxPrefix...), buf...)
	if err := db.lvdb.Put(beaconBuf, h[:], nil); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}

func (db *db) GetIndexOfBeaconBlock(h *common.Hash) (uint64, error) {
	key := append(append(beaconPrefix, blockKeyIdxPrefix...), h[:]...)
	b, err := db.lvdb.Get(key, nil)
	//{bea-i-[hash]}:index
	if err != nil {
		return 0, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.get"))
	}

	var idx uint64
	if err := binary.Read(bytes.NewReader(b[:8]), binary.LittleEndian, &idx); err != nil {
		return 0, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "binary.Read"))
	}
	return idx, nil
}

func (db *db) GetBeaconBlockHashByIndex(idx uint64) (*common.Hash, error) {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, idx)
	//bea-i-{index}:[hash]
	beaconBuf := append(append(beaconPrefix, blockKeyIdxPrefix...), buf...)
	b, err := db.lvdb.Get(beaconBuf, nil)
	if err != nil {
		return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	h := new(common.Hash)
	_ = h.SetBytes(b[:])
	return h, nil
}

func (db *db) FetchBeaconBlockChain() ([]*common.Hash, error) {
	keys := []*common.Hash{}
	prefix := append(beaconPrefix, blockKeyPrefix...)
	// prefix: bea-b-...
	iter := db.lvdb.NewIterator(util.BytesPrefix(prefix), nil)
	for iter.Next() {
		h := new(common.Hash)
		_ = h.SetBytes(iter.Key()[len(prefix):])
		keys = append(keys, h)
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "iter.Error"))
	}
	return keys, nil
}

//StoreCrossShard store which crossShardBlk from which shard has been include in which beacon block height
func (db *db) StoreAcceptedShardToBeacon(shardID byte, blkHeight uint64, shardBlkHash *common.Hash) error {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, blkHeight)
	prefix := append([]byte{shardID}, shardBlkHash[:]...)
	// stb-ShardID-ShardBlockHash : BeaconBlockHeight
	key := append(shardToBeaconKeyPrefix, prefix...)
	if err := db.lvdb.Put(key, buf, nil); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}

func (db *db) HasAcceptedShardToBeacon(shardID byte, shardBlkHash *common.Hash) error {
	prefix := append([]byte{shardID}, shardBlkHash[:]...)
	// stb-ShardID-ShardBlockHash : BeaconBlockHeight
	key := append(shardToBeaconKeyPrefix, prefix...)
	if ok, _ := db.HasValue(key); ok {
		return nil
	}
	return database.NewDatabaseError(database.BlockExisted, errors.Errorf("Cross Shard Block doesn't exist"))
}

func (db *db) GetAcceptedShardToBeacon(shardID byte, shardBlkHash *common.Hash) (uint64, error) {
	prefix := append([]byte{shardID}, shardBlkHash[:]...)
	// stb-ShardID-ShardBlockHash : BeaconBlockHeight
	key := append(shardToBeaconKeyPrefix, prefix...)
	b, err := db.Get(key)
	if err != nil {
		return 0, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.get"))
	}
	var idx uint64
	if err := binary.Read(bytes.NewReader(b[:]), binary.LittleEndian, &idx); err != nil {
		return 0, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "binary.Read"))
	}
	return idx, nil
}

func (db *db) StoreCommitteeByHeight(blkHeight uint64, v interface{}) error {
	//key: bea-s-com-{height}
	//value: all shard committee
	key := append(beaconPrefix, shardIDPrefix...)
	key = append(key, committeePrefix...)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, blkHeight)
	key = append(key, buf[:]...)

	val, err := json.Marshal(v)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Marshal"))
	}

	if err := db.lvdb.Put(key, val, nil); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}

func (db *db) FetchCommitteeByHeight(blkHeight uint64) ([]byte, error) {
	key := append(beaconPrefix, shardIDPrefix...)
	key = append(key, committeePrefix...)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, blkHeight)
	key = append(key, buf[:]...)

	b, err := db.lvdb.Get(key, nil)
	if err != nil {
		return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.get"))
	}
	return b, nil
}

func (db *db) StoreCommitteeByEpoch(blkEpoch uint64, v interface{}) error {
	//key: bea-s-com-ep-{epoch}
	//value: all shard committee
	key := append(beaconPrefix, shardIDPrefix...)
	key = append(key, committeePrefix...)
	key = append(key, epochPrefix...)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, blkEpoch)
	key = append(key, buf[:]...)

	val, err := json.Marshal(v)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Marshal"))
	}

	if err := db.lvdb.Put(key, val, nil); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}

func (db *db) FetchCommitteeByEpoch(blkEpoch uint64) ([]byte, error) {
	key := append(beaconPrefix, shardIDPrefix...)
	key = append(key, committeePrefix...)
	key = append(key, epochPrefix...)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, blkEpoch)
	key = append(key, buf[:]...)

	b, err := db.lvdb.Get(key, nil)
	if err != nil {
		return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.get"))
	}
	return b, nil
}
func (db *db) HasCommitteeByEpoch(blkEpoch uint64) (bool, error) {
	key := append(beaconPrefix, shardIDPrefix...)
	key = append(key, committeePrefix...)
	key = append(key, epochPrefix...)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, blkEpoch)
	key = append(key, buf[:]...)

	exist, err := db.lvdb.Has(key, nil)
	if err != nil {
		return false, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.get"))
	}
	return exist, nil
}
