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

func GetShardHeaderHashKey(hash common.Hash) []byte {
	return append(shardHeaderHashPrefix, hash[:]...)
}

func GetShardBlockHashKey(hash common.Hash) []byte {
	return append(shardBlockHashPrefix, hash[:]...)
}

func GetShardHeaderIndexKey(index uint64, hash common.Hash) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, index)
	key := append(shardHeaderIndexPrefix, buf...)
	key = append(key, splitter...)
	return append(key, hash[:]...)
}

func GetShardBlockIndexKey(index uint64, hash common.Hash) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, index)
	key := append(shardBlockIndexPrefix, buf...)
	key = append(key, splitter...)
	return append(key, hash[:]...)
}

func GetShardHeaderIndexPrefix(index uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, index)
	key := append(shardHeaderIndexPrefix, buf...)
	return key
}

func GetShardBlockIndexPrefix(index uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, index)
	key := append(shardBlockIndexPrefix, buf...)
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

func GetTxHashKey(hash common.Hash) []byte {
	return append(txHashPrefix, hash[:]...)
}
