package pdex

import (
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

func InitStatesFromDB(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
) (map[uint]State, error) {
	res := make(map[uint]State)
	if beaconHeight >= config.Param().PDexParams.Pdexv3BreakPointHeight {
		if beaconHeight == config.Param().PDexParams.Pdexv3BreakPointHeight {
			res[AmplifierVersion] = newStateV2()
		} else {
			state, err := initStateV2FromDB(stateDB)
			if err != nil {
				return res, err
			}
			res[AmplifierVersion] = state
		}
	}
	if beaconHeight == 0 || beaconHeight == 1 {
		res[BasicVersion] = newStateV1()
	} else {
		state, err := initStateV1(stateDB, beaconHeight)
		if err != nil {
			return res, err
		}
		res[BasicVersion] = state
	}
	return res, nil
}

func InitStateFromDB(stateDB *statedb.StateDB, beaconHeight uint64, version uint) (State, error) {
	switch version {
	case BasicVersion:
		if beaconHeight == 0 || beaconHeight == 1 {
			return newStateV1(), nil
		}
		return initStateV1(stateDB, beaconHeight)
	case AmplifierVersion:
		if beaconHeight < config.Param().PDexParams.Pdexv3BreakPointHeight {
			return nil, fmt.Errorf("[pdex] Beacon height %v < Pdexv3BreakPointHeight %v", beaconHeight, config.Param().PDexParams.Pdexv3BreakPointHeight)
		}
		return initStateV2FromDB(stateDB)
	default:
		return nil, errors.New("Can not recognize version")
	}
}

func initStateV1(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
) (*stateV1, error) {
	waitingContributions, err := statedb.GetWaitingPDEContributions(stateDB, beaconHeight)
	if err != nil {
		return nil, err
	}
	poolPairs, err := statedb.GetPDEPoolPair(stateDB, beaconHeight)
	if err != nil {
		return nil, err
	}
	shares, err := statedb.GetPDEShares(stateDB, beaconHeight)
	if err != nil {
		return nil, err
	}
	tradingFees, err := statedb.GetPDETradingFees(stateDB, beaconHeight)
	if err != nil {
		return nil, err
	}
	return newStateV1WithValue(
		waitingContributions,
		poolPairs,
		shares,
		tradingFees,
	), nil
}

func initStateV2FromDB(
	stateDB *statedb.StateDB,
) (*stateV2, error) {
	paramsState, err := statedb.GetPdexv3Params(stateDB)
	params := NewParamsWithValue(paramsState)
	if err != nil {
		return nil, err
	}
	waitingContributions, err := statedb.GetPdexv3WaitingContributions(stateDB)
	if err != nil {
		return nil, err
	}
	poolPairs, err := initPoolPairStatesFromDB(stateDB)
	if err != nil {
		return nil, err
	}
	nftIDs, err := statedb.GetPdexv3NftIDs(stateDB)
	if err != nil {
		return nil, err
	}
	stakingPools, err := initStakingPoolsFromDB(params.StakingPoolsShare, stateDB)
	if err != nil {
		return nil, err
	}
	return newStateV2WithValue(
		waitingContributions, make(map[string]rawdbv2.Pdexv3Contribution),
		poolPairs, params, stakingPools, nftIDs,
	), nil
}

func initPoolPairStatesFromDB(stateDB *statedb.StateDB) (map[string]*PoolPairState, error) {
	poolPairsStates, err := statedb.GetPdexv3PoolPairs(stateDB)
	if err != nil {
		return nil, err
	}
	res := make(map[string]*PoolPairState)
	for poolPairID, poolPairState := range poolPairsStates {
		lpFeesPerShare, err := statedb.GetPdexv3PoolPairLpFeesPerShares(stateDB, poolPairID)
		if err != nil {
			return nil, err
		}
		protocolFees, err := statedb.GetPdexv3PoolPairProtocolFees(stateDB, poolPairID)
		if err != nil {
			return nil, err
		}
		stakingPoolFees, err := statedb.GetPdexv3PoolPairStakingPoolFees(stateDB, poolPairID)
		if err != nil {
			return nil, err
		}
		shares, err := initShares(poolPairID, stateDB)
		if err != nil {
			return nil, err
		}

		orderbook := &Orderbook{[]*Order{}}
		orderMap, err := statedb.GetPdexv3Orders(stateDB, poolPairState.PoolPairID())
		if err != nil {
			return nil, err
		}
		for _, item := range orderMap {
			v := item.Value()
			orderbook.InsertOrder(&v)
		}
		tempMakingVolume, err := statedb.GetPdexv3PoolPairMakingVolume(stateDB, poolPairID)
		if err != nil {
			return nil, err
		}
		makingVolume := make(map[common.Hash]*MakingVolume)
		for tokenID, value := range tempMakingVolume {
			if makingVolume[tokenID] == nil {
				makingVolume[tokenID] = NewMakingVolume()
			}
			for nftID, amount := range value {
				makingVolume[tokenID].volume[nftID] = amount
			}
		}
		tempOrderReward, err := statedb.GetPdexv3PoolPairOrderReward(stateDB, poolPairID)
		if err != nil {
			return nil, err
		}
		orderReward := make(map[string]*OrderReward)
		for nftID, value := range tempOrderReward {
			if orderReward[nftID] == nil {
				orderReward[nftID] = NewOrderReward()
			}
			for tokenID, amount := range value {
				orderReward[nftID].uncollectedRewards[tokenID] = amount
			}
		}
		lmRewardsPerShare, err := statedb.GetPdexv3PoolPairLmRewardPerShares(stateDB, poolPairID)
		if err != nil {
			return nil, err
		}
		lmLockedShare, err := statedb.GetPdexv3PoolPairLmLockedShare(stateDB, poolPairID)
		if err != nil {
			return nil, err
		}

		poolPair := NewPoolPairStateWithValue(
			poolPairState.Value(), shares, *orderbook,
			lpFeesPerShare, lmRewardsPerShare, protocolFees, stakingPoolFees,
			makingVolume, orderReward, lmLockedShare,
		)
		res[poolPairID] = poolPair
	}
	return res, nil
}

func initShares(poolPairID string, stateDB *statedb.StateDB) (map[string]*Share, error) {
	res := make(map[string]*Share)
	shareStates, err := statedb.GetPdexv3Shares(stateDB, poolPairID)
	if err != nil {
		return nil, err
	}
	for nftID, shareState := range shareStates {
		tradingFees, err := statedb.GetPdexv3ShareTradingFees(stateDB, poolPairID, nftID)
		if err != nil {
			return nil, err
		}
		lastLPFeesPerShare, err := statedb.GetPdexv3ShareLastLpFeesPerShare(stateDB, poolPairID, nftID)
		if err != nil {
			return nil, err
		}
		lastLmRewardsPerShare, err := statedb.GetPdexv3ShareLastLmRewardPerShare(stateDB, poolPairID, nftID)
		if err != nil {
			return nil, err
		}
		res[nftID] = NewShareWithValue(
			shareState.Amount(), shareState.LmLockedAmount(), tradingFees, lastLPFeesPerShare, lastLmRewardsPerShare,
		)
	}
	return res, nil
}

func initStakers(stakingPoolID string, stateDB *statedb.StateDB) (map[string]*Staker, uint64, error) {
	res := make(map[string]*Staker)
	totalLiquidity := uint64(0)
	stakerStates, err := statedb.GetPdexv3Stakers(stateDB, stakingPoolID)
	if err != nil {
		return res, totalLiquidity, err
	}
	for nftID, stakerState := range stakerStates {
		totalLiquidity += stakerState.Liquidity()
		rewards, err := statedb.GetPdexv3StakerRewards(stateDB, stakingPoolID, nftID)
		if err != nil {
			return res, totalLiquidity, err
		}
		lastRewardsPerShare, err := statedb.GetPdexv3StakerLastRewardsPerShare(stateDB, stakingPoolID, nftID)
		if err != nil {
			return res, totalLiquidity, err
		}
		res[nftID] = NewStakerWithValue(stakerState.Liquidity(), rewards, lastRewardsPerShare)
	}
	return res, totalLiquidity, nil
}

func initStakingPoolsFromDB(stakingPoolsShare map[string]uint, stateDB *statedb.StateDB) (map[string]*StakingPoolState, error) {
	res := map[string]*StakingPoolState{}
	for stakingPoolID := range stakingPoolsShare {
		stakers, liquidity, err := initStakers(stakingPoolID, stateDB)
		if err != nil {
			return nil, err
		}
		rewardsPerShare, err := statedb.GetPdexv3StakingPoolRewardsPerShare(stateDB, stakingPoolID)
		if err != nil {
			return nil, err
		}
		res[stakingPoolID] = NewStakingPoolStateWithValue(liquidity, stakers, rewardsPerShare)
	}
	return res, nil
}

func InitWaitingContributionsFromDB(stateDB *statedb.StateDB) (map[string]rawdbv2.Pdexv3Contribution, error) {
	return statedb.GetPdexv3WaitingContributions(stateDB)
}

func InitPoolPairIDsFromDB(stateDB *statedb.StateDB) ([]string, error) {
	res := []string{}
	poolPairsStates, err := statedb.GetPdexv3PoolPairs(stateDB)
	if err != nil {
		return nil, err
	}
	for k := range poolPairsStates {
		res = append(res, k)
	}
	return res, nil
}

func InitIntermediatePoolPairStatesFromDB(stateDB *statedb.StateDB) (map[string]*PoolPairState, error) {
	res := make(map[string]*PoolPairState)
	poolPairsStates, err := statedb.GetPdexv3PoolPairs(stateDB)
	if err != nil {
		return nil, err
	}
	for poolPairID, poolPairState := range poolPairsStates {
		lpFeesPerShare, err := statedb.GetPdexv3PoolPairLpFeesPerShares(stateDB, poolPairID)
		if err != nil {
			return nil, err
		}
		protocolFees, err := statedb.GetPdexv3PoolPairProtocolFees(stateDB, poolPairID)
		if err != nil {
			return nil, err
		}
		stakingPoolFees, err := statedb.GetPdexv3PoolPairStakingPoolFees(stateDB, poolPairID)
		if err != nil {
			return nil, err
		}
		lmRewardsPerShare, err := statedb.GetPdexv3PoolPairLmRewardPerShares(stateDB, poolPairID)
		if err != nil {
			return nil, err
		}
		lmLockedShare, err := statedb.GetPdexv3PoolPairLmLockedShare(stateDB, poolPairID)
		if err != nil {
			return nil, err
		}

		poolPair := NewPoolPairStateWithValue(
			poolPairState.Value(), nil, Orderbook{},
			lpFeesPerShare, lmRewardsPerShare, protocolFees, stakingPoolFees,
			map[common.Hash]*MakingVolume{}, map[string]*OrderReward{}, lmLockedShare,
		)
		res[poolPairID] = poolPair
	}
	return res, nil
}

func InitLiquidityPoolPairStatesFromDB(stateDB *statedb.StateDB) (map[string]*PoolPairState, error) {
	res := make(map[string]*PoolPairState)
	poolPairsStates, err := statedb.GetPdexv3PoolPairs(stateDB)
	if err != nil {
		return nil, err
	}
	for poolPairID, poolPairState := range poolPairsStates {
		orderbook := &Orderbook{[]*Order{}}
		orderMap, err := statedb.GetPdexv3Orders(stateDB, poolPairID)
		if err != nil {
			return nil, err
		}
		for _, item := range orderMap {
			v := item.Value()
			orderbook.InsertOrder(&v)
		}

		poolPair := NewPoolPairStateWithValue(
			poolPairState.Value(), nil, *orderbook,
			nil, nil, nil, nil,
			nil, nil, nil,
		)
		res[poolPairID] = poolPair
	}
	return res, nil
}

func InitFullPoolPairStatesFromDB(stateDB *statedb.StateDB) (map[string]*PoolPairState, error) {
	res := make(map[string]*PoolPairState)
	poolPairsStates, err := statedb.GetPdexv3PoolPairs(stateDB)
	if err != nil {
		return nil, err
	}
	for poolPairID, poolPairState := range poolPairsStates {
		lpFeesPerShare, err := statedb.GetPdexv3PoolPairLpFeesPerShares(stateDB, poolPairID)
		if err != nil {
			return nil, err
		}
		protocolFees, err := statedb.GetPdexv3PoolPairProtocolFees(stateDB, poolPairID)
		if err != nil {
			return nil, err
		}
		stakingPoolFees, err := statedb.GetPdexv3PoolPairStakingPoolFees(stateDB, poolPairID)
		if err != nil {
			return nil, err
		}
		tempMakingVolume, err := statedb.GetPdexv3PoolPairMakingVolume(stateDB, poolPairID)
		if err != nil {
			return nil, err
		}
		makingVolume := make(map[common.Hash]*MakingVolume)
		for tokenID, value := range tempMakingVolume {
			if makingVolume[tokenID] == nil {
				makingVolume[tokenID] = NewMakingVolume()
			}
			for nftID, amount := range value {
				makingVolume[tokenID].volume[nftID] = amount
			}
		}
		tempOrderReward, err := statedb.GetPdexv3PoolPairOrderReward(stateDB, poolPairID)
		if err != nil {
			return nil, err
		}
		orderReward := make(map[string]*OrderReward)
		for nftID, value := range tempOrderReward {
			if orderReward[nftID] == nil {
				orderReward[nftID] = NewOrderReward()
			}
			for tokenID, amount := range value {
				orderReward[nftID].uncollectedRewards[tokenID] = amount
			}
		}
		orderbook := &Orderbook{[]*Order{}}
		orderMap, err := statedb.GetPdexv3Orders(stateDB, poolPairID)
		if err != nil {
			return nil, err
		}
		for _, item := range orderMap {
			v := item.Value()
			orderbook.InsertOrder(&v)
		}
		shares, err := initShares(poolPairID, stateDB)
		if err != nil {
			return nil, err
		}
		lmRewardsPerShare, err := statedb.GetPdexv3PoolPairLmRewardPerShares(stateDB, poolPairID)
		if err != nil {
			return nil, err
		}
		lmLockedShare, err := statedb.GetPdexv3PoolPairLmLockedShare(stateDB, poolPairID)
		if err != nil {
			return nil, err
		}

		poolPair := NewPoolPairStateWithValue(
			poolPairState.Value(), shares, Orderbook{},
			lpFeesPerShare, lmRewardsPerShare, protocolFees, stakingPoolFees,
			makingVolume, orderReward, lmLockedShare,
		)
		res[poolPairID] = poolPair
	}
	return res, nil
}

func InitParamFromDB(stateDB *statedb.StateDB) (*Params, error) {
	paramsState, err := statedb.GetPdexv3Params(stateDB)
	params := NewParamsWithValue(paramsState)
	if err != nil {
		return nil, err
	}
	return params, nil
}

func InitStakingPoolsFromDB(stateDB *statedb.StateDB) (map[string]*StakingPoolState, error) {
	paramsState, err := statedb.GetPdexv3Params(stateDB)
	params := NewParamsWithValue(paramsState)
	if err != nil {
		return nil, err
	}
	return initStakingPoolsFromDB(params.StakingPoolsShare, stateDB)
}

func InitNftIDsFromDB(stateDB *statedb.StateDB) (map[string]uint64, error) {
	return statedb.GetPdexv3NftIDs(stateDB)
}

func InitStakingPoolFromDB(stateDB *statedb.StateDB, stakingPoolID string) (*StakingPoolState, error) {
	res := &StakingPoolState{}
	stakers, liquidity, err := initStakers(stakingPoolID, stateDB)
	if err != nil {
		return nil, err
	}
	rewardsPerShare, err := statedb.GetPdexv3StakingPoolRewardsPerShare(stateDB, stakingPoolID)
	if err != nil {
		return nil, err
	}
	res = NewStakingPoolStateWithValue(liquidity, stakers, rewardsPerShare)
	return res, nil
}

func InitPoolPair(stateDB *statedb.StateDB, poolPairID string) (*PoolPairState, error) {
	poolPairState, err := statedb.GetPdexv3PoolPair(stateDB, poolPairID)
	if err != nil {
		return nil, err
	}

	lpFeesPerShare, err := statedb.GetPdexv3PoolPairLpFeesPerShares(stateDB, poolPairID)
	if err != nil {
		return nil, err
	}
	protocolFees, err := statedb.GetPdexv3PoolPairProtocolFees(stateDB, poolPairID)
	if err != nil {
		return nil, err
	}
	stakingPoolFees, err := statedb.GetPdexv3PoolPairStakingPoolFees(stateDB, poolPairID)
	if err != nil {
		return nil, err
	}
	tempMakingVolume, err := statedb.GetPdexv3PoolPairMakingVolume(stateDB, poolPairID)
	if err != nil {
		return nil, err
	}
	makingVolume := make(map[common.Hash]*MakingVolume)
	for tokenID, value := range tempMakingVolume {
		if makingVolume[tokenID] == nil {
			makingVolume[tokenID] = NewMakingVolume()
		}
		for nftID, amount := range value {
			makingVolume[tokenID].volume[nftID] = amount
		}
	}
	tempOrderReward, err := statedb.GetPdexv3PoolPairOrderReward(stateDB, poolPairID)
	if err != nil {
		return nil, err
	}
	orderReward := make(map[string]*OrderReward)
	for nftID, value := range tempOrderReward {
		if orderReward[nftID] == nil {
			orderReward[nftID] = NewOrderReward()
		}
		for tokenID, amount := range value {
			orderReward[nftID].uncollectedRewards[tokenID] = amount
		}
	}
	orderbook := &Orderbook{[]*Order{}}
	orderMap, err := statedb.GetPdexv3Orders(stateDB, poolPairID)
	if err != nil {
		return nil, err
	}
	for _, item := range orderMap {
		v := item.Value()
		orderbook.InsertOrder(&v)
	}
	shares, err := initShares(poolPairID, stateDB)
	if err != nil {
		return nil, err
	}
	lmRewardsPerShare, err := statedb.GetPdexv3PoolPairLmRewardPerShares(stateDB, poolPairID)
	if err != nil {
		return nil, err
	}
	lmLockedShare, err := statedb.GetPdexv3PoolPairLmLockedShare(stateDB, poolPairID)
	if err != nil {
		return nil, err
	}

	return NewPoolPairStateWithValue(
		poolPairState.Value(), shares, *orderbook,
		lpFeesPerShare, lmRewardsPerShare, protocolFees, stakingPoolFees,
		makingVolume, orderReward, lmLockedShare,
	), nil
}

func InitPoolPairShares(stateDB *statedb.StateDB, poolPairID string) (map[string]*Share, error) {
	shares, err := initShares(poolPairID, stateDB)
	if err != nil {
		return nil, err
	}
	return shares, nil
}

func InitPoolPairOrders(stateDB *statedb.StateDB, poolPairID string) (*Orderbook, error) {
	orderbook := &Orderbook{[]*Order{}}
	orderMap, err := statedb.GetPdexv3Orders(stateDB, poolPairID)
	if err != nil {
		return nil, err
	}
	for _, item := range orderMap {
		v := item.Value()
		orderbook.InsertOrder(&v)
	}
	return orderbook, nil
}

func InitPoolPairOrderRewards(stateDB *statedb.StateDB, poolPairID string) (map[string]*OrderReward, error) {
	rewards, err := statedb.GetPdexv3PoolPairOrderReward(stateDB, poolPairID)
	if err != nil {
		return nil, err
	}

	orderRewards := map[string]*OrderReward{}
	for orderID, reward := range rewards {
		orderRewards[orderID] = NewOrderReward()
		for tokenID, amount := range reward {
			orderRewards[orderID].uncollectedRewards[tokenID] = amount
		}
	}
	return orderRewards, nil
}

func InitStateV2FromDBWithoutNftIDs(stateDB *statedb.StateDB, beaconHeight uint64) (*stateV2, error) {
	if beaconHeight < config.Param().PDexParams.Pdexv3BreakPointHeight {
		return nil, fmt.Errorf("[pdex] Beacon height %v < Pdexv3BreakPointHeight %v", beaconHeight, config.Param().PDexParams.Pdexv3BreakPointHeight)
	}
	paramsState, err := statedb.GetPdexv3Params(stateDB)
	params := NewParamsWithValue(paramsState)
	if err != nil {
		return nil, err
	}
	waitingContributions, err := statedb.GetPdexv3WaitingContributions(stateDB)
	if err != nil {
		return nil, err
	}
	poolPairs, err := initPoolPairStatesFromDB(stateDB)
	if err != nil {
		return nil, err
	}
	stakingPools, err := initStakingPoolsFromDB(params.StakingPoolsShare, stateDB)
	if err != nil {
		return nil, err
	}
	return newStateV2WithValue(
		waitingContributions, make(map[string]rawdbv2.Pdexv3Contribution),
		poolPairs, params, stakingPools, map[string]uint64{},
	), nil
}
