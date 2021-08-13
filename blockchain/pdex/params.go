package pdex

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type Params struct {
	DefaultFeeRateBPS               uint            // the default value if fee rate is not specific in FeeRateBPS (default 0.3% ~ 30 BPS)
	FeeRateBPS                      map[string]uint // map: pool ID -> fee rate (0.1% ~ 10 BPS)
	PRVDiscountPercent              uint            // percent of fee that will be discounted if using PRV as the trading token fee (default: 25%)
	LimitProtocolFeePercent         uint            // percent of fees from limit orders
	LimitStakingPoolRewardPercent   uint            // percent of fees from limit orders
	TradingProtocolFeePercent       uint            // percent of fees that is rewarded for the core team (default: 0%)
	TradingStakingPoolRewardPercent uint            // percent of fees that is distributed for staking pools (PRV, PDEX, ..., default: 10%)
	PDEXRewardPoolPairsShare        map[string]uint // map: pool pair ID -> PDEX reward share weight
	StakingPoolsShare               map[string]uint // map: staking tokenID -> pool staking share weight
}

func NewParams() *Params {
	return &Params{
		DefaultFeeRateBPS:               InitFeeRateBPS,
		FeeRateBPS:                      map[string]uint{},
		PRVDiscountPercent:              InitPRVDiscountPercent,
		LimitProtocolFeePercent:         InitProtocolFeePercent,
		LimitStakingPoolRewardPercent:   InitStakingPoolRewardPercent,
		TradingProtocolFeePercent:       InitProtocolFeePercent,
		TradingStakingPoolRewardPercent: InitStakingPoolRewardPercent,
		PDEXRewardPoolPairsShare:        map[string]uint{},
		StakingPoolsShare:               map[string]uint{},
	}
}

func NewParamsWithValue(paramsState *statedb.Pdexv3Params) *Params {
	return &Params{
		DefaultFeeRateBPS:               paramsState.DefaultFeeRateBPS(),
		FeeRateBPS:                      paramsState.FeeRateBPS(),
		PRVDiscountPercent:              paramsState.PRVDiscountPercent(),
		LimitProtocolFeePercent:         paramsState.LimitProtocolFeePercent(),
		LimitStakingPoolRewardPercent:   paramsState.LimitStakingPoolRewardPercent(),
		TradingProtocolFeePercent:       paramsState.TradingProtocolFeePercent(),
		TradingStakingPoolRewardPercent: paramsState.TradingStakingPoolRewardPercent(),
		PDEXRewardPoolPairsShare:        paramsState.PDEXRewardPoolPairsShare(),
		StakingPoolsShare:               paramsState.StakingPoolsShare(),
	}
}

func (p *Params) Clone() *Params {
	result := NewParams()
	result = p

	clonedFeeRateBPS := map[string]uint{}
	for k, v := range p.FeeRateBPS {
		clonedFeeRateBPS[k] = v
	}
	clonedPDEXRewardPoolPairsShare := map[string]uint{}
	for k, v := range p.PDEXRewardPoolPairsShare {
		clonedPDEXRewardPoolPairsShare[k] = v
	}
	clonedStakingPoolsShare := map[string]uint{}
	for k, v := range p.StakingPoolsShare {
		clonedStakingPoolsShare[k] = v
	}
	result.FeeRateBPS = clonedFeeRateBPS
	result.PDEXRewardPoolPairsShare = clonedPDEXRewardPoolPairsShare
	result.StakingPoolsShare = clonedStakingPoolsShare

	return result
}

func isValidPdexv3Params(
	params *Params,
	pairs map[string]*PoolPairState,
	stakingPools map[string]*StakingPoolState,
) (bool, string) {
	if params.DefaultFeeRateBPS > MaxFeeRateBPS {
		return false, "Default fee rate is too high"
	}
	for pairID, feeRate := range params.FeeRateBPS {
		_, isExisted := pairs[pairID]
		if !isExisted {
			return false, fmt.Sprintf("Pair %v is not existed", pairID)
		}
		if feeRate > MaxFeeRateBPS {
			return false, fmt.Sprintf("Fee rate of pair %v is too high", pairID)
		}
	}
	if params.PRVDiscountPercent > MaxPRVDiscountPercent {
		return false, "PRV discount percent is too high"
	}
	if params.TradingStakingPoolRewardPercent+params.TradingProtocolFeePercent > 100 {
		return false, "Sum of trading's staking pool + protocol fee is invalid"
	}
	if params.LimitProtocolFeePercent+params.LimitStakingPoolRewardPercent > 100 {
		return false, "Sum of limit order's staking pool + protocol fee is invalid"
	}
	for pairID := range params.PDEXRewardPoolPairsShare {
		_, isExisted := pairs[pairID]
		if !isExisted {
			return false, fmt.Sprintf("Pair %v is not existed", pairID)
		}
	}
	for stakingPoolID := range params.StakingPoolsShare {
		_, isExisted := stakingPools[stakingPoolID]
		if !isExisted {
			return false, fmt.Sprintf("Staking pool %v is not existed", stakingPoolID)
		}
	}
	return true, ""
}
