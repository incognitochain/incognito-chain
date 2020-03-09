package lvdb

import (
	"bytes"
	"encoding/binary"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/pkg/errors"

	"github.com/incognitochain/incognito-chain/database"
)

func (db *db) StoreBeaconHashConfirmCrossShardHeight(fromShard, toShard byte, height uint64, beaconHash string) error {
	//beah-cx-height-{fromShard}-{toShard}-{curHeight} = beaconHash
	key := append(beaconHashConfirmCrossShardHeightKeyPrefix, fromShard)
	key = append(key, []byte("-")...)
	key = append(key, toShard)
	key = append(key, []byte("-")...)
	curHeightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(curHeightBytes, height)
	key = append(key, curHeightBytes...)

	if err := db.Put(key, []byte(beaconHash)); err != nil {
		return database.NewDatabaseError(database.StoreBeaconHashConfirmCrossShardHeight, err)
	}
	return nil
}

func (db *db) FetchBeaconHashConfirmCrossShardHeight(fromShard, toShard byte, height uint64) ([]byte, error) {
	//beah-cx-height-{fromShard}-{toShard}-{curHeight} = beaconHash
	key := append(beaconHashConfirmCrossShardHeightKeyPrefix, fromShard)
	key = append(key, []byte("-")...)
	key = append(key, toShard)
	key = append(key, []byte("-")...)
	curHeightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(curHeightBytes, height)
	key = append(key, curHeightBytes...)

	info, err := db.Get(key)
	if err != nil {
		return nil, database.NewDatabaseError(database.FetchBeaconHashConfirmCrossShardHeight, err)
	}

	h, err := common.Hash{}.NewHashFromStr(string(info))
	if err != nil {
		return nil, database.NewDatabaseError(database.FetchBeaconHashConfirmCrossShardHeight, err)
	}

	return db.FetchBlock(*h)
}

func (db *db) StoreCrossShardNextHeight(fromShard byte, toShard byte, curHeight uint64, nextHeight uint64) error {
	//ncsh-{fromShard}-{toShard}-{curHeight} = nextHeight
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
		return database.NewDatabaseError(database.StoreCrossShardNextHeightError, err)
	}
	return nil
}

func (db *db) HasCrossShardNextHeight(key []byte) (bool, error) {
	exist, err := db.HasValue(key)
	if err != nil {
		return false, database.NewDatabaseError(database.HasCrossShardNextHeightError, err)
	} else {
		return exist, nil
	}
}

func (db *db) FetchCrossShardNextHeight(fromShard byte, toShard byte, curHeight uint64) (uint64, error) {
	//ncsh-{fromShard}-{toShard}-{curHeight} = nextHeight
	key := append(nextCrossShardKeyPrefix, fromShard)
	key = append(key, []byte("-")...)
	key = append(key, toShard)
	key = append(key, []byte("-")...)
	curHeightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(curHeightBytes, curHeight)
	key = append(key, curHeightBytes...)

	if _, err := db.HasCrossShardNextHeight(key); err != nil {
		return 0, database.NewDatabaseError(database.FetchCrossShardNextHeightError, err)
	}
	info, err := db.Get(key)
	if err != nil {
		return 0, database.NewDatabaseError(database.FetchCrossShardNextHeightError, err)
	}
	var nextHeight uint64
	err = binary.Read(bytes.NewReader(info[:8]), binary.LittleEndian, &nextHeight)
	return nextHeight, err
}

//StoreIncomingCrossShard which store crossShardHash from which shard has been include in which block height
func (db *db) StoreIncomingCrossShard(shardID byte, crossShardID byte, blkHeight uint64, crossBlkHash common.Hash, bd *[]database.BatchData) error {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, blkHeight)
	prefix := append([]byte{shardID}, append([]byte{crossShardID}, crossBlkHash[:]...)...)
	// csh-ShardID-CrossShardID-CrossShardBlockHash : ShardBlockHeight
	key := append(crossShardKeyPrefix, prefix...)
	if ok, _ := db.HasValue(key); ok {
		return database.NewDatabaseError(database.BlockExisted, errors.Errorf("block %d already exists", blkHeight))
	}

	if bd != nil {
		*bd = append(*bd, database.BatchData{key, buf})
		return nil
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
