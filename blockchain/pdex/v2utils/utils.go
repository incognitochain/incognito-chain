package v2utils

import (
	"math/big"
	"sort"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

type MintNftStatus struct {
	NftID       string `json:"NftID"`
	Status      byte   `json:"Status"`
	BurntAmount uint64 `json:"BurntAmount"`
}

type ContributionStatus struct {
	Status                  byte         `json:"Status"`
	Token0ID                string       `json:"Token0ID"`
	Token0ContributedAmount uint64       `json:"Token0ContributedAmount"`
	Token0ReturnedAmount    uint64       `json:"Token0ReturnedAmount"`
	Token1ID                string       `json:"Token1ID"`
	Token1ContributedAmount uint64       `json:"Token1ContributedAmount"`
	Token1ReturnedAmount    uint64       `json:"Token1ReturnedAmount"`
	PoolPairID              string       `json:"PoolPairID"`
	AccessID                *common.Hash `json:"AccessID,omitempty"`
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
) (*big.Int, map[string]*big.Int, byte) {
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

	sellToken0 := pairChange[0].Cmp(big.NewInt(0)) > 0
	for _, change := range orderChange {
		if change[0].Cmp(big.NewInt(0)) > 0 {
			sellToken0 = true
			break
		}
	}

	tradeDirection := TradeDirectionSell1
	if sellToken0 {
		tradeDirection = TradeDirectionSell0
	}

	return ammMakingAmt, orderMakingAmts, byte(tradeDirection)
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

	// To store the keys in slice in sorted order
	keys := make([]string, len(weightedOrderMakingAmts))
	i := 0
	for k := range weightedOrderMakingAmts {
		keys[i] = k
		i++
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	orderRewards := map[string]uint64{}
	for _, nftID := range keys {
		weight := weightedOrderMakingAmts[nftID]
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

func SplitOrderRewardLiquidityMining(
	volume map[string]*big.Int, amount *big.Int, tokenID common.Hash,
) map[string]uint64 {
	sumVolume := new(big.Int).SetUint64(0)
	for _, v := range volume {
		sumVolume.Add(sumVolume, v)
	}

	// To store the keys in slice in sorted order
	keys := make([]string, len(volume))
	i := 0
	for k := range volume {
		keys[i] = k
		i++
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	orderRewards := map[string]uint64{}
	remain := new(big.Int).Set(amount)
	for _, nftID := range keys {
		v := volume[nftID]
		orderReward := new(big.Int).SetUint64(0)
		if v.Cmp(new(big.Int).SetUint64(0)) > 0 {
			orderReward = new(big.Int).Mul(remain, v)
			orderReward.Div(orderReward, sumVolume)
		}

		orderRewards[nftID] = orderReward.Uint64()

		sumVolume.Sub(sumVolume, v)
		remain.Sub(remain, orderReward)
	}

	return orderRewards
}

type NFTAssetTagsCache map[string]*common.Hash

func (m *NFTAssetTagsCache) FromIDs(nftIDs map[string]uint64) (*NFTAssetTagsCache, error) {
	var result NFTAssetTagsCache
	if m == nil {
		result = make(map[string]*common.Hash)
	} else {
		result = *m
	}
	for idStr, _ := range nftIDs {
		tokenID, err := common.Hash{}.NewHashFromStr(idStr)
		if err != nil {
			return nil, err
		}
		assetTag := privacy.HashToPoint(tokenID[:])
		result[assetTag.String()] = tokenID
	}
	return &result, nil
}

func (m *NFTAssetTagsCache) Add(id common.Hash) {
	if m != nil {
		assetTag := privacy.HashToPoint(id[:])
		(*m)[assetTag.String()] = &id
	}
}
