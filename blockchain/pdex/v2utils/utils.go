package v2utils

import (
	"math/big"
)

type MintNftStatus struct {
	NftID       string `json:"NftID"`
	Status      byte   `json:"Status"`
	BurntAmount uint64 `json:"BurntAmount"`
}

type ContributionStatus struct {
	Status                  byte   `json:"Status"`
	Token0ID                string `json:"Token0ID"`
	Token0ContributedAmount uint64 `json:"Token0ContributedAmount"`
	Token0ReturnedAmount    uint64 `json:"Token0ReturnedAmount"`
	Token1ID                string `json:"Token1ID"`
	Token1ContributedAmount uint64 `json:"Token1ContributedAmount"`
	Token1ReturnedAmount    uint64 `json:"Token1ReturnedAmount"`
	PoolPairID              string `json:"PoolPairID"`
}

type WithdrawStatus struct {
	Status       byte   `json:"Status"`
	Token0ID     string `json:"Token0ID"`
	Token0Amount uint64 `json:"Token0Amount"`
	Token1ID     string `json:"Token1ID"`
	Token1Amount uint64 `json:"Token1Amount"`
}

type StakingStatus struct {
	Status        byte   `json:"Status"`
	NftID         string `json:"NftID"`
	StakingPoolID string `json:"StakingPoolID"`
	Liquidity     uint64 `json:"Liquidity"`
}

type UnstakingStatus struct {
	Status        byte   `json:"Status"`
	NftID         string `json:"NftID"`
	StakingPoolID string `json:"StakingPoolID"`
	Liquidity     uint64 `json:"Liquidity"`
}

func getMakingAmountFromChange(change [2]*big.Int) *big.Int {
	if change[0].Cmp(big.NewInt(0)) < 0 {
		return new(big.Int).Neg(change[0])
	}
	return new(big.Int).Neg(change[1])
}

func SplitTradingReward(
	reward *big.Int, ratio uint, bps uint,
	pairChange [2]*big.Int, orderChange map[string][2]*big.Int,
) (uint64, map[string]uint64) {
	weightedMakingAmt := new(big.Int).SetUint64(0)

	ammMakingAmt := getMakingAmountFromChange(pairChange)
	weightedAmmMakingAmt := new(big.Int).Mul(ammMakingAmt, new(big.Int).SetUint64(uint64(bps)))
	weightedMakingAmt.Add(weightedMakingAmt, weightedAmmMakingAmt)

	weightOrderMakingAmt := map[string]*big.Int{}
	for ordID, change := range orderChange {
		orderMakingAmt := getMakingAmountFromChange(change)
		weightOrderMakingAmt[ordID] = new(big.Int).Mul(orderMakingAmt, new(big.Int).SetUint64(uint64(2*ratio)))
		weightedMakingAmt.Add(weightedMakingAmt, weightOrderMakingAmt[ordID])
	}

	ammReward := new(big.Int).SetUint64(0)
	if weightedAmmMakingAmt.Cmp(new(big.Int).SetUint64(0)) > 0 {
		ammReward = new(big.Int).Mul(reward, weightedAmmMakingAmt)
		ammReward.Div(ammReward, weightedMakingAmt)
	}

	weightedMakingAmt.Sub(weightedMakingAmt, weightedAmmMakingAmt)
	reward.Sub(reward, ammReward)

	orderRewards := map[string]uint64{}
	for ordID := range orderChange {
		orderReward := new(big.Int).SetUint64(0)
		if weightOrderMakingAmt[ordID].Cmp(new(big.Int).SetUint64(0)) > 0 {
			orderReward = new(big.Int).Mul(reward, weightOrderMakingAmt[ordID])
			orderReward.Div(orderReward, weightedMakingAmt)
		}

		orderRewards[ordID] = orderReward.Uint64()

		weightedMakingAmt.Sub(weightedMakingAmt, weightOrderMakingAmt[ordID])
		reward.Sub(reward, orderReward)
	}

	return ammReward.Uint64(), orderRewards
}
