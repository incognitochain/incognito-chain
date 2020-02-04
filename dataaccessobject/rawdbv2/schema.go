package rawdbv2

import (
	"github.com/incognitochain/incognito-chain/common"
)

// Header key will be used for light mode in the future
var (
	lastShardBlockKey            = []byte("LastShardBlock")
	lastShardHeaderKey           = []byte("LastShardHeader")
	lastBeaconBlockKey           = []byte("LastBeaconBlock")
	lastBeaconHeaderKey          = []byte("LastBeaconHeader")
	beaconBestStatePrefix        = []byte("BeaconBestState")
	shardBestStatePrefix         = []byte("ShardBestState" + string(splitter))
	shardHashToBlockPrefix       = []byte("s-b-h" + string(splitter))
	shardIndexToBlockHashPrefix  = []byte("s-b-i" + string(splitter))
	shardBlockHashToIndexPrefix  = []byte("s-b-H" + string(splitter))
	shardHeaderHashPrefix        = []byte("s-h-h" + string(splitter))
	shardHeaderIndexPrefix       = []byte("s-h-i" + string(splitter))
	beaconHashToBlockPrefix      = []byte("b-b-h" + string(splitter))
	beaconIndexToBlockHashPrefix = []byte("b-b-i" + string(splitter))
	beaconBlockHashToIndexPrefix = []byte("b-b-H" + string(splitter))
	txHashPrefix                 = []byte("tx-h" + string(splitter))
	crossShardNextHeightPrefix   = []byte("c-s-n-h" + string(splitter))
	feeEstimatorPrefix           = []byte("fee-est" + string(splitter))
	rootHashPrefix               = []byte("R-H-")
	consensusStateRoot           = []byte("co" + string(splitter))
	rewardStateRoot              = []byte("re" + string(splitter))
	featureStateRoot             = []byte("fe" + string(splitter))
	slashStateRoot               = []byte("sl" + string(splitter))
	splitter                     = []byte("-[-]-")
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

// ============================= Shard =======================================
func GetShardHashToHeaderKey(shardID byte, hash common.Hash) []byte {
	return append(append(shardHeaderHashPrefix, shardID), hash[:]...)
}

func GetShardHashToBlockKey(hash common.Hash) []byte {
	return append(shardHashToBlockPrefix, hash[:]...)
}

func GetShardHeaderIndexKey(shardID byte, index uint64, hash common.Hash) []byte {
	buf := common.Uint64ToBytes(index)
	key := append(shardHeaderIndexPrefix, shardID)
	key = append(key, splitter...)
	key = append(key, buf...)
	key = append(key, splitter...)
	return append(key, hash[:]...)
}

func GetShardHeaderIndexPrefix(shardID byte, index uint64) []byte {
	buf := common.Uint64ToBytes(index)
	key := append(shardHeaderIndexPrefix, shardID)
	key = append(key, buf...)
	return key
}

func GetShardIndexToBlockHashKey(shardID byte, index uint64, hash common.Hash) []byte {
	buf := common.Uint64ToBytes(index)
	key := append(shardIndexToBlockHashPrefix, shardID)
	key = append(key, splitter...)
	key = append(key, buf...)
	key = append(key, splitter...)
	return append(key, hash[:]...)
}

func GetShardIndexToBlockHashPrefix(shardID byte, index uint64) []byte {
	buf := common.Uint64ToBytes(index)
	key := append(shardIndexToBlockHashPrefix, shardID)
	key = append(key, splitter...)
	key = append(key, buf...)
	return key
}

func GetShardBlockHashToIndexKey(hash common.Hash) []byte {
	return append(shardBlockHashToIndexPrefix, hash[:]...)
}

func GetShardBestStateKey(shardID byte) []byte {
	return append(shardBestStatePrefix, shardID)
}

// ============================= BEACON =======================================
func GetBeaconHashToBlockKey(hash common.Hash) []byte {
	return append(beaconHashToBlockPrefix, hash[:]...)
}

func GetBeaconIndexToBlockHashKey(index uint64, hash common.Hash) []byte {
	buf := common.Uint64ToBytes(index)
	key := append(beaconIndexToBlockHashPrefix, buf...)
	key = append(key, splitter...)
	return append(key, hash[:]...)
}

func GetBeaconIndexToBlockHashPrefix(index uint64) []byte {
	buf := common.Uint64ToBytes(index)
	key := append(beaconIndexToBlockHashPrefix, buf...)
	return key
}

func GetBeaconBlockHashToIndexKey(hash common.Hash) []byte {
	return append(beaconBlockHashToIndexPrefix, hash[:]...)
}

func GetBeaconBestStateKey() []byte {
	return beaconBestStatePrefix
}

// ============================= Transaction =======================================
func GetTransactionHashKey(hash common.Hash) []byte {
	return append(txHashPrefix, hash[:]...)
}
func GetFeeEstimatorPrefix(shardID byte) []byte {
	return append(feeEstimatorPrefix, shardID)
}

// ============================= Cross Shard =======================================
func GetCrossShardNextHeightKey(fromShard byte, toShard byte, height uint64) []byte {
	buf := common.Uint64ToBytes(height)
	key := append(crossShardNextHeightPrefix, fromShard)
	key = append(key, []byte("-")...)
	key = append(key, toShard)
	key = append(key, []byte("-")...)
	key = append(key, buf...)
	return key
}

// ============================= State Root =======================================
func GetConsensusStateRootHashKey(height uint64) []byte {
	buf := common.Uint64ToBytes(height)
	key := append(rootHashPrefix, consensusStateRoot...)
	key = append(key, buf...)
	return key
}

func GetRewardStateRootHashKey(height uint64) []byte {
	buf := common.Uint64ToBytes(height)
	key := append(rootHashPrefix, rewardStateRoot...)
	key = append(key, buf...)
	return key
}

func GetFeatureStateRootHashKey(height uint64) []byte {
	buf := common.Uint64ToBytes(height)
	key := append(rootHashPrefix, featureStateRoot...)
	key = append(key, buf...)
	return key
}

func GetSlashStateRootHashKey(height uint64) []byte {
	buf := common.Uint64ToBytes(height)
	key := append(rootHashPrefix, slashStateRoot...)
	key = append(key, buf...)
	return key
}
