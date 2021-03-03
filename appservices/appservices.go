package appservices

import (
	"context"
	"github.com/incognitochain/incognito-chain/appservices/data"
	"github.com/incognitochain/incognito-chain/appservices/storage"
	"github.com/incognitochain/incognito-chain/appservices/storage/impl"
	_ "github.com/incognitochain/incognito-chain/appservices/storage/impl"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
)

// config is a descriptor containing the memory pool configuration.
type AppConfig struct {
	BlockChain *blockchain.BlockChain // Block chain of node
}

type AppService struct {
	// The following variables must only be used atomically.
	config AppConfig
}

func (app *AppService) Init(cfg *AppConfig) {
	app.config = *cfg
}

func (app *AppService) PublishBeaconState(beaconState *blockchain.BeaconBestState) error {
	Logger.log.Debugf("Publish beaconState with hash %v at height %d", beaconState.BestBlock.Hash().String(), beaconState.BeaconHeight)
	beacon := data.NewBeaconFromBeaconState(beaconState)
	err := storage.StoreLatestBeaconFinalState(context.TODO(), beacon)
	if err != nil && !impl.IsMongoDupKey(err) {
		return err
	}
	return nil
}

func (app *AppService) PublishShardState(shardBestState *blockchain.ShardBestState) error {
	Logger.log.Infof("Publish shardState with hash %v at height %d of Shard ID: %d", shardBestState.BestBlock.Hash().String(), shardBestState.BeaconHeight, shardBestState.ShardID)
	shard := data.NewShardFromShardState(shardBestState)
	err := storage.StoreLatestShardFinalState(context.TODO(), shard)
	if err != nil && !impl.IsMongoDupKey(err) {
		return err
	}
	return nil
}

func (app *AppService) PublishPDEState(pdeContributionStore *rawdbv2.PDEContributionStore, pdeTradeStore *rawdbv2.PDETradeStore, pdeCrossTradeStore *rawdbv2.PDECrossTradeStore,
	pdeWithdrawalStatusStore *rawdbv2.PDEWithdrawalStatusStore, pdeFeeWithdrawalStatusStore *rawdbv2.PDEFeeWithdrawalStatusStore) error {
	Logger.log.Infof("Publish pdeStateStore")
	err := storage.StorePDEShareState(context.TODO(), pdeContributionStore, pdeTradeStore, pdeCrossTradeStore, pdeWithdrawalStatusStore, pdeFeeWithdrawalStatusStore)
	if err != nil && !impl.IsMongoDupKey(err) {
		return err
	}
	return nil
}
