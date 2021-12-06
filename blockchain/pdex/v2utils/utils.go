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

func GetMakingVolumes(
	pairChange [2]*big.Int, orderChange map[string][2]*big.Int,
	nftIDs map[string]string,
) (*big.Int, map[string]*big.Int) {
	ammMakingAmt := getMakingAmountFromChange(pairChange)

	orderMakingAmts := map[string]*big.Int{}
	for ordID, change := range orderChange {
		orderMakingAmt := getMakingAmountFromChange(change)
		nftID := nftIDs[ordID]
		if _, ok := orderMakingAmts[nftID]; !ok {
			orderMakingAmts[nftID] = new(big.Int).SetUint64(0)
		}
		orderMakingAmts[nftID].Add(orderMakingAmts[nftID], orderMakingAmt)
	}

	return ammMakingAmt, orderMakingAmts
}

func SplitTradingReward(
	reward *big.Int, ratio uint, bps uint,
	ammMakingAmt *big.Int, orderMakingAmts map[string]*big.Int,
) (uint64, map[string]uint64) {
	if ratio == 0 {
		return reward.Uint64(), map[string]uint64{}
	}
	weightedMakingAmt := new(big.Int).SetUint64(0)

	weightedAmmMakingAmt := new(big.Int).Mul(ammMakingAmt, new(big.Int).SetUint64(uint64(bps)))
	weightedMakingAmt.Add(weightedMakingAmt, weightedAmmMakingAmt)

	weightedOrderMakingAmts := map[string]*big.Int{}
	for nftID, amt := range orderMakingAmts {
		weight := new(big.Int).Mul(amt, new(big.Int).SetUint64(uint64(ratio)))
		weightedOrderMakingAmts[nftID] = weight
		weightedMakingAmt.Add(weightedMakingAmt, weight)
	}

	ammReward := new(big.Int).SetUint64(0)
	if weightedAmmMakingAmt.Cmp(new(big.Int).SetUint64(0)) > 0 {
		ammReward = new(big.Int).Mul(reward, weightedAmmMakingAmt)
		ammReward.Div(ammReward, weightedMakingAmt)
	}

	weightedMakingAmt.Sub(weightedMakingAmt, weightedAmmMakingAmt)
	reward.Sub(reward, ammReward)

	orderRewards := map[string]uint64{}
	for nftID, weight := range weightedOrderMakingAmts {
		orderReward := new(big.Int).SetUint64(0)
		if weight.Cmp(new(big.Int).SetUint64(0)) > 0 {
			orderReward = new(big.Int).Mul(reward, weight)
			orderReward.Div(orderReward, weightedMakingAmt)
		}

		orderRewards[nftID] = orderReward.Uint64()

		weightedMakingAmt.Sub(weightedMakingAmt, weight)
		reward.Sub(reward, orderReward)
	}

	return ammReward.Uint64(), orderRewards
}
