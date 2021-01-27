package storage

import (
	"context"
	"github.com/incognitochain/incognito-chain/appservices/data"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
)

func StoreLatestBeaconFinalState(ctx context.Context, beacon *data.Beacon) error {
	Logger.log.Infof("Store beacon with block hash %v and block height %d", beacon.BlockHash, beacon.Height)
	return GetDBDriver(MONGODB).GetBeaconStateRepository().StoreLatestBeaconState(ctx, beacon)
}

func StoreLatestShardFinalState(ctx context.Context, shard *data.Shard) error {
	Logger.log.Infof("Store shard with block hash %v and block height %d of Shard ID %d", shard.BlockHash, shard.Height, shard.ShardID)
	return GetDBDriver(MONGODB).GetShardStateRepository().StoreLatestShardState(ctx, shard)
}

func StorePDEShareState(ctx context.Context, pdeContributionStore *rawdbv2.PDEContributionStore, pdeTradeStore *rawdbv2.PDETradeStore, pdeCrossTradeStore *rawdbv2.PDECrossTradeStore,
pdeWithdrawalStatusStore *rawdbv2.PDEWithdrawalStatusStore, pdeFeeWithdrawalStatusStore *rawdbv2.PDEFeeWithdrawalStatusStore) error {
	Logger.log.Infof("Store pdeShare")
	return GetDBDriver(MONGODB).GetPDEStateRepository().StoreLatestPDEBestState(ctx, pdeContributionStore, pdeTradeStore, pdeCrossTradeStore, pdeWithdrawalStatusStore, pdeFeeWithdrawalStatusStore)
}