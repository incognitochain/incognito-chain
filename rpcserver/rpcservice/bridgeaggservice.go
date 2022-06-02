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
		BeaconTimeStamp:     beaconTimeStamp,
		UnifiedTokenInfos:   bridgeAggState.UnifiedTokenVaults(),
		WaitingUnshieldReqs: bridgeAggState.WaitingUnshieldReqs(),
		Param:               bridgeAggState.Param(),
		BaseDecimal:         config.Param().BridgeAggParam.BaseDecimal,
		MaxLenOfPath:        config.Param().BridgeAggParam.MaxLenOfPath,
	}
	return res, nil
}

func (blockService BlockService) BridgeAggEstimateFeeByBurntAmount(unifiedTokenID, tokenID common.Hash, burntAmount uint64) (interface{}, error) {
	beaconBestView := blockService.BlockChain.GetBeaconBestState()
	state := beaconBestView.BridgeAggManager().State()

	vaults, ok := state.UnifiedTokenVaults()[unifiedTokenID]
	if !ok {
		return nil, fmt.Errorf("Invalid unifiedTokenID %v", unifiedTokenID.String())
	}

	vault, ok := vaults[tokenID]
	if !ok {
		return nil, fmt.Errorf("Invalid tokenID %v", tokenID.String())
	}

	_, fee, err := bridgeagg.CalUnshieldFeeByBurnAmount(vault, burntAmount, state.Param().PercentFeeWithDec())
	if err != nil {
		return nil, NewRPCError(BridgeAggEstimateFeeByBurntAmountError, err)
	}

	receivedAmt := burntAmount - fee
	maxFee := fee
	if fee > 0 {
		maxFee = calMaxUnshieldFee(receivedAmt, state.Param().PercentFeeWithDec(), config.Param().BridgeAggParam.PercentFeeDecimal)
	}

	return &jsonresult.BridgeAggEstimateFee{
		Fee:            fee,
		ReceivedAmount: receivedAmt,
		BurntAmount:    burntAmount,
		MaxFee:         maxFee,
		MaxBurntAmount: burntAmount,
	}, nil
}

func (blockService BlockService) BridgeAggEstimateFeeByExpectedAmount(unifiedTokenID, tokenID common.Hash, amount uint64) (interface{}, error) {
	beaconBestView := blockService.BlockChain.GetBeaconBestState()
	state := beaconBestView.BridgeAggManager().State()

	vaults, ok := state.UnifiedTokenVaults()[unifiedTokenID]
	if !ok {
		return nil, fmt.Errorf("Invalid unifiedTokenID %v", unifiedTokenID.String())
	}

	vault, ok := vaults[tokenID]
	if !ok {
		return nil, fmt.Errorf("Invalid tokenID %v", tokenID.String())
	}

	_, fee, err := bridgeagg.CalUnshieldFeeByReceivedAmount(vault, amount, state.Param().PercentFeeWithDec())
	if err != nil {
		return nil, NewRPCError(BridgeAggEstimateFeeByBurntAmountError, err)
	}

	burnAmt := amount + fee
	maxFee := fee
	maxBurnAmount := burnAmt
	if fee > 0 {
		maxFee = calMaxUnshieldFee(amount, state.Param().PercentFeeWithDec(), config.Param().BridgeAggParam.PercentFeeDecimal)
		maxBurnAmount = amount + maxFee
	}

	return &jsonresult.BridgeAggEstimateFee{
		Fee:            fee,
		ReceivedAmount: amount,
		BurntAmount:    burnAmt,
		MaxFee:         maxFee,
		MaxBurntAmount: maxBurnAmount,
	}, nil
}

func calMaxUnshieldFee(receivedAmt uint64, percentFeeWithDec uint64, percentFeeDec uint64) uint64 {
	tmp := new(big.Int).Mul(new(big.Int).SetUint64(receivedAmt), new(big.Int).SetUint64(percentFeeWithDec))
	tmp.Div(tmp, new(big.Int).SetUint64(percentFeeDec))
	return tmp.Uint64()
}

func (blockService BlockService) BridgeAggEstimateReward(unifiedTokenID, tokenID common.Hash, amount uint64) (interface{}, error) {
	beaconBestView := blockService.BlockChain.GetBeaconBestState()
	state := beaconBestView.BridgeAggManager().State()

	vaults, ok := state.UnifiedTokenVaults()[unifiedTokenID]
	if !ok {
		return nil, fmt.Errorf("Invalid unifiedTokenID %v", unifiedTokenID.String())
	}

	vault, ok := vaults[tokenID]
	if !ok {
		return nil, fmt.Errorf("Invalid tokenID %v", tokenID.String())
	}

	reward, err := bridgeagg.CalRewardForRefillVault(vault, amount)
	if err != nil {
		return nil, NewRPCError(BridgeAggEstimateFeeByBurntAmountError, err)
	}

	return &jsonresult.BridgeAggEstimateReward{
		ReceivedAmount: amount + reward,
		Reward:         reward,
	}, nil
}
