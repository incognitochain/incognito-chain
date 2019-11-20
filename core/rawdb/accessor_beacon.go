package rawdb

import (
	"bytes"
	"encoding/binary"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
)

func StoreBeaconBlock(db incdb.Database, v interface{}, hash common.Hash, bd *[]incdb.BatchData) error {
	var (
		// b-{hash}
		keyBlockHash = prefixWithHashKey(string(blockKeyPrefix), hash)
		// bea-b-{hash}
		keyBeaconBlock = append(append(beaconPrefix, blockKeyPrefix...), hash[:]...)
	)
	if ok, _ := db.Has(keyBeaconBlock); ok {
		return NewRawdbError(BlockExisted, errors.Errorf("block %+v already exists", hash))
	}
	val, err := json.Marshal(v)
	if err != nil {
		return NewRawdbError(StoreBeaconBlockError, err)
	}

	if bd != nil {
		*bd = append(*bd, incdb.BatchData{keyBeaconBlock, keyBlockHash})
		*bd = append(*bd, incdb.BatchData{keyBlockHash, val})
		return nil
	}
	if err := db.Put(keyBeaconBlock, keyBlockHash); err != nil {
		return NewRawdbError(StoreBeaconBlockError, err)
	}
	if err := db.Put(keyBlockHash, val); err != nil {
		return NewRawdbError(StoreBeaconBlockError, err)
	}
	return nil
}

func HasBeaconBlock(db incdb.Database, hash common.Hash) (bool, error) {
	key := append(append(beaconPrefix, blockKeyPrefix...), hash[:]...)
	_, err := db.Has(key)
	if err != nil {
		return false, NewRawdbError(HasBeaconBlockError, err)
	} else {
		keyB := append(blockKeyPrefix, hash[:]...)
		existsB, err := db.Has(keyB)
		if err != nil {
			return false, NewRawdbError(HasBeaconBlockError, err)
		} else {
			return existsB, nil
		}
	}
}

func FetchBeaconBlock(db incdb.Database, hash common.Hash) ([]byte, error) {
	if _, err := HasBeaconBlock(db, hash); err != nil {
		return []byte{}, NewRawdbError(FetchBeaconBlockError, err)
	}
	// b-{hash}
	keyBlockHash := prefixWithHashKey(string(blockKeyPrefix), hash)
	block, err := db.Get(keyBlockHash)
	if err != nil {
		return nil, NewRawdbError(FetchBeaconBlockError, err)
	}
	ret := make([]byte, len(block))
	copy(ret, block)
	return ret, nil
}

func StoreBeaconBlockIndex(db incdb.Database, hash common.Hash, idx uint64) error {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, idx)
	//{bea-i-{hash}}:index
	key := append(append(beaconPrefix, blockKeyIdxPrefix...), hash[:]...)
	if err := db.Put(key, buf); err != nil {
		return NewRawdbError(StoreBeaconBlockIndexError, err)
	}
	//bea-i-{index}:[hash]
	beaconBuf := append(append(beaconPrefix, blockKeyIdxPrefix...), buf...)
	if err := db.Put(beaconBuf, hash[:]); err != nil {
		return NewRawdbError(StoreBeaconBlockIndexError, err)
	}
	return nil
}

func GetIndexOfBeaconBlock(db incdb.Database, hash common.Hash) (uint64, error) {
	key := append(append(beaconPrefix, blockKeyIdxPrefix...), hash[:]...)
	b, err := db.Get(key)
	//{bea-i-[hash]}:index
	if err != nil {
		return 0, NewRawdbError(GetIndexOfBeaconBlockError, err, hash.String())
	}

	var idx uint64
	if err := binary.Read(bytes.NewReader(b[:8]), binary.LittleEndian, &idx); err != nil {
		return 0, NewRawdbError(GetIndexOfBeaconBlockError, err, hash.String())
	}
	return idx, nil
}

func DeleteBeaconBlock(db incdb.Database, hash common.Hash, idx uint64) error {
	// Delete block
	var (
		// bea-b-{hash}
		keyBeaconBlock = append(append(beaconPrefix, blockKeyPrefix...), hash[:]...)
		// b-{hash}
		keyBlockHash = prefixWithHashKey(string(blockKeyPrefix), hash)
	)
	err := db.Delete(keyBeaconBlock)
	if err != nil {
		return NewRawdbError(DeleteBeaconBlockError, err, hash.String(), idx)
	}
	err = db.Delete(keyBlockHash)
	if err != nil {
		return NewRawdbError(DeleteBeaconBlockError, err, hash.String(), idx)
	}

	// delete by index
	// bea-i-{hash} -> index
	keyIndex := append(append(beaconPrefix, blockKeyIdxPrefix...), hash[:]...)
	err = db.Delete(keyIndex)
	if err != nil {
		return NewRawdbError(DeleteBeaconBlockError, err, hash.String(), idx)
	}

	// index -> {hash}
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, idx)
	err = db.Delete(buf)
	if err != nil {
		return NewRawdbError(DeleteBeaconBlockError, err, hash.String(), idx)
	}
	return nil
}

func StoreBeaconBestState(db incdb.Database, v interface{}, bd *[]incdb.BatchData) error {
	val, err := json.Marshal(v)
	if err != nil {
		return NewRawdbError(StoreBeaconBestStateError, err)
	}
	key := beaconBestBlockkeyPrefix

	if bd != nil {
		*bd = append(*bd, incdb.BatchData{key, val})
		return nil
	}
	if err := db.Put(key, val); err != nil {
		return NewRawdbError(StoreBeaconBestStateError, err)
	}
	return nil
}
func FetchBeaconBestState(db incdb.Database) ([]byte, error) {
	key := beaconBestBlockkeyPrefix
	block, err := db.Get(key)
	if err != nil {
		return nil, NewRawdbError(FetchBeaconBestStateError, err)
	}
	return block, nil
}
func CleanBeaconBestState(db incdb.Database) error {
	key := beaconBestBlockkeyPrefix
	err := db.Delete(key)
	if err != nil {
		return NewRawdbError(CleanBeaconBestStateError, err)
	}
	return nil
}

func GetBeaconBlockHashByIndex(db incdb.Database, idx uint64) (common.Hash, error) {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, idx)
	//bea-i-{index}:[hash]
	beaconBuf := append(append(beaconPrefix, blockKeyIdxPrefix...), buf...)
	b, err := db.Get(beaconBuf)
	if err != nil {
		return common.Hash{}, NewRawdbError(GetBeaconBlockHashByIndexError, err, idx)
	}
	h := new(common.Hash)
	err = h.SetBytes(b[:])
	if err != nil {
		return common.Hash{}, NewRawdbError(GetBeaconBlockHashByIndexError, err, idx)
	}
	return *h, nil
}

//StoreCrossShard store which crossShardBlk from which shard has been include in which beacon block height
func StoreAcceptedShardToBeacon(db incdb.Database, shardID byte, blkHeight uint64, shardBlkHash common.Hash) error {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, blkHeight)
	prefix := append([]byte{shardID}, shardBlkHash[:]...)
	// stb-ShardID-ShardBlockHash : BeaconBlockHeight
	key := append(shardToBeaconKeyPrefix, prefix...)
	if err := db.Put(key, buf); err != nil {
		return NewRawdbError(StoreAcceptedShardToBeaconError, err)
	}
	return nil
}

func HasAcceptedShardToBeacon(db incdb.Database, shardID byte, shardBlkHash common.Hash) error {
	prefix := append([]byte{shardID}, shardBlkHash[:]...)
	// stb-ShardID-ShardBlockHash : BeaconBlockHeight
	key := append(shardToBeaconKeyPrefix, prefix...)
	if ok, _ := db.Has(key); ok {
		return nil
	}
	return NewRawdbError(BlockExisted, errors.Errorf("Cross Shard Block doesn't exist"))
}

func GetAcceptedShardToBeacon(db incdb.Database, shardID byte, shardBlkHash common.Hash) (uint64, error) {
	prefix := append([]byte{shardID}, shardBlkHash[:]...)
	// stb-ShardID-ShardBlockHash : BeaconBlockHeight
	key := append(shardToBeaconKeyPrefix, prefix...)
	b, err := db.Get(key)
	if err != nil {
		return 0, NewRawdbError(GetAcceptedShardToBeaconError, err)
	}
	var idx uint64
	if err := binary.Read(bytes.NewReader(b[:]), binary.LittleEndian, &idx); err != nil {
		return 0, NewRawdbError(GetAcceptedShardToBeaconError, err)
	}
	return idx, nil
}

func StoreBeaconCommitteeByHeight(db incdb.Database, height uint64, v interface{}) error {
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
		return NewRawdbError(StoreBeaconCommitteeByHeightError, err)
	}

	if err := db.Put(key, val); err != nil {
		return NewRawdbError(StoreBeaconCommitteeByHeightError, err)
	}
	return nil
}

func StoreRewardReceiverByHeight(db incdb.Database, height uint64, v interface{}) error {
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
		return NewRawdbError(JsonMarshalError, err)
	}
	if err := db.Put(key, val); err != nil {
		return NewRawdbError(LvdbPutError, err)
	}
	return nil
}

func StoreShardCommitteeByHeight(db incdb.Database, height uint64, v interface{}) error {
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
		return NewRawdbError(StoreShardCommitteeByHeightError, err)
	}

	if err := db.Put(key, val); err != nil {
		return NewRawdbError(StoreShardCommitteeByHeightError, err)
	}
	return nil
}

func FetchShardCommitteeByHeight(db incdb.Database, height uint64) ([]byte, error) {
	key := append(beaconPrefix, shardIDPrefix...)
	key = append(key, committeePrefix...)
	key = append(key, heightPrefix...)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, height)
	key = append(key, buf[:]...)

	res, err := db.Get(key)
	if err != nil {
		return nil, NewRawdbError(FetchShardCommitteeByHeightError, err)
	}
	return res, nil
}

func FetchBeaconCommitteeByHeight(db incdb.Database, height uint64) ([]byte, error) {
	key := append(beaconPrefix)
	key = append(key, committeePrefix...)
	key = append(key, heightPrefix...)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, height)
	key = append(key, buf[:]...)

	res, err := db.Get(key)
	if err != nil {
		return nil, NewRawdbError(FetchBeaconCommitteeByHeightError, err)
	}
	return res, nil
}

func FetchRewardReceiverByHeight(db incdb.Database, height uint64) ([]byte, error) {
	//key: bea-rewardreceiver-ep-{height}
	//value: all reward receiver payment address
	key := append(beaconPrefix)
	key = append(key, rewardReceiverPrefix...)
	key = append(key, heightPrefix...)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, height)
	key = append(key, buf[:]...)

	res, err := db.Get(key)
	if err != nil {
		return nil, NewRawdbError(LvdbGetError, err)
	}
	return res, nil
}

func StoreAutoStakingByHeight(db incdb.Database, height uint64, v interface{}) error {
	//key: bea-aust-ep-{height}
	//value: auto staking: map[string]bool
	key := append(beaconPrefix, autoStakingPrefix...)
	key = append(key, heightPrefix...)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, height)
	key = append(key, buf[:]...)

	val, err := json.Marshal(v)
	if err != nil {
		return NewRawdbError(StoreAutoStakingByHeightError, err)
	}

	if err := db.Put(key, val); err != nil {
		return NewRawdbError(StoreAutoStakingByHeightError, err)
	}
	return nil
}

func FetchAutoStakingByHeight(db incdb.Database, height uint64) ([]byte, error) {
	//key: bea-aust-ep-{height}
	//value: auto staking: map[string]bool
	key := append(beaconPrefix, autoStakingPrefix...)
	key = append(key, heightPrefix...)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, height)
	key = append(key, buf[:]...)
	res, err := db.Get(key)
	if err != nil {
		return nil, NewRawdbError(FetchAutoStakingByHeightError, err)
	}
	return res, nil
}
