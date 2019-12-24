package rawdbv2

import (
	"encoding/binary"
	"github.com/incognitochain/incognito-chain/common"
)

// Header key will be used for light mode in the future
var (
	lastShardBlockKey       = []byte("LastShardBlock")
	lastShardHeaderKey      = []byte("LastShardHeader")
	lastBeaconBlockKey      = []byte("LastBeaconBlock")
	lastBeaconHeaderKey     = []byte("LastBeaconHeader")
	shardBlockHashPrefix    = []byte("s-b-h")
	shardBlockIndexPrefix   = []byte("s-b-i")
	shardHeaderHashPrefix   = []byte("s-h-h")
	shardHeaderIndexPrefix  = []byte("s-h-i")
	beaconBlockHashPrefix   = []byte("b-b-h")
	beaconBlockIndexPrefix  = []byte("b-b-i")
	beaconHeaderHashPrefix  = []byte("b-h-h")
	beaconHeaderIndexPrefix = []byte("b-h-i")
	txHashPrefix            = []byte("tx-h")
	splitter                = []byte("-[-]-")
)

func GetLastShardBlockKey() []byte {
	return lastShardBlockKey
}

func GetLastShardHeaderKey() []byte {
	return lastShardHeaderKey
}

func GetLastBeaconBlockKey() []byte {
	return lastBeaconBlockKey
}

func GetLastBeaconHeaderKey() []byte {
	return lastBeaconHeaderKey
}

func GetShardHeaderHashKey(shardID byte, hash common.Hash) []byte {
	return append(append(shardHeaderHashPrefix, shardID), hash[:]...)
}

func GetShardBlockHashKey(shardID byte, hash common.Hash) []byte {
	return append(append(shardBlockHashPrefix, shardID), hash[:]...)
}

func GetShardBlockPrefixByShardID(shardID byte) []byte {
	return append(shardBlockHashPrefix, shardID)
}

func GetShardHeaderIndexKey(shardID byte, index uint64, hash common.Hash) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, index)
	key := append(shardHeaderIndexPrefix, shardID)
	key = append(key, buf...)
	key = append(key, splitter...)
	return append(key, hash[:]...)
}

func GetShardBlockIndexKey(shardID byte, index uint64, hash common.Hash) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, index)
	key := append(shardBlockIndexPrefix, shardID)
	key = append(key, buf...)
	key = append(key, splitter...)
	return append(key, hash[:]...)
}

func GetShardHeaderIndexPrefix(shardID byte, index uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, index)
	key := append(shardHeaderIndexPrefix, shardID)
	key = append(key, buf...)
	return key
}

func GetShardBlockIndexPrefix(shardID byte, index uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, index)
	key := append(shardBlockIndexPrefix, shardID)
	key = append(key, buf...)
	return key
}

func GetBeaconHeaderHashKey(hash common.Hash) []byte {
	return append(beaconHeaderHashPrefix, hash[:]...)
}

func GetBeaconBlockHashKey(hash common.Hash) []byte {
	return append(beaconBlockHashPrefix, hash[:]...)
}

func GetBeaconHeaderIndexKey(index uint64, hash common.Hash) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, index)
	key := append(beaconHeaderIndexPrefix, buf...)
	key = append(key, splitter...)
	return append(key, hash[:]...)
}

func GetBeaconBlockIndexKey(index uint64, hash common.Hash) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, index)
	key := append(beaconBlockIndexPrefix, buf...)
	key = append(key, splitter...)
	return append(key, hash[:]...)
}

func GetBeaconHeaderIndexPrefix(index uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, index)
	key := append(beaconHeaderIndexPrefix, buf...)
	return key
}

func GetBeaconBlockIndexPrefix(index uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, index)
	key := append(beaconBlockIndexPrefix, buf...)
	return key
}

func GetTransactionHashKey(hash common.Hash) []byte {
	return append(txHashPrefix, hash[:]...)
}
