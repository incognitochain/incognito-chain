package repository

import (
	"context"
	"github.com/incognitochain/incognito-chain/appservices/data"
)

type BeaconStateRepository interface {
	StoreLatestBeaconState(ctx context.Context ,beacon *data.Beacon) error
}