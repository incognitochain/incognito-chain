package rawdbv2

import (
	"github.com/incognitochain/incognito-chain/common"
)

// Header key will be used for light mode in the future
var (
	lastShardBlockKey                  = []byte("LastShardBlock" + string(splitter))
	lastBeaconBlockKey                 = []byte("LastBeaconBlock")
	beaconViewsPrefix                  = []byte("BeaconViews")
	shardBestStatePrefix               = []byte("ShardViews" + string(splitter))
	shardHashToBlockPrefix             = []byte("s-b-h" + string(splitter))
	viewPrefix                         = []byte("V" + string(splitter))
	shardIndexToBlockHashPrefix        = []byte("s-b-i" + string(splitter))
	beaconConfirmShardBlockPrefix      = []byte("b-c-s" + string(splitter))
	shardBlockHashToIndexPrefix        = []byte("s-b-H" + string(splitter))
	beaconHashToBlockPrefix            = []byte("b-b-h" + string(splitter))
	blockHashToFFIndexPrefix           = []byte("b-h-ff-i" + string(splitter))
	blockHashToValidationDataPrefix    = []byte("b-h-v-d" + string(splitter))
	beaconIndexToBlockHashPrefix       = []byte("b-b-i" + string(splitter))
	beaconBlockHashToIndexPrefix       = []byte("b-b-H" + string(splitter))
	txHashPrefix                       = []byte("tx-h" + string(splitter))
	crossShardNextHeightPrefix         = []byte("c-s-n-h" + string(splitter))
	lastBeaconHeightConfirmCrossShard  = []byte("p-c-c-s" + string(splitter))
	feeEstimatorPrefix                 = []byte("fee-est" + string(splitter))
	txByPublicKeyPrefix                = []byte("tx-pb")
	rootHashPrefix                     = []byte("R-H-")
	shardRootHashPrefix                = []byte("S-R-H-")
	beaconRootHashPrefix               = []byte("B-R-H-")
	beaconConsensusRootHashPrefix      = []byte("b-co" + string(splitter))
	beaconRewardRequestRootHashPrefix  = []byte("b-re" + string(splitter))
	beaconFeatureRootHashPrefix        = []byte("b-fe" + string(splitter))
	beaconSlashRootHashPrefix          = []byte("b-sl" + string(splitter))
	shardCommitteeRewardRootHashPrefix = []byte("s-cr" + string(splitter))
	shardConsensusRootHashPrefix       = []byte("s-co" + string(splitter))
	stakingTxPrefix                    = []byte("st-tx-" + string(splitter))
	shardTransactionRootHashPrefix     = []byte("s-tx" + string(splitter))
	shardSlashRootHashPrefix           = []byte("s-sl" + string(splitter))
	shardFeatureRootHashPrefix         = []byte("s-fe" + string(splitter))
	previousBestStatePrefix            = []byte("previous-best-state" + string(splitter))
	splitter                           = []byte("-[-]-")

	// output coins by OTA key storage (optional)
	// this will use its own separate folder
	reindexedOutputCoinPrefix = []byte("reindexed-output-coin" + string(splitter))
	reindexedKeysPrefix       = []byte("reindexed-key" + string(splitter))
	coinHashKeysPrefix        = []byte("coinhash-key" + string(splitter))
	txByCoinIndexPrefix       = []byte("tx-index" + string(splitter))
	txBySerialNumberPrefix    = []byte("tx-sn" + string(splitter))
	pruneStatusPrefix         = []byte("p-s")

	cacheCommitteeFromBlockPrefix = []byte("c-c-f-b" + string(splitter))
)

func GetLastShardBlockKey(shardID byte) []byte {
	temp := make([]byte, 0, len(lastShardBlockKey))
	temp = append(temp, lastShardBlockKey...)
	return append(temp, shardID)
}

func GetLastBeaconBlockKey() []byte {
	temp := make([]byte, 0, len(lastBeaconBlockKey))
	temp = append(temp, lastBeaconBlockKey...)
	return temp
}

// ============================= View =======================================
func GetViewPrefixWithValue(view common.Hash) []byte {
	temp := make([]byte, 0, len(viewPrefix))
	temp = append(temp, viewPrefix...)
	key := append(temp, view[:]...)
	return append(key, splitter...)
}

func GetViewBeaconKey(view common.Hash, height uint64) []byte {
	key := GetViewPrefixWithValue(view)
	buf := common.Uint64ToBytes(height)
	return append(key, buf...)
}

func GetViewShardKey(view common.Hash, shardID byte, height uint64) []byte {
	key := GetViewPrefixWithValue(view)
	key = append(key, shardID)
	key = append(key, splitter...)
	buf := common.Uint64ToBytes(height)
	return append(key, buf...)
}

// ============================= Shard =======================================
func GetShardHashToBlockKey(hash common.Hash) []byte {
	temp := make([]byte, 0, len(shardHashToBlockPrefix))
	temp = append(temp, shardHashToBlockPrefix...)
	return append(temp, hash[:]...)
}

func GetShardIndexToBlockHashKey(shardID byte, index uint64, hash common.Hash) []byte {
	buf := common.Uint64ToBytes(index)
	temp := make([]byte, 0, len(shardIndexToBlockHashPrefix))
	temp = append(temp, shardIndexToBlockHashPrefix...)
	key := append(temp, shardID)
	key = append(key, splitter...)
	key = append(key, buf...)
	key = append(key, splitter...)
	return append(key, hash[:]...)
}

func GetShardIndexToBlockHashPrefix(shardID byte, index uint64) []byte {
	buf := common.Uint64ToBytes(index)
	temp := make([]byte, 0, len(shardIndexToBlockHashPrefix))
	temp = append(temp, shardIndexToBlockHashPrefix...)
	key := append(temp, shardID)
	key = append(key, splitter...)
	key = append(key, buf...)
	return key
}

func GetShardBlockHashToIndexKey(hash common.Hash) []byte {
	temp := make([]byte, 0, len(shardBlockHashToIndexPrefix))
	temp = append(temp, shardBlockHashToIndexPrefix...)
	return append(temp, hash[:]...)
}

func GetShardBestStateKey(shardID byte) []byte {
	temp := make([]byte, 0, len(shardBestStatePrefix))
	temp = append(temp, shardBestStatePrefix...)
	return append(temp, shardID)
}

//

func GetBlockHashToFFIndexKey(hash common.Hash) []byte {
	temp := make([]byte, 0, len(blockHashToFFIndexPrefix))
	temp = append(temp, blockHashToFFIndexPrefix...)
	return append(temp, hash[:]...)
}

func GetBlockHashToValidationDataKey(hash common.Hash) []byte {
	temp := make([]byte, 0, len(blockHashToValidationDataPrefix))
	temp = append(temp, blockHashToValidationDataPrefix...)
	return append(temp, hash[:]...)
}

// ============================= BEACON =======================================
func GetBeaconHashToBlockKey(hash common.Hash) []byte {
	temp := make([]byte, 0, len(beaconHashToBlockPrefix))
	temp = append(temp, beaconHashToBlockPrefix...)
	return append(temp, hash[:]...)
}

func GetBeaconIndexToBlockHashKey(index uint64) []byte {
	buf := common.Uint64ToBytes(index)
	temp := make([]byte, 0, len(beaconIndexToBlockHashPrefix))
	temp = append(temp, beaconIndexToBlockHashPrefix...)
	key := append(temp, buf...)
	return key
}

func GetBeaconBlockHashToIndexKey(hash common.Hash) []byte {
	temp := make([]byte, 0, len(beaconBlockHashToIndexPrefix))
	temp = append(temp, beaconBlockHashToIndexPrefix...)
	return append(temp, hash[:]...)
}

func GetBeaconViewsKey() []byte {
	temp := make([]byte, 0, len(beaconViewsPrefix))
	temp = append(temp, beaconViewsPrefix...)
	return temp
}

func GetCacheCommitteeFromBlockKey(hash common.Hash, cid int) []byte {
	temp := make([]byte, 0, len(cacheCommitteeFromBlockPrefix))
	temp = append(temp, cacheCommitteeFromBlockPrefix...)
	temp = append(temp, hash[:]...)
	buf := common.Int32ToBytes(int32(cid))
	return append(temp, buf...)
}

// ============================= Transaction =======================================
func GetTransactionHashKey(hash common.Hash) []byte {
	temp := make([]byte, 0, len(txHashPrefix))
	temp = append(temp, txHashPrefix...)
	return append(temp, hash[:]...)
}
func GetFeeEstimatorPrefix(shardID byte) []byte {
	temp := make([]byte, 0, len(feeEstimatorPrefix))
	temp = append(temp, feeEstimatorPrefix...)
	return append(temp, shardID)
}

func GetStoreTxByPublicKey(publicKey []byte, txID common.Hash, shardID byte) []byte {
	temp := make([]byte, 0, len(txByPublicKeyPrefix))
	temp = append(temp, txByPublicKeyPrefix...)
	key := append(temp, publicKey...)
	key = append(key, txID.GetBytes()...)
	key = append(key, shardID)
	return key
}

func GetStoreTxByPublicPrefix(publicKey []byte) []byte {
	temp := make([]byte, 0, len(txByPublicKeyPrefix))
	temp = append(temp, txByPublicKeyPrefix...)
	return append(temp, publicKey...)
}

// ============================= Cross Shard =======================================
func GetCrossShardNextHeightKey(fromShard byte, toShard byte, height uint64) []byte {
	buf := common.Uint64ToBytes(height)
	temp := make([]byte, 0, len(crossShardNextHeightPrefix))
	temp = append(temp, crossShardNextHeightPrefix...)
	key := append(temp, fromShard)
	key = append(key, []byte("-")...)
	key = append(key, toShard)
	key = append(key, []byte("-")...)
	key = append(key, buf...)
	return key
}

// ============================= State Root =======================================
func GetRootHashPrefix() []byte {
	temp := make([]byte, 0, len(rootHashPrefix))
	temp = append(temp, rootHashPrefix...)
	return temp
}
func GetBeaconConsensusRootHashKey(height uint64) []byte {
	buf := common.Uint64ToBytes(height)
	rootHashPrefix := GetRootHashPrefix()
	temp := make([]byte, 0, len(beaconConsensusRootHashPrefix))
	temp = append(temp, beaconConsensusRootHashPrefix...)
	key := append(rootHashPrefix, temp...)
	key = append(key, buf...)
	return key
}

func GetBeaconRewardRootHashKey(height uint64) []byte {
	buf := common.Uint64ToBytes(height)
	rootHashPrefix := GetRootHashPrefix()
	temp := make([]byte, 0, len(beaconRewardRequestRootHashPrefix))
	temp = append(temp, beaconRewardRequestRootHashPrefix...)
	key := append(rootHashPrefix, temp...)
	key = append(key, buf...)
	return key
}

func GetBeaconFeatureRootHashKey(height uint64) []byte {
	buf := common.Uint64ToBytes(height)
	rootHashPrefix := GetRootHashPrefix()
	temp := make([]byte, 0, len(beaconFeatureRootHashPrefix))
	temp = append(temp, beaconFeatureRootHashPrefix...)
	key := append(rootHashPrefix, temp...)
	key = append(key, buf...)
	return key
}

func GetBeaconSlashRootHashKey(height uint64) []byte {
	buf := common.Uint64ToBytes(height)
	rootHashPrefix := GetRootHashPrefix()
	temp := make([]byte, 0, len(beaconSlashRootHashPrefix))
	temp = append(temp, beaconSlashRootHashPrefix...)
	key := append(rootHashPrefix, temp...)
	key = append(key, buf...)
	return key
}

func GetShardCommitteeRewardRootHashKey(shardID byte, height uint64) []byte {
	buf := common.Uint64ToBytes(height)
	rootHashPrefix := GetRootHashPrefix()
	temp := make([]byte, 0, len(shardCommitteeRewardRootHashPrefix))
	temp = append(temp, shardCommitteeRewardRootHashPrefix...)
	key := append(rootHashPrefix, temp...)
	key = append(key, shardID)
	key = append(key, splitter...)
	key = append(key, buf...)
	return key
}

func GetBeaconConfirmShardBlockPrefix(shardID byte, index uint64) []byte {
	buf := common.Uint64ToBytes(index)
	temp := make([]byte, 0, len(beaconConfirmShardBlockPrefix))
	temp = append(temp, beaconConfirmShardBlockPrefix...)
	key := append(temp, shardID)
	key = append(key, splitter...)
	key = append(key, buf...)
	return key
}

func GetShardRootsHashKey(shardID byte, hash common.Hash) []byte {
	temp := make([]byte, 0, len(shardRootHashPrefix))
	temp = append(temp, shardRootHashPrefix...)
	key := append(temp, shardID)
	key = append(key, splitter...)
	key = append(key, hash.Bytes()...)
	return key
}

func GetBeaconRootsHashKey(hash common.Hash) []byte {
	temp := make([]byte, 0, len(beaconRootHashPrefix))
	temp = append(temp, beaconRootHashPrefix...)
	key := append(temp, splitter...)
	key = append(key, hash.Bytes()...)
	return key
}

func GetShardConsensusRootHashKey(shardID byte, height uint64) []byte {
	buf := common.Uint64ToBytes(height)
	rootHashPrefix := GetRootHashPrefix()
	temp := make([]byte, 0, len(shardConsensusRootHashPrefix))
	temp = append(temp, shardConsensusRootHashPrefix...)
	key := append(rootHashPrefix, temp...)
	key = append(key, shardID)
	key = append(key, splitter...)
	key = append(key, buf...)
	return key
}

func GetShardTransactionRootHashKey(shardID byte, height uint64) []byte {
	buf := common.Uint64ToBytes(height)
	rootHashPrefix := GetRootHashPrefix()
	temp := make([]byte, 0, len(shardTransactionRootHashPrefix))
	temp = append(temp, shardTransactionRootHashPrefix...)
	key := append(rootHashPrefix, temp...)
	key = append(key, shardID)
	key = append(key, splitter...)
	key = append(key, buf...)
	return key
}

func GetShardSlashRootHashKey(shardID byte, height uint64) []byte {
	buf := common.Uint64ToBytes(height)
	rootHashPrefix := GetRootHashPrefix()
	temp := make([]byte, 0, len(shardSlashRootHashPrefix))
	temp = append(temp, shardSlashRootHashPrefix...)
	key := append(rootHashPrefix, temp...)
	key = append(key, shardID)
	key = append(key, splitter...)
	key = append(key, buf...)
	return key
}

func GetShardFeatureRootHashKey(shardID byte, height uint64) []byte {
	buf := common.Uint64ToBytes(height)
	rootHashPrefix := GetRootHashPrefix()
	temp := make([]byte, 0, len(shardFeatureRootHashPrefix))
	temp = append(temp, shardFeatureRootHashPrefix...)
	key := append(rootHashPrefix, temp...)
	key = append(key, shardID)
	key = append(key, splitter...)
	key = append(key, buf...)
	return key
}

func GetPreviousBestStateKey(shardID int) []byte {
	temp := make([]byte, 0, len(previousBestStatePrefix))
	temp = append(temp, previousBestStatePrefix...)
	return append(temp, byte(shardID))
}

func GetLastBeaconHeightConfirmCrossShardKey() []byte {
	temp := make([]byte, 0, len(lastBeaconHeightConfirmCrossShard))
	temp = append(temp, lastBeaconHeightConfirmCrossShard...)
	return temp
}

func GetStakingTxKey(shardID byte, hash common.Hash) []byte {
	temp := make([]byte, 0, len(stakingTxPrefix))
	temp = append(temp, stakingTxPrefix...)
	key := append(temp, shardID)
	key = append(key, hash.Bytes()...)
	return key
}

// getBeaconPreCommitteeInfoKey ...
func getBeaconPreCommitteeInfoKey(hash common.Hash) []byte {
	return hash.Bytes()
}

// getShardPreCommitteeInfoKey ...
func getShardPreCommitteeInfoKey(hash common.Hash) []byte {
	return hash.Bytes()
}

// getShardPreCommitteeInfoForShardKey...
func getShardPreCommitteeInfoForShardKey(hash common.Hash) []byte {
	return hash.Bytes()
}

// getShardPendingValidatorsKey ...
func getShardPendingValidatorsKey(hash common.Hash) []byte {
	return hash.Bytes()
}

// ============================= Coin By OTA Key =======================================

const (
	outcoinPrefixHashKeyLength          = 12
	outcoinPrefixKeyLength              = 20
	coinHashPrefixKeyLength             = 20
	txByCoinIndexPrefixHashKeyLength    = 12
	txByCoinIndexPrefixKeyLength        = 20
	txBySerialNumberPrefixHashKeyLength = 12
	txBySerialNumberPrefixKeyLength     = 20
)

func getIndexedOutputCoinPrefix(tokenID common.Hash, shardID byte, publicKey []byte) []byte {
	h := common.HashH(append(reindexedOutputCoinPrefix, append(tokenID[:], append(publicKey, shardID)...)...))
	return h[:][:outcoinPrefixHashKeyLength]
}

func getIndexedKeysPrefix() []byte {
	h := common.HashH(reindexedKeysPrefix)
	return h[:][:outcoinPrefixHashKeyLength]
}

func getCoinHashKeysPrefix() []byte {
	h := common.HashH(coinHashKeysPrefix)
	return h[:][:coinHashPrefixKeyLength]
}

func generateIndexedOutputCoinObjectKey(tokenID common.Hash, shardID byte, publicKey []byte, outputCoin []byte) []byte {
	prefixHash := getIndexedOutputCoinPrefix(tokenID, shardID, publicKey)
	valueHash := common.HashH(outputCoin)
	return append(prefixHash, valueHash[:][:outcoinPrefixKeyLength]...)
}

func generateIndexedOTAKeyObjectKey(theKey []byte) []byte {
	prefixHash := getIndexedKeysPrefix()
	valueHash := common.HashH(theKey)
	return append(prefixHash, valueHash[:][:outcoinPrefixKeyLength]...)
}

func generateCachedCoinHashObjectKey(theCoinHash []byte) []byte {
	prefixHash := getCoinHashKeysPrefix()
	valueHash := common.HashH(theCoinHash)
	return append(prefixHash, valueHash[:][:coinHashPrefixKeyLength]...)
}

func getTxByCoinIndexPrefix() []byte {
	h := common.HashH(txByCoinIndexPrefix)
	return h[:][:txByCoinIndexPrefixHashKeyLength]
}

func generateTxByCoinIndexObjectKey(index []byte, tokenID common.Hash, shardID byte) []byte {
	prefixHash := getTxByCoinIndexPrefix()

	valueToBeHashed := append(index, shardID)
	valueToBeHashed = append(valueToBeHashed, tokenID.Bytes()...)
	valueHash := common.HashH(valueToBeHashed)

	return append(prefixHash, valueHash[:][:txByCoinIndexPrefixKeyLength]...)
}

func getTxBySerialNumberPrefix() []byte {
	h := common.HashH(txBySerialNumberPrefix)
	return h[:][:txByCoinIndexPrefixHashKeyLength]
}

func generateTxBySerialNumberObjectKey(serialNumber []byte, tokenID common.Hash, shardID byte) []byte {
	if tokenID.String() != common.PRVIDStr {
		tokenID = common.ConfidentialAssetID
	}
	prefixHash := getTxBySerialNumberPrefix()

	valueToBeHashed := append(serialNumber, shardID)
	valueToBeHashed = append(valueToBeHashed, tokenID.Bytes()...)
	valueHash := common.HashH(valueToBeHashed)

	return append(prefixHash, valueHash[:][:txBySerialNumberPrefixKeyLength]...)
}

// ============================= State prune =======================================

func GetPruneStatusKey() []byte {
	temp := make([]byte, 0, len(pruneStatusPrefix))
	temp = append(temp, pruneStatusPrefix...)
	return append(temp)
}

func GetShardRootsHashPrefix(shardID byte) []byte {
	temp := make([]byte, 0, len(shardRootHashPrefix))
	temp = append(temp, shardRootHashPrefix...)
	key := append(temp, shardID)
	key = append(key, splitter...)
	return key
}
