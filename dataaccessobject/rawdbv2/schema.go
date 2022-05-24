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

	PreimagePrefix = []byte("secure-key-") // PreimagePrefix + hash -> preimage
)

// ============================= Shard =======================================
func GetShardHashToBlockKey(hash common.Hash) []byte {
	temp := make([]byte, 0, len(shardHashToBlockPrefix))
	temp = append(temp, shardHashToBlockPrefix...)
	return append(temp, hash[:]...)
}

func GetHashToBlockIndexKey(hash common.Hash) []byte {
	temp := make([]byte, 0, len("blkindex"))
	temp = append(temp, "blkindex"...)
	return append(temp, hash[:]...)
}

func GetHashToBlockValidationKey(hash common.Hash) []byte {
	temp := make([]byte, 0, len("blkval"))
	temp = append(temp, "blkval"...)
	return append(temp, hash[:]...)
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

func GetShardBestStateKey(shardID byte) []byte {
	temp := make([]byte, 0, len(shardBestStatePrefix))
	temp = append(temp, shardBestStatePrefix...)
	return append(temp, shardID)
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

func GetBeaconViewsKey() []byte {
	temp := make([]byte, 0, len(beaconViewsPrefix))
	temp = append(temp, beaconViewsPrefix...)
	return temp
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

func GetLastBeaconHeightConfirmCrossShardKey() []byte {
	temp := make([]byte, 0, len(lastBeaconHeightConfirmCrossShard))
	temp = append(temp, lastBeaconHeightConfirmCrossShard...)
	return temp
}

// ============================= Coin By OTA Key =======================================

const (
	outcoinPrefixHashKeyLength       = 12
	outcoinPrefixKeyLength           = 20
	coinHashPrefixKeyLength          = 20
	txByCoinIndexPrefixHashKeyLength = 12
	txByCoinIndexPrefixKeyLength     = 20
	txBySerialNumberPrefixKeyLength  = 20
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

// preimageKey = PreimagePrefix + hash
func preimageKey(hash common.Hash) []byte {
	temp := make([]byte, len(PreimagePrefix))
	copy(temp, PreimagePrefix)
	return append(temp, hash.Bytes()...)
}
