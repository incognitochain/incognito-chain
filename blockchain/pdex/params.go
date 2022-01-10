package pdex

import (
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type Params struct {
	DefaultFeeRateBPS                 uint            // the default value if fee rate is not specific in FeeRateBPS (default 0.3% ~ 30 BPS)
	FeeRateBPS                        map[string]uint // map: pool ID -> fee rate (0.1% ~ 10 BPS)
	PRVDiscountPercent                uint            // percent of fee that will be discounted if using PRV as the trading token fee (default: 25%)
	TradingProtocolFeePercent         uint            // percent of fees that is rewarded for the core team (default: 0%)
	TradingStakingPoolRewardPercent   uint            // percent of fees that is distributed for staking pools (PRV, PDEX, ..., default: 10%)
	PDEXRewardPoolPairsShare          map[string]uint // map: pool pair ID -> PDEX reward share weight
	StakingPoolsShare                 map[string]uint // map: staking tokenID -> pool staking share weight
	StakingRewardTokens               []common.Hash   // list of staking reward tokens
	MintNftRequireAmount              uint64          // amount prv for depositing to pdex
	MaxOrdersPerNft                   uint            // max orders per nft
	AutoWithdrawOrderLimitAmount      uint            // max orders will be auto withdraw each shard for each blocks
	MinPRVReserveTradingRate          uint64          // min prv reserve for checking price of trading fee paid by PRV
	DefaultOrderTradingRewardRatioBPS uint            // default order trading reward ratio BPS
	OrderTradingRewardRatioBPS        map[string]uint // map: pool ID -> the weight of LOP rewards compared with LP rewards (0.1% ~ 10 BPS)
	OrderLiquidityMiningBPS           map[string]uint // map: pool ID -> order liquidity mining BPS (1500 BPS -> the percent of reward for AMM : buy orders : sell orders will be 70% : 15% : 15%)
	DAOContributingPercent            uint            // percent of DAO reward that is using for liquidity mining (default: 0%)
	MiningRewardPendingBlocks         uint64          // liquidity mining reward of each contribution will be pending for this number of blocks (default: 0)
	OrderMiningRewardRatioBPS         map[string]uint // not used yet
}

func NewParams() *Params {
	return &Params{
		FeeRateBPS:                 map[string]uint{},
		PDEXRewardPoolPairsShare:   map[string]uint{},
		StakingPoolsShare:          map[string]uint{},
		StakingRewardTokens:        []common.Hash{},
		OrderTradingRewardRatioBPS: map[string]uint{},
		OrderLiquidityMiningBPS:    map[string]uint{},
		OrderMiningRewardRatioBPS:  map[string]uint{},
	}
}

func NewParamsWithValue(paramsState *statedb.Pdexv3Params) *Params {
	return &Params{
		DefaultFeeRateBPS:                 paramsState.DefaultFeeRateBPS(),
		FeeRateBPS:                        paramsState.FeeRateBPS(),
		PRVDiscountPercent:                paramsState.PRVDiscountPercent(),
		TradingProtocolFeePercent:         paramsState.TradingProtocolFeePercent(),
		TradingStakingPoolRewardPercent:   paramsState.TradingStakingPoolRewardPercent(),
		PDEXRewardPoolPairsShare:          paramsState.PDEXRewardPoolPairsShare(),
		StakingPoolsShare:                 paramsState.StakingPoolsShare(),
		StakingRewardTokens:               paramsState.StakingRewardTokens(),
		MintNftRequireAmount:              paramsState.MintNftRequireAmount(),
		MaxOrdersPerNft:                   paramsState.MaxOrdersPerNft(),
		AutoWithdrawOrderLimitAmount:      paramsState.AutoWithdrawOrderLimitAmount(),
		MinPRVReserveTradingRate:          paramsState.MinPRVReserveTradingRate(),
		DefaultOrderTradingRewardRatioBPS: paramsState.DefaultOrderTradingRewardRatioBPS(),
		OrderTradingRewardRatioBPS:        paramsState.OrderTradingRewardRatioBPS(),
		OrderLiquidityMiningBPS:           paramsState.OrderLiquidityMiningBPS(),
		DAOContributingPercent:            paramsState.DAOContributingPercent(),
		MiningRewardPendingBlocks:         paramsState.MiningRewardPendingBlocks(),
		OrderMiningRewardRatioBPS:         paramsState.OrderMiningRewardRatioBPS(),
	}
}

func (p *Params) Clone() *Params {
	result := &Params{}
	*result = *p

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
	clonedOrderTradingRewardRatioBPS := map[string]uint{}
	for k, v := range p.OrderTradingRewardRatioBPS {
		clonedOrderTradingRewardRatioBPS[k] = v
	}
	clonedOrderLiquidityMiningBPS := map[string]uint{}
	for k, v := range p.OrderLiquidityMiningBPS {
		clonedOrderLiquidityMiningBPS[k] = v
	}
	clonedOrderMiningRewardRatioBPS := map[string]uint{}
	for k, v := range p.OrderMiningRewardRatioBPS {
		clonedOrderMiningRewardRatioBPS[k] = v
	}
	result.FeeRateBPS = clonedFeeRateBPS
	result.PDEXRewardPoolPairsShare = clonedPDEXRewardPoolPairsShare
	result.StakingPoolsShare = clonedStakingPoolsShare
	result.OrderTradingRewardRatioBPS = clonedOrderTradingRewardRatioBPS
	result.OrderLiquidityMiningBPS = clonedOrderLiquidityMiningBPS
	result.OrderMiningRewardRatioBPS = clonedOrderMiningRewardRatioBPS

	return result
}

func isValidPdexv3Params(
	params *Params,
	pairs map[string]*PoolPairState,
) (bool, string) {
	if params.DefaultFeeRateBPS > MaxFeeRateBPS {
		return false, "Default fee rate is too high"
	}
	if params.DefaultFeeRateBPS == 0 {
		return false, "Default fee rate is 0"
	}
	for pairID, feeRate := range params.FeeRateBPS {
		_, isExisted := pairs[pairID]
		if !isExisted {
			return false, fmt.Sprintf("Pair %v is not existed", pairID)
		}
		if feeRate > MaxFeeRateBPS {
			return false, fmt.Sprintf("Fee rate of pair %v is too high", pairID)
		}
		if feeRate == 0 {
			return false, fmt.Sprintf("Fee rate of pair %v is 0", pairID)
		}
	}
	if params.PRVDiscountPercent > MaxPRVDiscountPercent {
		return false, "PRV discount percent is too high"
	}
	if params.TradingStakingPoolRewardPercent+params.TradingProtocolFeePercent > 100 {
		return false, "Sum of trading's staking pool + protocol fee is invalid"
	}
	for pairID := range params.PDEXRewardPoolPairsShare {
		_, isExisted := pairs[pairID]
		if !isExisted {
			return false, fmt.Sprintf("Pair %v is not existed", pairID)
		}
	}
	for stakingPoolID := range params.StakingPoolsShare {
		_, err := common.Hash{}.NewHashFromStr(stakingPoolID)
		if err != nil {
			return false, fmt.Sprintf("%v", err)
		}
	}
	if params.DefaultOrderTradingRewardRatioBPS > BPS {
		return false, "Default order trading reward ratio is too high"
	}
	for pairID, ratioBPS := range params.OrderTradingRewardRatioBPS {
		_, isExisted := pairs[pairID]
		if !isExisted {
			return false, fmt.Sprintf("Pair %v does not exist", pairID)
		}
		if ratioBPS > BPS {
			return false, fmt.Sprintf("Order trading reward ratio of pair %v is too high", pairID)
		}
	}
	for pairID, ratioBPS := range params.OrderLiquidityMiningBPS {
		_, isExisted := pairs[pairID]
		if !isExisted {
			return false, fmt.Sprintf("Pair %v does not exist", pairID)
		}
		if ratioBPS >= BPS/2 {
			return false, fmt.Sprintf("Order liquidity mining BPS of pair %v is too high", pairID)
		}
	}
	if params.DAOContributingPercent > 100 {
		return false, "DAO contributing percent is too high"
	}
	if params.DAOContributingPercent > 0 && len(params.PDEXRewardPoolPairsShare) == 0 {
		return false, "Liquidity mining is enabled but no reward pool is specified"
	}
	return true, ""
}

func (params *Params) IsZeroValue() bool {
	return reflect.DeepEqual(params, NewParams()) || params == nil
}

func (params *Params) readConfig() *Params {
	res := &Params{
		DefaultFeeRateBPS:               config.Param().PDexParams.Params.DefaultFeeRateBPS,
		PRVDiscountPercent:              config.Param().PDexParams.Params.PRVDiscountPercent,
		TradingProtocolFeePercent:       config.Param().PDexParams.Params.TradingProtocolFeePercent,
		TradingStakingPoolRewardPercent: config.Param().PDexParams.Params.TradingStakingPoolRewardPercent,
		StakingPoolsShare:               config.Param().PDexParams.Params.StakingPoolsShare,
		MintNftRequireAmount:            config.Param().PDexParams.Params.MintNftRequireAmount,
		MaxOrdersPerNft:                 config.Param().PDexParams.Params.MaxOrdersPerNft,
		AutoWithdrawOrderLimitAmount:    config.Param().PDexParams.Params.AutoWithdrawOrderLimitAmount,
		MinPRVReserveTradingRate:        config.Param().PDexParams.Params.MinPRVReserveTradingRate,
	}
	if res.FeeRateBPS == nil {
		res.FeeRateBPS = make(map[string]uint)
	}
	if res.StakingPoolsShare == nil {
		res.StakingPoolsShare = make(map[string]uint)
	}
	if res.PDEXRewardPoolPairsShare == nil {
		res.PDEXRewardPoolPairsShare = make(map[string]uint)
	}
	if res.OrderTradingRewardRatioBPS == nil {
		res.OrderTradingRewardRatioBPS = make(map[string]uint)
	}
	if res.OrderLiquidityMiningBPS == nil {
		res.OrderLiquidityMiningBPS = make(map[string]uint)
	}
	if res.OrderMiningRewardRatioBPS == nil {
		res.OrderMiningRewardRatioBPS = make(map[string]uint)
	}
	return res
}
