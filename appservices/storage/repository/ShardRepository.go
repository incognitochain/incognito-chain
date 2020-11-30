package repository

import (
	"context"
	"github.com/incognitochain/incognito-chain/appservices/storage/model"
"github.com/incognitochain/incognito-chain/common"
)

type ShardStateStorer interface {
	StoreShardState (ctx context.Context, shardState model.ShardState) error
}

type ShardStateRetriver interface {
	GetShardStateByHash (hash common.Hash) model.ShardState
	GetShardStateByHeight(height uint64) model.ShardState
	GetAllShardState (offset uint, limit uint) []model.ShardState
	GetLatestShardState () []model.ShardState
}

type ShardStateRepository interface {
	ShardStateStorer
	ShardStateRetriver
}
