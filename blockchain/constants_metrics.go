package blockchain

import (
	"github.com/incognitochain/incognito-chain/metrics"
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
