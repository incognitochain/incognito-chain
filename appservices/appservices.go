package appservices

import (
	"context"
	"github.com/incognitochain/incognito-chain/appservices/data"
	"github.com/incognitochain/incognito-chain/appservices/storage"
	_ "github.com/incognitochain/incognito-chain/appservices/storage/impl"
	"github.com/incognitochain/incognito-chain/blockchain"
)

// config is a descriptor containing the memory pool configuration.
type AppConfig struct {
	BlockChain         *blockchain.BlockChain       // Block chain of node
}

type AppService struct {
	// The following variables must only be used atomically.
	config             AppConfig
}

func (app *AppService) Init(cfg *AppConfig) {
	app.config = *cfg
}

func (app *AppService) PublishBeaconState(beaconState *blockchain.BeaconBestState) error {
	Logger.log.Debugf("Publish beaconState with hash %v at height %d", beaconState.BestBlock.Hash().String(), beaconState.BeaconHeight)
	beacon := data.NewBeaconFromBeaconState(beaconState)
	return storage.StoreLatestBeaconFinalState(context.TODO(), beacon)
}

func (app *AppService) PublishShardState(shardBestState *blockchain.ShardBestState) error {
	Logger.log.Infof("Publish shardState with hash %v at height %d of Shard ID: %d", shardBestState.BestBlock.Hash().String(), shardBestState.BeaconHeight, shardBestState.ShardID)
	shard := data.NewShardFromShardState(shardBestState)
	return storage.StoreLatestShardFinalState(context.TODO(), shard)
}
