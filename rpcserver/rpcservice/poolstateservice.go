package rpcservice

import (
	"errors"
	"github.com/incognitochain/incognito-chain/mempool"
)


type PoolStateService struct{
	crossShardPool *mempool.CrossShardPool
}

func (poolStateService PoolStateService) GetNextCrossShard(fromShard byte, toShard byte, startHeight uint64) uint64 {
	return mempool.GetCrossShardPool(toShard).GetNextCrossShardHeight(fromShard, toShard, startHeight)
}

func (poolStateService PoolStateService) GetBeaconPoolState() ([]uint64, error){
	beaconPool := mempool.GetBeaconPool()
	if beaconPool == nil {
		Logger.log.Debugf("handleGetBeaconPoolState result: %+v", nil)
		return nil,  errors.New("Beacon Pool not init")
	}
	return beaconPool.GetAllBlockHeight(), nil
}

func (poolStateService PoolStateService) GetShardPoolState(shardID byte) (*mempool.ShardPool, error){
	shardPool := mempool.GetShardPool(shardID)
	if shardPool == nil {
		Logger.log.Debugf("handleGetShardPoolState result: %+v", nil)
		return nil, errors.New("Shard to Beacon Pool not init")
	}

	return shardPool, nil
}

func (poolStateService PoolStateService) GetShardPoolLatestValidHeight(shardID byte) (uint64, error){
	shardPool, err := poolStateService.GetShardPoolState(shardID)
	if err != nil{
		return uint64(0), err
	}

	return shardPool.GetLatestValidBlockHeight(), nil
}

func (poolStateService PoolStateService) GetShardToBeaconPoolStateV2() (map[byte][]uint64, map[byte]uint64, error){
	shardToBeaconPool := mempool.GetShardToBeaconPool()
	if shardToBeaconPool == nil {
		Logger.log.Debugf("handleGetShardToBeaconPoolStateV2 result: %+v", nil)
		return nil, nil, errors.New("Shard to Beacon Pool not init")
	}
	allBlockHeight := shardToBeaconPool.GetAllBlockHeight()
	allLatestBlockHeight := shardToBeaconPool.GetLatestValidPendingBlockHeight()

	return allBlockHeight, allLatestBlockHeight, nil
}

func (poolStateService PoolStateService) GetCrossShardPoolStateV2(shardID byte) (map[byte][]uint64, map[byte][]uint64, error){
	crossShardPool := mempool.GetCrossShardPool(shardID)
	if crossShardPool == nil {
		Logger.log.Debugf("handleGetCrossShardPoolStateV2 result: %+v", nil)
		return nil, nil, errors.New("Cross Shard Pool not init")
	}
	allValidBlockHeight := crossShardPool.GetValidBlockHeight()
	allPendingBlockHeight := crossShardPool.GetPendingBlockHeight()

	return allValidBlockHeight, allPendingBlockHeight, nil
}

func (poolStateService PoolStateService) GetShardPoolStateV2(shardID byte) (*mempool.ShardPool, error){
	shardPool := mempool.GetShardPool(shardID)
	if shardPool == nil {
		Logger.log.Debugf("handleGetShardPoolStateV2 result: %+v", nil)
		return nil, errors.New("Shard to Beacon Pool not init")
	}
	return shardPool, nil
}

func (poolStateService PoolStateService) GetBeaconPoolStateV2() (*mempool.BeaconPool, error){
	beaconPool := mempool.GetBeaconPool()
	if beaconPool == nil {
		Logger.log.Debugf("handleGetBeaconPoolStateV2 result: %+v", nil)
		return nil, errors.New("Beacon Pool not init")
	}

	return beaconPool, nil
}


