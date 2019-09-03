package lvdb

import (
	"bytes"
	"encoding/binary"
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/pkg/errors"
)

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

func (db *db) StoreBeaconBlock(v interface{}, hash common.Hash) error {
	var (
		// b-{hash}
		keyBlockHash = addPrefixToKeyHash(string(blockKeyPrefix), hash)
		// bea-b-{hash}
		keyBeaconBlock = append(append(beaconPrefix, blockKeyPrefix...), hash[:]...)
	)
	if ok, _ := db.HasValue(keyBeaconBlock); ok {
		return database.NewDatabaseError(database.BlockExisted, errors.Errorf("block %+v already exists", hash))
	}
	val, err := json.Marshal(v)
	if err != nil {
		return database.NewDatabaseError(database.StoreBeaconBlockError, err)
	}
	if err := db.Put(keyBeaconBlock, keyBlockHash); err != nil {
		return database.NewDatabaseError(database.StoreBeaconBlockError, err)
	}
	if err := db.Put(keyBlockHash, val); err != nil {
		return database.NewDatabaseError(database.StoreBeaconBlockError, err)
	}
	return nil
}

func (db *db) HasBeaconBlock(hash common.Hash) (bool, error) {
	key := append(append(beaconPrefix, blockKeyPrefix...), hash[:]...)
	_, err := db.HasValue(key)
	if err != nil {
		return false, database.NewDatabaseError(database.HasBeaconBlockError, err)
	} else {
		keyB := append(blockKeyPrefix, hash[:]...)
		existsB, err := db.HasValue(keyB)
		if err != nil {
			return false, database.NewDatabaseError(database.HasBeaconBlockError, err)
		} else {
			return existsB, nil
		}
	}
}

func (db *db) FetchBeaconBlock(hash common.Hash) ([]byte, error) {
	if _, err := db.HasBeaconBlock(hash); err != nil {
		return []byte{}, database.NewDatabaseError(database.FetchBeaconBlockError, err)
	}
	// b-{hash}
	keyBlockHash := addPrefixToKeyHash(string(blockKeyPrefix), hash)
	block, err := db.Get(keyBlockHash)
	if err != nil {
		return nil, database.NewDatabaseError(database.FetchBeaconBlockError, err)
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
		return database.NewDatabaseError(database.StoreBeaconBlockIndexError, err)
	}
	//bea-i-{index}:[hash]
	beaconBuf := append(append(beaconPrefix, blockKeyIdxPrefix...), buf...)
	if err := db.Put(beaconBuf, hash[:]); err != nil {
		return database.NewDatabaseError(database.StoreBeaconBlockIndexError, err)
	}
	return nil
}

func (db *db) GetIndexOfBeaconBlock(hash common.Hash) (uint64, error) {
	key := append(append(beaconPrefix, blockKeyIdxPrefix...), hash[:]...)
	b, err := db.Get(key)
	//{bea-i-[hash]}:index
	if err != nil {
		return 0, database.NewDatabaseError(database.GetIndexOfBeaconBlockError, err, hash.String())
	}

	var idx uint64
	if err := binary.Read(bytes.NewReader(b[:8]), binary.LittleEndian, &idx); err != nil {
		return 0, database.NewDatabaseError(database.GetIndexOfBeaconBlockError, err, hash.String())
	}
	return idx, nil
}

func (db *db) DeleteBeaconBlock(hash common.Hash, idx uint64) error {
	// Delete block
	var (
		// bea-b-{hash}
		keyBeaconBlock = append(append(beaconPrefix, blockKeyPrefix...), hash[:]...)
		// b-{hash}
		keyBlockHash = addPrefixToKeyHash(string(blockKeyPrefix), hash)
	)
	err := db.Delete(keyBeaconBlock)
	if err != nil {
		return database.NewDatabaseError(database.DeleteBeaconBlockError, err, hash.String(), idx)
	}
	err = db.Delete(keyBlockHash)
	if err != nil {
		return database.NewDatabaseError(database.DeleteBeaconBlockError, err, hash.String(), idx)
	}

	// delete by index
	// bea-i-{hash} -> index
	keyIndex := append(append(beaconPrefix, blockKeyIdxPrefix...), hash[:]...)
	err = db.Delete(keyIndex)
	if err != nil {
		return database.NewDatabaseError(database.DeleteBeaconBlockError, err, hash.String(), idx)
	}

	// index -> {hash}
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, idx)
	err = db.Delete(buf)
	if err != nil {
		return database.NewDatabaseError(database.DeleteBeaconBlockError, err, hash.String(), idx)
	}
	return nil
}

func (db *db) StoreBeaconBestState(v interface{}) error {
	val, err := json.Marshal(v)
	if err != nil {
		return database.NewDatabaseError(database.StoreBeaconBestStateError, err)
	}
	key := beaconBestBlockkeyPrefix
	if err := db.Put(key, val); err != nil {
		return database.NewDatabaseError(database.StoreBeaconBestStateError, err)
	}
	return nil
}

func (db *db) FetchBeaconBestState() ([]byte, error) {
	key := beaconBestBlockkeyPrefix
	block, err := db.Get(key)
	if err != nil {
		return nil, database.NewDatabaseError(database.FetchBeaconBestStateError, err)
	}
	return block, nil
}

func (db *db) CleanBeaconBestState() error {
	key := beaconBestBlockkeyPrefix
	err := db.Delete(key)
	if err != nil {
		return database.NewDatabaseError(database.CleanBeaconBestStateError, err)
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
		return common.Hash{}, database.NewDatabaseError(database.GetBeaconBlockHashByIndexError, err, idx)
	}
	h := new(common.Hash)
	err = h.SetBytes(b[:])
	if err != nil {
		return common.Hash{}, database.NewDatabaseError(database.GetBeaconBlockHashByIndexError, err, idx)
	}
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
		return database.NewDatabaseError(database.StoreAcceptedShardToBeaconError, err)
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
		return 0, database.NewDatabaseError(database.GetAcceptedShardToBeaconError, err)
	}
	var idx uint64
	if err := binary.Read(bytes.NewReader(b[:]), binary.LittleEndian, &idx); err != nil {
		return 0, database.NewDatabaseError(database.GetAcceptedShardToBeaconError, err)
	}
	return idx, nil
}

func (db *db) StoreBeaconCommitteeByHeight(height uint64, v interface{}) error {
	//key: bea-com-ep-{height}
	//value: all shard committee
	key := append(beaconPrefix)
	key = append(key, committeePrefix...)
	key = append(key, heightPrefix...)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, height)
	key = append(key, buf[:]...)

	val, err := json.Marshal(v)
	if err != nil {
		return database.NewDatabaseError(database.StoreBeaconCommitteeByHeightError, err)
	}

	if err := db.lvdb.Put(key, val, nil); err != nil {
		return database.NewDatabaseError(database.StoreBeaconCommitteeByHeightError, err)
	}
	return nil
}

func (db *db) StoreRewardReceiverByHeight(height uint64, v interface{}) error {
	//key: bea-rewardreceiver-ep-{height}
	//value: all reward receiver payment address
	key := append(beaconPrefix)
	key = append(key, rewardReceiverPrefix...)
	key = append(key, heightPrefix...)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, height)
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

func (db *db) StoreShardCommitteeByHeight(height uint64, v interface{}) error {
	//key: bea-s-com-ep-{height}
	//value: all shard committee
	key := append(beaconPrefix, shardIDPrefix...)
	key = append(key, committeePrefix...)
	key = append(key, heightPrefix...)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, height)
	key = append(key, buf[:]...)

	val, err := json.Marshal(v)
	if err != nil {
		return database.NewDatabaseError(database.StoreShardCommitteeByHeightError, err)
	}

	if err := db.Put(key, val); err != nil {
		return database.NewDatabaseError(database.StoreShardCommitteeByHeightError, err)
	}
	return nil
}

func (db *db) FetchShardCommitteeByHeight(height uint64) ([]byte, error) {
	key := append(beaconPrefix, shardIDPrefix...)
	key = append(key, committeePrefix...)
	key = append(key, heightPrefix...)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, height)
	key = append(key, buf[:]...)

	b, err := db.Get(key)
	if err != nil {
		return nil, database.NewDatabaseError(database.FetchShardCommitteeByHeightError, err)
	}
	return b, nil
}

func (db *db) FetchBeaconCommitteeByHeight(height uint64) ([]byte, error) {
	key := append(beaconPrefix)
	key = append(key, committeePrefix...)
	key = append(key, heightPrefix...)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, height)
	key = append(key, buf[:]...)

	b, err := db.lvdb.Get(key, nil)
	if err != nil {
		return nil, database.NewDatabaseError(database.FetchBeaconCommitteeByHeightError, err)
	}
	return b, nil
}

func (db *db) FetchRewardReceiverByHeight(height uint64) ([]byte, error) {
	//key: bea-rewardreceiver-ep-{height}
	//value: all reward receiver payment address
	key := append(beaconPrefix)
	key = append(key, rewardReceiverPrefix...)
	key = append(key, heightPrefix...)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, height)
	key = append(key, buf[:]...)

	b, err := db.lvdb.Get(key, nil)
	if err != nil {
		return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.get"))
	}
	return b, nil
}

func (db *db) HasCommitteeByHeight(height uint64) (bool, error) {
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

func (db *db) StoreAutoStakingByHeight(height uint64, v interface{}) error {
	//key: bea-aust-ep-{height}
	//value: auto staking: map[string]bool
	key := append(beaconPrefix, autoStakingPrefix...)
	key = append(key, heightPrefix...)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, height)
	key = append(key, buf[:]...)

	val, err := json.Marshal(v)
	if err != nil {
		return database.NewDatabaseError(database.StoreAutoStakingByHeightError, err)
	}

	if err := db.Put(key, val); err != nil {
		return database.NewDatabaseError(database.StoreAutoStakingByHeightError, err)
	}
	return nil
}

func (db *db) FetchAutoStakingByHeight(height uint64) ([]byte, error) {
	//key: bea-aust-ep-{height}
	//value: auto staking: map[string]bool
	key := append(beaconPrefix, autoStakingPrefix...)
	key = append(key, heightPrefix...)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, height)
	key = append(key, buf[:]...)

	b, err := db.Get(key)
	if err != nil {
		return nil, database.NewDatabaseError(database.FetchAutoStakingByHeightError, err)
	}
	return b, nil
}
