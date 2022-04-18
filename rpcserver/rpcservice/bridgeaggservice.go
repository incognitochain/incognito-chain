package rpcservice

import (
	"fmt"
	"math/big"

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
	expectedAmount, err := bridgeagg.EstimateActualAmountByBurntAmount(x, y, burntAmount, vault.IsPaused())
	return &jsonresult.BridgeAggEstimateFee{
		ExpectedAmount: expectedAmount,
		Fee:            burntAmount - expectedAmount,
		BurntAmount:    burntAmount,
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
	fee, err := bridgeagg.CalculateDeltaY(x, y, amount, bridgeagg.SubOperator, vault.IsPaused())
	if amount > x {
		return nil, NewRPCError(BridgeAggEstimateFeeByExpectedAmountError, fmt.Errorf("Unshield amount %v > vault amount %v", amount, x))
	}
	burntAmount := big.NewInt(0).Add(big.NewInt(0).SetUint64(fee), big.NewInt(0).SetUint64(amount))
	if !burntAmount.IsUint64() {
		return nil, NewRPCError(BridgeAggEstimateFeeByExpectedAmountError, fmt.Errorf("Value is not unit64"))
	}
	return &jsonresult.BridgeAggEstimateFee{
		ExpectedAmount: amount,
		Fee:            fee,
		BurntAmount:    burntAmount.Uint64(),
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
	amt, err := bridgeagg.CalculateShieldActualAmount(x, y, amount, vault.IsPaused())
	return &jsonresult.BridgeAggEstimateReward{
		ReceivedAmount: amt,
		Reward:         amt - amount,
	}, err
}
