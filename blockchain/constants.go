package blockchain

import (
	"github.com/incognitochain/incognito-chain/common"
	"time"

	"github.com/incognitochain/incognito-chain/metrics"
)

//Network fixed params
const (
	// SHARD_BLOCK_VERSION is the current latest supported block version.
	VERSION                            = 1
	RANDOM_NUMBER                      = 3
	SHARD_BLOCK_VERSION                = 1
	DefaultMaxBlkReqPerPeer            = 900
	MinCommitteeSize                   = 3 // min size to run bft
	WorkerNumber                       = 5
	MAX_S2B_BLOCK                      = 30
	MAX_BEACON_BLOCK                   = 20
	LowerBoundPercentForIncDAO         = 3
	UpperBoundPercentForIncDAO         = 10
	TestRandom                         = true
	ValidateTimeForSpamRequestTxs      = 1581565837 // GMT: Thursday, February 13, 2020 3:50:37 AM. From this time, block will be checked spam request-reward tx
	TransactionBatchSize               = 30
	SpareTime                          = 1000             // in mili-second
	DefaultMaxBlockSyncTime            = 30 * time.Second // in second
	NumberOfFixedBeaconBlockValidators = 4
	NumberOfFixedShardBlockValidators  = 4
	Duration                           = 1000000
	MaxSubsetCommittees                = 2
	SFV3_MinShardCommitteeSize         = 8
)

// burning addresses
const (
	burningAddress  = "15pABFiJVeh9D5uiQEhQX4SVibGGbdAVipQxBdxkmDqAJaoG1EdFKHBrNfs"
	burningAddress2 = "12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA"
)

var (
	shardInsertBlockTimer                  = metrics.NewRegisteredTimer("shard/insert", nil)
	shardVerifyPreprocesingTimer           = metrics.NewRegisteredTimer("shard/verify/preprocessing", nil)
	shardVerifyPreprocesingForPreSignTimer = metrics.NewRegisteredTimer("shard/verify/preprocessingpresign", nil)
	shardVerifyWithBestStateTimer          = metrics.NewRegisteredTimer("shard/verify/withbeststate", nil)
	shardVerifyPostProcessingTimer         = metrics.NewRegisteredTimer("shard/verify/postprocessing", nil)
	shardStoreBlockTimer                   = metrics.NewRegisteredTimer("shard/storeblock", nil)
	shardUpdateBestStateTimer              = metrics.NewRegisteredTimer("shard/updatebeststate", nil)

	beaconInsertBlockTimer                  = metrics.NewRegisteredTimer("beacon/insert", nil)
	beaconVerifyPreprocesingTimer           = metrics.NewRegisteredTimer("beacon/verify/preprocessing", nil)
	beaconVerifyPreprocesingForPreSignTimer = metrics.NewRegisteredTimer("beacon/verify/preprocessingpresign", nil)
	beaconVerifyWithBestStateTimer          = metrics.NewRegisteredTimer("beacon/verify/withbeststate", nil)
	beaconVerifyPostProcessingTimer         = metrics.NewRegisteredTimer("beacon/verify/postprocessing", nil)
	beaconStoreBlockTimer                   = metrics.NewRegisteredTimer("beacon/storeblock", nil)
	beaconUpdateBestStateTimer              = metrics.NewRegisteredTimer("beacon/updatebeststate", nil)
)

var (
	BeaconSyncMode = ARCHIVE_SYNC_MODE
	ShardSyncMode  = ARCHIVE_SYNC_MODE
	configCache4GB = CacheConfig{
		trieJournalCacheSize: 32,
		blockTriesInMemory:   uint64(500),
		trieNodeLimit:        common.StorageSize(128 * 1024 * 1024),
		trieImgsLimit:        common.StorageSize(4 * 1024 * 1024),
	}
	configCache8GB = CacheConfig{
		trieJournalCacheSize: 32,
		blockTriesInMemory:   uint64(2000),
		trieNodeLimit:        common.StorageSize(512 * 1024 * 1024),
		trieImgsLimit:        common.StorageSize(4 * 1024 * 1024),
	}
	configCache16GB = CacheConfig{
		trieJournalCacheSize: 32,
		blockTriesInMemory:   uint64(10000),
		trieNodeLimit:        common.StorageSize(2 * 1024 * 1024 * 1024),
		trieImgsLimit:        common.StorageSize(4 * 1024 * 1024),
	}
	configCache32GB = CacheConfig{
		trieJournalCacheSize: 32,
		blockTriesInMemory:   uint64(20000),
		trieNodeLimit:        common.StorageSize(4 * 1024 * 1024 * 1024),
		trieImgsLimit:        common.StorageSize(4 * 1024 * 1024),
	}
)

const (
	FULL_SYNC_MODE    = "full-sync"
	ARCHIVE_SYNC_MODE = "archive-sync-mode"
)
