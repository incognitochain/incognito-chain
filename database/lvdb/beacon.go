package lvdb

import (
	"bytes"
	"encoding/binary"
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/pkg/errors"
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

func (db *db) StoreBeaconBlock(v interface{}, hash common.Hash) error {
	var (
		// b-{hash}
		keyBlockHash = db.GetKey(string(blockKeyPrefix), hash)
		// bea-b-{hash}
		keyBeaconBlock = append(append(beaconPrefix, blockKeyPrefix...), hash[:]...)
	)
	if ok, _ := db.HasValue(keyBeaconBlock); ok {
		return database.NewDatabaseError(database.BlockExisted, errors.Errorf("block %+v already exists", hash))
	}
	val, err := json.Marshal(v)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Marshal"))
	}
	if err := db.Put(keyBeaconBlock, keyBlockHash); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.Put"))
	}
	if err := db.Put(keyBlockHash, val); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.Put"))
	}
	return nil
}

func (db *db) HasBeaconBlock(hash common.Hash) (bool, error) {
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

func (db *db) FetchBeaconBlock(hash common.Hash) ([]byte, error) {
	if _, err := db.HasBeaconBlock(hash); err != nil {
		return []byte{}, err
	}
	// b-{hash}
	keyBlockHash := db.GetKey(string(blockKeyPrefix), hash)
	block, err := db.Get(keyBlockHash)
	if err != nil {
		return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	ret := make([]byte, len(block))
	copy(ret, block)
	return ret, nil
}

func (db *db) StoreBeaconBlockIndex(hash common.Hash, idx uint64) error {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, idx)
	//{bea-i-{hash}}:index
	key := append(append(beaconPrefix, blockKeyIdxPrefix...), hash[:]...)
	if err := db.Put(key, buf); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}
	//bea-i-{index}:[hash]
	beaconBuf := append(append(beaconPrefix, blockKeyIdxPrefix...), buf...)
	if err := db.Put(beaconBuf, hash[:]); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}

func (db *db) GetIndexOfBeaconBlock(hash common.Hash) (uint64, error) {
	key := append(append(beaconPrefix, blockKeyIdxPrefix...), hash[:]...)
	b, err := db.Get(key)
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

func (db *db) DeleteBeaconBlock(hash common.Hash, idx uint64) error {
	// Delete block
	var (
		// bea-b-{hash}
		keyBeaconBlock = append(append(beaconPrefix, blockKeyPrefix...), hash[:]...)
		// b-{hash}
		keyBlockHash = db.GetKey(string(blockKeyPrefix), hash)
	)
	err := db.Delete(keyBeaconBlock)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Delete"))
	}
	err = db.Delete(keyBlockHash)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Delete"))
	}

	// delete by index
	// bea-i-{hash} -> index
	keyIndex := append(append(beaconPrefix, blockKeyIdxPrefix...), hash[:]...)
	err = db.Delete(keyIndex)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}

	// index -> {hash}
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, idx)
	err = db.Delete(buf)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
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
	err := db.Delete(key)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.delete"))
	}
	return nil
}
func (db *db) GetBeaconBlockHashByIndex(idx uint64) (common.Hash, error) {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, idx)
	//bea-i-{index}:[hash]
	beaconBuf := append(append(beaconPrefix, blockKeyIdxPrefix...), buf...)
	b, err := db.Get(beaconBuf)
	if err != nil {
		return common.Hash{}, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	h := new(common.Hash)
	_ = h.SetBytes(b[:])
	return *h, nil
}

//StoreCrossShard store which crossShardBlk from which shard has been include in which beacon block height
func (db *db) StoreAcceptedShardToBeacon(shardID byte, blkHeight uint64, shardBlkHash common.Hash) error {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, blkHeight)
	prefix := append([]byte{shardID}, shardBlkHash[:]...)
	// stb-ShardID-ShardBlockHash : BeaconBlockHeight
	key := append(shardToBeaconKeyPrefix, prefix...)
	if err := db.Put(key, buf); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}

func (db *db) HasAcceptedShardToBeacon(shardID byte, shardBlkHash common.Hash) error {
	prefix := append([]byte{shardID}, shardBlkHash[:]...)
	// stb-ShardID-ShardBlockHash : BeaconBlockHeight
	key := append(shardToBeaconKeyPrefix, prefix...)
	if ok, _ := db.HasValue(key); ok {
		return nil
	}
	return database.NewDatabaseError(database.BlockExisted, errors.Errorf("Cross Shard Block doesn't exist"))
}

func (db *db) GetAcceptedShardToBeacon(shardID byte, shardBlkHash common.Hash) (uint64, error) {
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

/*func (db *db) StoreCommitteeByHeight(blkHeight uint64, v interface{}) error {
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
	if err := db.Put(key, val); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}*/

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

	if err := db.Put(key, val); err != nil {
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

	b, err := db.Get(key)
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

	exist, err := db.HasValue(key)
	if err != nil {
		return false, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.get"))
	}
	return exist, nil
}
