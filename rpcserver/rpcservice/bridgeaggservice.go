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
		BaseDecimal:       config.Param().BridgeAggParam.BaseDecimal,
		MaxLenOfPath:      config.Param().BridgeAggParam.MaxLenOfPath,
	}
	return res, nil
}

func (blockService BlockService) BridgeAggEstimateFeeByBurntAmount(unifiedTokenID common.Hash, networkID uint, burntAmount uint64) (interface{}, error) {
	beaconBestView := blockService.BlockChain.GetBeaconBestState()
	state := beaconBestView.BridgeAggState()
	vault, err := bridgeagg.GetVault(state.UnifiedTokenInfos(), unifiedTokenID, networkID)
	if err != nil {
		return nil, NewRPCError(BridgeAggEstimateFeeByBurntAmountError, err)
	}

	x := vault.Reserve()
	y := vault.CurrentRewardReserve()
	receivedAmount, err := bridgeagg.EstimateActualAmountByBurntAmount(x, y, burntAmount)
	return &jsonresult.BridgeAggEstimateFee{
		ReceivedAmount: receivedAmount,
		Fee:            burntAmount - receivedAmount,
	}, err
}

func (blockService BlockService) BridgeAggEstimateFeeByExpectedAmount(unifiedTokenID common.Hash, networkID uint, amount uint64) (interface{}, error) {
	beaconBestView := blockService.BlockChain.GetBeaconBestState()
	state := beaconBestView.BridgeAggState()
	vault, err := bridgeagg.GetVault(state.UnifiedTokenInfos(), unifiedTokenID, networkID)
	if err != nil {
		return nil, NewRPCError(BridgeAggEstimateFeeByExpectedAmountError, err)
	}

	x := vault.Reserve()
	y := vault.CurrentRewardReserve()
	amt, err := bridgeagg.CalculateActualAmount(x, y, amount, bridgeagg.SubOperator)
	if amount > x {
		return nil, NewRPCError(BridgeAggEstimateFeeByExpectedAmountError, fmt.Errorf("Unshield amount %v > vault amount %v", amount, x))
	}
	return &jsonresult.BridgeAggEstimateFee{
		ReceivedAmount: amt,
		Fee:            amount - amt,
	}, err
}

func (blockService BlockService) BridgeAggEstimateReward(unifiedTokenID common.Hash, networkID uint, amount uint64) (interface{}, error) {
	beaconBestView := blockService.BlockChain.GetBeaconBestState()
	state := beaconBestView.BridgeAggState()
	vault, err := bridgeagg.GetVault(state.UnifiedTokenInfos(), unifiedTokenID, networkID)
	if err != nil {
		return nil, NewRPCError(BridgeAggEstimateRewardError, err)
	}

	x := vault.Reserve()
	y := vault.CurrentRewardReserve()
	amt, err := bridgeagg.CalculateActualAmount(x, y, amount, bridgeagg.AddOperator)
	return &jsonresult.BridgeAggEstimateReward{
		ReceivedAmount: amt,
		Reward:         amt - amount,
	}, err
}
