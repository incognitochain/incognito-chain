package repository

import (
	"context"
	"github.com/incognitochain/incognito-chain/appservices/storage/model"
	"github.com/incognitochain/incognito-chain/common"
)

type BeaconStateStorer interface {
	StoreBeaconState (ctx context.Context, beaconState model.BeaconState) error
}

type BeaconStateRetriver interface {
	GetBeaconStateByHash (hash common.Hash) model.BeaconState
	GetBeaconStateByHeight(height uint64) model.BeaconState
	GetAllBeaconState (offset uint, limit uint) []model.BeaconState
	GetLatestBeaconState () []model.BeaconState
}

type BeaconStateRepository interface {
	BeaconStateStorer
	BeaconStateRetriver
}