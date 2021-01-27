package repository

import (
	"context"
	"github.com/incognitochain/incognito-chain/appservices/data"
)

type ShardStateRepository interface {
	StoreLatestShardState(ctx context.Context ,shard *data.Shard) error
}
