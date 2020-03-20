package rawdbv2

import (
	"github.com/incognitochain/incognito-chain/common"
)

// Header key will be used for light mode in the future
var (
	lastShardBlockKey                  = []byte("LastShardBlock")
	lastShardHeaderKey                 = []byte("LastShardHeader")
	lastBeaconBlockKey                 = []byte("LastBeaconBlock")
	lastBeaconHeaderKey                = []byte("LastBeaconHeader")
	beaconBestStatePrefix              = []byte("BeaconBestState")
	shardBestStatePrefix               = []byte("ShardBestState" + string(splitter))
	shardHashToBlockPrefix             = []byte("s-b-h" + string(splitter))
	shardIndexToBlockHashPrefix        = []byte("s-b-i" + string(splitter))
	shardBlockHashToIndexPrefix        = []byte("s-b-H" + string(splitter))
	shardHeaderHashPrefix              = []byte("s-h-h" + string(splitter))
	shardHeaderIndexPrefix             = []byte("s-h-i" + string(splitter))
	beaconHashToBlockPrefix            = []byte("b-b-h" + string(splitter))
	beaconIndexToBlockHashPrefix       = []byte("b-b-i" + string(splitter))
	beaconBlockHashToIndexPrefix       = []byte("b-b-H" + string(splitter))
	txHashPrefix                       = []byte("tx-h" + string(splitter))
	crossShardNextHeightPrefix         = []byte("c-s-n-h" + string(splitter))
	feeEstimatorPrefix                 = []byte("fee-est" + string(splitter))
	txByPublicKeyPrefix                = []byte("tx-pb")
	rootHashPrefix                     = []byte("R-H-")
	beaconConsensusRootHashPrefix      = []byte("b-co" + string(splitter))
	beaconRewardRequestRootHashPrefix  = []byte("b-re" + string(splitter))
	beaconFeatureRootHashPrefix        = []byte("b-fe" + string(splitter))
	beaconSlashRootHashPrefix          = []byte("b-sl" + string(splitter))
	shardCommitteeRewardRootHashPrefix = []byte("s-cr" + string(splitter))
	shardConsensusRootHashPrefix       = []byte("s-co" + string(splitter))
	shardTransactionRootHashPrefix     = []byte("s-tx" + string(splitter))
	shardSlashRootHashPrefix           = []byte("s-sl" + string(splitter))
	shardFeatureRootHashPrefix         = []byte("s-fe" + string(splitter))
	previousBestStatePrefix            = []byte("previous-best-state" + string(splitter))
	splitter                           = []byte("-[-]-")
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

func GetStoreTxByPublicKey(publicKey []byte, txID common.Hash, shardID byte) []byte {
	key := append(txByPublicKeyPrefix, publicKey...)
	key = append(key, txID.GetBytes()...)
	key = append(key, shardID)
	return key
}

func GetStoreTxByPublicPrefix(publicKey []byte) []byte {
	return append(txByPublicKeyPrefix, publicKey...)
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
func GetBeaconConsensusRootHashKey(height uint64) []byte {
	buf := common.Uint64ToBytes(height)
	key := append(rootHashPrefix, beaconConsensusRootHashPrefix...)
	key = append(key, buf...)
	return key
}

func GetBeaconRewardRootHashKey(height uint64) []byte {
	buf := common.Uint64ToBytes(height)
	key := append(rootHashPrefix, beaconRewardRequestRootHashPrefix...)
	key = append(key, buf...)
	return key
}

func GetBeaconFeatureRootHashKey(height uint64) []byte {
	buf := common.Uint64ToBytes(height)
	key := append(rootHashPrefix, beaconFeatureRootHashPrefix...)
	key = append(key, buf...)
	return key
}

func GetBeaconSlashRootHashKey(height uint64) []byte {
	buf := common.Uint64ToBytes(height)
	key := append(rootHashPrefix, beaconSlashRootHashPrefix...)
	key = append(key, buf...)
	return key
}

func GetShardCommitteeRewardRootHashKey(shardID byte, height uint64) []byte {
	buf := common.Uint64ToBytes(height)
	key := append(rootHashPrefix, shardCommitteeRewardRootHashPrefix...)
	key = append(key, shardID)
	key = append(key, splitter...)
	key = append(key, buf...)
	return key
}

func GetShardConsensusRootHashKey(shardID byte, height uint64) []byte {
	buf := common.Uint64ToBytes(height)
	key := append(rootHashPrefix, shardConsensusRootHashPrefix...)
	key = append(key, shardID)
	key = append(key, splitter...)
	key = append(key, buf...)
	return key
}

func GetShardTransactionRootHashKey(shardID byte, height uint64) []byte {
	buf := common.Uint64ToBytes(height)
	key := append(rootHashPrefix, shardTransactionRootHashPrefix...)
	key = append(key, shardID)
	key = append(key, splitter...)
	key = append(key, buf...)
	return key
}

func GetShardSlashRootHashKey(shardID byte, height uint64) []byte {
	buf := common.Uint64ToBytes(height)
	key := append(rootHashPrefix, shardSlashRootHashPrefix...)
	key = append(key, shardID)
	key = append(key, splitter...)
	key = append(key, buf...)
	return key
}

func GetShardFeatureRootHashKey(shardID byte, height uint64) []byte {
	buf := common.Uint64ToBytes(height)
	key := append(rootHashPrefix, shardFeatureRootHashPrefix...)
	key = append(key, shardID)
	key = append(key, splitter...)
	key = append(key, buf...)
	return key
}

func GetPreviousBestStateKey(shardID int) []byte {
	return append(previousBestStatePrefix, byte(shardID))
}
