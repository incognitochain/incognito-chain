package rpcservice

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/blockchain/bridgehub"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

func (blockService BlockService) GetBridgeHubState(
	beaconHeight uint64,
) (interface{}, error) {
	beaconBestView := blockService.BlockChain.GetBeaconBestState()
	if beaconHeight == 0 {
		beaconHeight = beaconBestView.BeaconHeight
	}

	beaconFeatureStateRootHash, err := blockService.BlockChain.GetBeaconFeatureRootHash(beaconBestView, uint64(beaconHeight))
	if err != nil {
		return nil, NewRPCError(GetBridgeHubStateError, fmt.Errorf("Can't found ConsensusStateRootHash of beacon height %+v, error %+v", beaconHeight, err))
	}
	beaconFeatureStateDB, err := statedb.NewWithPrefixTrie(beaconFeatureStateRootHash, statedb.NewDatabaseAccessWarper(blockService.BlockChain.GetBeaconChainDatabase()))
	if err != nil {
		return nil, NewRPCError(GetBridgeHubStateError, err)
	}

	beaconBlocks, err := blockService.BlockChain.GetBeaconBlockByHeight(uint64(beaconHeight))
	if err != nil {
		return nil, NewRPCError(GetBridgeHubStateError, err)
	}
	beaconBlock := beaconBlocks[0]
	beaconTimeStamp := beaconBlock.Header.Timestamp

	res, err := getBridgeHubState(beaconHeight, beaconTimeStamp, beaconFeatureStateDB)
	if err != nil {
		return nil, NewRPCError(GetBridgeHubStateError, err)
	}
	return res, nil
}

func getBridgeHubState(
	beaconHeight uint64, beaconTimeStamp int64, stateDB *statedb.StateDB,
) (interface{}, error) {
	bridgeHubState, err := bridgehub.InitStateFromDB(stateDB)
	if err != nil {
		return nil, NewRPCError(GetBridgeHubStateError, err)
	}

	res := &jsonresult.BridgeHubState{
		BeaconTimeStamp: beaconTimeStamp,
		StakingInfos:    bridgeHubState.StakingInfos(),
		BridgeInfos:     bridgeHubState.BridgeInfos(),
		TokenPrices:     bridgeHubState.TokenPrices(),
		Params:          bridgeHubState.Params(),
	}
	return res, nil
}
