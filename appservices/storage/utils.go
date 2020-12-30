package storage

import (
	"context"
	"github.com/incognitochain/incognito-chain/appservices/data"
)

func StoreLatestBeaconFinalState(ctx context.Context, beacon *data.Beacon) error {
	Logger.log.Infof("Store beacon with block hash %v and block height %d", beacon.BlockHash, beacon.Height)
	return GetDBDriver(MONGODB).GetBeaconStateRepository().StoreLatestBeaconState(ctx, beacon)
}

func StoreLatestShardFinalState(ctx context.Context, shard *data.Shard) error {
	Logger.log.Infof("Store shard with block hash %v and block height %d of Shard ID %d", shard.BlockHash, shard.Height, shard.ShardID)
	return GetDBDriver(MONGODB).GetShardStateRepository().StoreLatestShardState(ctx, shard)
}