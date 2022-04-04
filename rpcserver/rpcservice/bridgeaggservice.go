package rpcservice

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/blockchain/bridgeagg"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

func (blockService BlockService) GetBridgeAggState(
	beaconHeight uint64,
) (interface{}, error) {
	beaconBestView := blockService.BlockChain.GetBeaconBestState()
	if beaconHeight == 0 {
		beaconHeight = beaconBestView.BeaconHeight
	}

	beaconFeatureStateRootHash, err := blockService.BlockChain.GetBeaconFeatureRootHash(beaconBestView, uint64(beaconHeight))
	if err != nil {
		return nil, NewRPCError(GetBridgeAggStateError, fmt.Errorf("Can't found ConsensusStateRootHash of beacon height %+v, error %+v", beaconHeight, err))
	}
	beaconFeatureStateDB, err := statedb.NewWithPrefixTrie(beaconFeatureStateRootHash, statedb.NewDatabaseAccessWarper(blockService.BlockChain.GetBeaconChainDatabase()))
	if err != nil {
		return nil, NewRPCError(GetBridgeAggStateError, err)
	}

	beaconBlocks, err := blockService.BlockChain.GetBeaconBlockByHeight(uint64(beaconHeight))
	if err != nil {
		return nil, NewRPCError(GetBridgeAggStateError, err)
	}
	beaconBlock := beaconBlocks[0]
	beaconTimeStamp := beaconBlock.Header.Timestamp

	res, err := getBridgeAggState(beaconHeight, beaconTimeStamp, beaconFeatureStateDB)
	if err != nil {
		return nil, NewRPCError(GetBridgeAggStateError, err)
	}
	return res, nil
}

func getBridgeAggState(
	beaconHeight uint64, beaconTimeStamp int64, stateDB *statedb.StateDB,
) (interface{}, error) {
	bridgeAggState, err := bridgeagg.InitStateFromDB(stateDB)
	if err != nil {
		return nil, NewRPCError(GetBridgeAggStateError, err)
	}

	res := &jsonresult.BridgeAggState{
		BeaconTimeStamp:   beaconTimeStamp,
		UnifiedTokenInfos: bridgeAggState.UnifiedTokenInfos(),
		BaseUpperDecimal:  config.Param().BridgeAggParam.BaseUpperDecimal,
		BaseLowerDecimal:  config.Param().BridgeAggParam.BaseLowerDecimal,
	}
	return res, nil
}

func (blockService BlockService) EstimateReceivedAmount(unifiedTokenID common.Hash, networkID uint, burntAmount uint64) (interface{}, error) {
	beaconBestView := blockService.BlockChain.GetBeaconBestState()
	state := beaconBestView.BridgeAggState()
	vault, err := bridgeagg.GetVault(state.UnifiedTokenInfos(), unifiedTokenID, networkID)
	if err != nil {
		return nil, NewRPCError(BridgeAggEstimateReceivedAmountError, err)
	}

	x := vault.Reserve()
	y := vault.CurrentRewardReserve()
	receivedAmount, err := bridgeagg.EstimateActualAmountByBurntAmount(x, y, burntAmount)
	return &jsonresult.BridgeAggEstimateReceivedAmount{
		ReceivedAmount: receivedAmount,
		Fee:            burntAmount - receivedAmount,
	}, err
}
