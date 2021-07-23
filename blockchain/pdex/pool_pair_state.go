package pdex

import (
	"math/big"
	"sort"

	metadataPdexV3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
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
	contribution0, contribution1 *Contribution,
	metaData0, metaData1 *metadataPdexV3.AddLiquidity,
) (
	Contribution, Contribution,
	metadataPdexV3.AddLiquidity, metadataPdexV3.AddLiquidity,
) {
	if contribution0.tokenID == p.token0ID {
		return *contribution0, *contribution1, *metaData0, *metaData1
	}
	return *contribution1, *contribution0, *metaData1, *metaData0
}

func (p *PoolPairState) computeActualContributedAmounts(
	contribution0, contribution1 Contribution,
) (uint64, uint64, uint64, uint64) {
	contribution0Amount := big.NewInt(0)
	tempAmt := big.NewInt(0)
	tempAmt.Mul(
		new(big.Int).SetUint64(contribution1.tokenAmount),
		new(big.Int).SetUint64(p.token0RealAmount),
	)
	tempAmt.Div(
		tempAmt,
		new(big.Int).SetUint64(p.token1RealAmount),
	)
	if tempAmt.Uint64() > contribution0.tokenAmount {
		contribution0Amount = new(big.Int).SetUint64(contribution0.tokenAmount)
	} else {
		contribution0Amount = tempAmt
	}
	contribution1Amount := big.NewInt(0)
	contribution1Amount.Mul(
		contribution0Amount,
		new(big.Int).SetUint64(p.token1RealAmount),
	)
	contribution1Amount.Div(
		contribution1Amount,
		new(big.Int).SetUint64(p.token0RealAmount),
	)
	actualContribution0Amt := contribution0Amount.Uint64()
	actualContribution1Amt := contribution1Amount.Uint64()
	return actualContribution0Amt, contribution0.tokenAmount - actualContribution0Amt, actualContribution1Amt, contribution1.tokenAmount - actualContribution1Amt
}

func (p *PoolPairState) addContributions(contribution0, contribution1 Contribution) (string, error) {
	contributions := []Contribution{contribution0, contribution1}
	sort.Slice(contributions, func(i, j int) bool {
		return contributions[i].tokenID < contributions[j].tokenID
	})

	p.token0RealAmount += contributions[0].tokenAmount
	p.token1RealAmount += contributions[1].tokenAmount
	p.token0VirtualAmount += contributions[0].tokenAmount
	p.token1VirtualAmount += contributions[1].tokenAmount

	nfctID, err := p.addShare(contributions[0].tokenAmount)
	if err != nil {
		return utils.EmptyString, err
	}
	p.currentContributionID++
	return nfctID, nil
}

func (p *PoolPairState) addShare(tokenAmount uint64) (string, error) {
	res := utils.EmptyString
	//TODO: @tin
	return res, nil
}
