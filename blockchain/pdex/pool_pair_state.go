package pdex

import (
	"sort"

	"github.com/incognitochain/incognito-chain/utils"
)

type PoolPairState struct {
	token0ID              string
	token1ID              string
	token0RealAmount      uint64
	token1RealAmount      uint64
	shares                map[string]uint64
	tradingFees           map[string]map[string]uint64
	currentContributionID uint64
	token0VirtualAmount   uint64
	token1VirtualAmount   uint64
	amplifier             uint
}

func initPoolPairState(contribution0, contribution1 Contribution) *PoolPairState {
	contributions := []Contribution{contribution0, contribution1}
	sort.Slice(contributions, func(i, j int) bool {
		return contributions[i].tokenID < contributions[j].tokenID
	})
	token0VirtualAmount, token1VirtualAmount := calculateVirtualAmount(
		contributions[0].tokenAmount,
		contributions[1].tokenAmount,
		contributions[0].amplifier,
	)
	return NewPoolPairStateWithValue(
		contributions[0].tokenID, contributions[1].tokenID,
		contributions[0].tokenAmount, contributions[1].tokenAmount,
		token0VirtualAmount, token1VirtualAmount,
		0,
		contributions[0].amplifier,
		make(map[string]uint64),
		make(map[string]map[string]uint64),
	)
}

func NewPoolPairState() *PoolPairState {
	return &PoolPairState{
		shares:      make(map[string]uint64),
		tradingFees: make(map[string]map[string]uint64),
	}
}

func NewPoolPairStateWithValue(
	token0ID, token1ID string,
	token0RealAmount, token1RealAmount,
	token0VirtualAmount, token1VirtualAmount,
	currentContributionID uint64,
	amplifier uint,
	shares map[string]uint64,
	tradingFees map[string]map[string]uint64,
) *PoolPairState {
	return &PoolPairState{
		token0ID:              token0ID,
		token1ID:              token1ID,
		token0RealAmount:      token0RealAmount,
		token1RealAmount:      token1RealAmount,
		token0VirtualAmount:   token0VirtualAmount,
		token1VirtualAmount:   token1VirtualAmount,
		currentContributionID: currentContributionID,
		amplifier:             amplifier,
		shares:                shares,
		tradingFees:           tradingFees,
	}
}

func (p *PoolPairState) getContributionsByOrder(
	contribution0, contribution1 Contribution,
) (Contribution, Contribution) {
	if contribution0.tokenID == p.token0ID {
		return contribution0, contribution1
	}
	return contribution1, contribution0
}

func (p *PoolPairState) computeActualContributedAmounts(
	contribution0, contribution1 Contribution,
) (uint64, uint64, uint64, uint64) {
	return 0, 0, 0, 0
}

func (p *PoolPairState) addContributions(contribution0, contribution1 Contribution) (string, error) {
	nfctID, err := p.addShare(contribution0.tokenAmount)
	if err != nil {
		return utils.EmptyString, err
	}
	return nfctID, nil
}

func (p *PoolPairState) addShare(tokenAmount uint64) (string, error) {
	res := utils.EmptyString
	return res, nil
}
