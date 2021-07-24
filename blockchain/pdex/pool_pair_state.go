package pdex

import (
	"math/big"
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	metadataPdexV3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
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

func (p *PoolPairState) updateReserveAndShares(
	token0ID, token1ID string,
	token0Amount, token1Amount uint64,
) uint64 {
	var amount0, amount1 uint64
	if token0ID < token1ID {
		amount0 = token0Amount
		amount1 = token1Amount
	} else {
		amount0 = token1Amount
		amount1 = token0Amount
	}
	p.token0RealAmount += amount0
	p.token1RealAmount += amount1
	p.token0VirtualAmount += amount0
	p.token1VirtualAmount += amount1
	return amount0

}

func (p *PoolPairState) addShare(key string, amount, beaconHeight uint64) string {
	p.genNFT(key, beaconHeight)
	res := p.genNFT(key, beaconHeight)
	p.shares[res] = amount
	p.currentContributionID++
	return res
}

func (p *PoolPairState) genNFT(key string, beaconHeight uint64) string {
	id := key + strconv.FormatUint(p.currentContributionID, 10) + strconv.FormatUint(beaconHeight, 10)
	return common.HashH([]byte(id)).String()
}
