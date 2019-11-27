package rawdb

import (
	"bytes"
	"encoding/binary"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/pkg/errors"
)

func StoreCrossShardNextHeight(db incdb.Database, fromShard byte, toShard byte, curHeight uint64, nextHeight uint64) error {
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
		return NewRawdbError(StoreCrossShardNextHeightError, err)
	}
	return nil
}

func HasCrossShardNextHeight(db incdb.Database, key []byte) (bool, error) {
	exist, err := db.Has(key)
	if err != nil {
		return false, NewRawdbError(HasCrossShardNextHeightError, err)
	} else {
		return exist, nil
	}
}

func FetchCrossShardNextHeight(db incdb.Database, fromShard byte, toShard byte, curHeight uint64) (uint64, error) {
	//ncsh-{fromShard}-{toShard}-{curHeight} = nextHeight
	key := append(nextCrossShardKeyPrefix, fromShard)
	key = append(key, []byte("-")...)
	key = append(key, toShard)
	key = append(key, []byte("-")...)
	curHeightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(curHeightBytes, curHeight)
	key = append(key, curHeightBytes...)
	if _, err := HasCrossShardNextHeight(db, key); err != nil {
		return 0, NewRawdbError(FetchCrossShardNextHeightError, err)
	}
	info, err := db.Get(key)
	if err != nil {
		return 0, NewRawdbError(FetchCrossShardNextHeightError, err)
	}
	var nextHeight uint64
	err = binary.Read(bytes.NewReader(info[:8]), binary.LittleEndian, &nextHeight)
	if err != nil {
		return 0, NewRawdbError(BinaryReaderError, err)
	}
	return nextHeight, nil
}

//StoreIncomingCrossShard which store crossShardHash from which shard has been include in which block height
func StoreIncomingCrossShard(db incdb.Database, shardID byte, crossShardID byte, blkHeight uint64, crossBlkHash common.Hash, bd *[]incdb.BatchData) error {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, blkHeight)
	prefix := append([]byte{shardID}, append([]byte{crossShardID}, crossBlkHash[:]...)...)
	// csh-ShardID-CrossShardID-CrossShardBlockHash : ShardBlockHeight
	key := append(crossShardKeyPrefix, prefix...)
	if ok, _ := db.Has(key); ok {
		return NewRawdbError(BlockExisted, errors.Errorf("block %d already exists", blkHeight))
	}
	if bd != nil {
		*bd = append(*bd, incdb.BatchData{key, buf})
		return nil
	}
	if err := db.Put(key, buf); err != nil {
		return NewRawdbError(LvdbPutError, err)
	}
	return nil
}

func HasIncomingCrossShard(db incdb.Database, shardID byte, crossShardID byte, crossBlkHash common.Hash) error {
	prefix := append([]byte{shardID}, append([]byte{crossShardID}, crossBlkHash[:]...)...)
	// csh-ShardID-CrossShardID-CrossShardBlockHash : ShardBlockHeight
	key := append(crossShardKeyPrefix, prefix...)
	if ok, _ := db.Has(key); ok {
		return nil
	}
	return NewRawdbError(BlockExisted, errors.Errorf("Cross Shard Block doesn't exist"))
}

func GetIncomingCrossShard(db incdb.Database, shardID byte, crossShardID byte, crossBlkHash common.Hash) (uint64, error) {
	prefix := append([]byte{shardID}, append([]byte{crossShardID}, crossBlkHash[:]...)...)
	// csh-ShardID-CrossShardID-CrossShardBlockHash : ShardBlockHeight
	key := append(crossShardKeyPrefix, prefix...)
	b, err := db.Get(key)
	if err != nil {
		return 0, NewRawdbError(LvdbGetError, err)
	}
	var idx uint64
	if err := binary.Read(bytes.NewReader(b[:]), binary.LittleEndian, &idx); err != nil {
		return 0, NewRawdbError(BinaryReaderError, err)
	}
	return idx, nil
}
