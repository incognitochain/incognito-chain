package appservices

import (
	"context"
	"github.com/incognitochain/incognito-chain/appservices/data"
	"github.com/incognitochain/incognito-chain/appservices/storage"
	_ "github.com/incognitochain/incognito-chain/appservices/storage/impl"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/databasemp"
	"time"
)

// default value
const (
	defaultScanTime = 1 * time.Second
)

// config is a descriptor containing the memory pool configuration.
type AppConfig struct {
	BlockChain         *blockchain.BlockChain       // Block chain of node
	DataBaseAppService databasemp.DatabaseInterface // database is used for storage data in mempool into lvdb

}

type AppService struct {
	// The following variables must only be used atomically.
	config             AppConfig
	latestBeaconHeight uint64
	latestShardHeight  [255]uint64
}

func (app *AppService) Init(cfg *AppConfig) {
	app.config = *cfg
	app.latestBeaconHeight = app.getProcessedFinalBeaconHeight()
	numShards := app.config.BlockChain.GetActiveShardNumber()
	for i := 0; i < numShards; i++ {
		app.latestShardHeight[i] = app.getProcessedFinalShardHeight(byte(i))
	}
}

func (app *AppService) Start(cQuit chan struct{}) {
	/*Logger.log.Infof("Start app service")
	beaconTicker := time.Tick(defaultScanTime)
	shardTicker := time.Tick(defaultScanTime)
	for { //actor loop
		select {
			case <-cQuit:
				return
			case <- beaconTicker:
				beacons := app.getFinalBeacons()
	*/ //json, err := json.Marshal(beacons)
	/*if err == nil {
		Logger.log.Infof(string(json))
	}*/
	/*			app.publishBeaconToExternal(beacons)
				//TODO: Call function to get new final height.
		case <- shardTicker:*/
	//shards := app.getAllFinalShardStates()
	//json, err := json.Marshal(shards)
	/*if err == nil {
		Logger.log.Infof(string(json))
	}*/
	//app.publishShardToExternal(shards)
	/*}
	}*/

}

func (app *AppService) PublishBeaconState(beaconState *blockchain.BeaconBestState) {
	Logger.log.Infof("Publish beaconState with hash %v at height %d", beaconState.Hash(), beaconState.BeaconHeight)
	beacon := data.NewBeaconFromBeaconState(beaconState)
	storage.StoreLatestBeaconFinalState(context.TODO(), beacon)
}

func (app *AppService) publishBeaconToExternal(beacons []*data.Beacon) {
	//TODO: in this stage store to db, next stage public to broker
	Logger.log.Infof("Public %d beacon to external", len(beacons))
	for _, beacon := range beacons {
		storage.StoreLatestBeaconFinalState(context.TODO(), beacon)
	}
	app.storeProcessedFinalBeaconHeight()
}

func (app *AppService) publishShardToExternal(shards []*data.Shard) {
	for _, shard := range shards {
		storage.StoreLatestShardFinalState(context.TODO(), shard)
	}
	app.storeAllFinalShardHeight()
}

func (app *AppService) getFinalBeacons() []*data.Beacon {

	beaconState, ok := app.config.BlockChain.BeaconChain.GetFinalView().(*blockchain.BeaconBestState)
	if !ok {
		return []*data.Beacon{}
	}
	var beacons []*data.Beacon
	Logger.log.Infof("Recently beacon height handle %d  ", app.latestBeaconHeight)

	for beaconState.GetHeight() > app.latestBeaconHeight {
		var err error
		Logger.log.Infof("Hanldle Block at %d with hash %v", beaconState.GetHeight(), beaconState.BestBlockHash)
		//Logger.log.Infof("Hanldle Block at  %v", beaconState)
		beacons = append(beacons, data.NewBeaconFromBeaconState(beaconState))
		//Specially Condition, this state is intial state we don't have a previously to get, break the loop
		if beaconState.GetHeight() == 1 {
			break
		}
		previousBlockHash := beaconState.GetPreviousHash()
		beaconState, err = app.config.BlockChain.GetDetailsBeaconViewStateDataFromBlockHash(*previousBlockHash)
		if err != nil {
			Logger.log.Errorf("Can't get state at block hash %v", previousBlockHash)
			return []*data.Beacon{}
		}
	}
	if len(beacons) > 0 {
		app.latestBeaconHeight = beacons[0].Height
	}
	return beacons
}

func (app *AppService) getAllFinalShardStates() []*data.Shard {
	numShards := app.config.BlockChain.GetActiveShardNumber()
	shards := make([]*data.Shard, 0)
	for i := 0; i < numShards; i++ {
		shards = append(shards, app.getFinalShardStatesOf(byte(i))...)
	}
	return shards
}

func (app *AppService) getFinalShardStatesOf(shardId byte) []*data.Shard {
	shardState, ok := app.config.BlockChain.ShardChain[shardId].GetFinalView().(*blockchain.ShardBestState)
	if !ok {
		return []*data.Shard{}
	}
	shards := make([]*data.Shard, 0)

	for shardState.GetHeight() > app.latestBeaconHeight {
		var err error
		Logger.log.Infof("Hanldle Block at %d with hash %v", shardState.GetHeight(), shardState.BestBlockHash)
		//Logger.log.Infof("Hanldle Block at  %v", beaconState)
		shards = append(shards, data.NewShardFromShardState(shardState))
		//Specially Condition, this state is intial state we don't have a previously to get break the loop
		if shardState.GetHeight() == 1 {
			break
		}
		previousBlockHash := shardState.GetPreviousHash()
		shardState, err = app.config.BlockChain.GetDetailsShardViewStateDataFromBlockHash(shardId, *previousBlockHash)
		if err != nil {
			Logger.log.Errorf("Can't get state at block hash %v", previousBlockHash)
			return []*data.Shard{}
		}
	}
	if len(shards) > 0 {
		app.latestBeaconHeight = shards[0].Height
	}
	return shards
}

func (app *AppService) storeProcessedFinalBeaconHeight() error {
	return app.config.DataBaseAppService.Put([]byte("Final-last-beacon-process-height"), common.Uint64ToBytes(app.latestBeaconHeight))
}

func (app *AppService) getProcessedFinalBeaconHeight() uint64 {
	b := make([]byte, 8)
	b, err := app.config.DataBaseAppService.Get([]byte("Final-last-beacon-process-height"))
	if err != nil {
		return 0
	}
	value, _ := common.BytesToUint64(b)
	return value
}

func (app *AppService) storeAllFinalShardHeight() error {
	for i := 0; i < app.config.BlockChain.GetActiveShardNumber(); i++ {
		if err := app.storeProcessedFinalShardHeight(byte(i)); err != nil {
			return err
		}
	}
	return nil
}

func (app *AppService) storeProcessedFinalShardHeight(shardId byte) error {
	return app.config.DataBaseAppService.Put(append([]byte("Final-last-shard-process-height-"), shardId), common.Uint64ToBytes(app.latestShardHeight[shardId]))
}

func (app *AppService) getProcessedFinalShardHeight(shardId byte) uint64 {
	b := make([]byte, 8)
	b, err := app.config.DataBaseAppService.Get(append([]byte("Final-last-shard-process-height-"), shardId))
	if err != nil {
		return 0
	}
	value, _ := common.BytesToUint64(b)
	return value
}
