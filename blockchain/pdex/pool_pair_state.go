package pdex

import (
	"math/big"
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataPdexV3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

type PoolPairState struct {
	state       statedb.Pdexv3PoolPairState
	shares      map[string]statedb.Pdexv3ShareState
	tradingFees map[string]map[string]uint64
}

func initPoolPairState(contribution0, contribution1 statedb.Pdexv3ContributionState) *PoolPairState {
	contributions := []statedb.Pdexv3ContributionState{contribution0, contribution1}
	sort.Slice(contributions, func(i, j int) bool {
		return contributions[i].TokenID() < contributions[j].TokenID()
	})
	token0VirtualAmount, token1VirtualAmount := calculateVirtualAmount(
		contributions[0].Amount(),
		contributions[1].Amount(),
		contributions[0].Amplifier(),
	)
	state := statedb.NewPdexv3PoolPairStateWithValue(
		contributions[0].TokenID(), contributions[1].TokenID(),
		contributions[0].Amount(), contributions[1].Amount(),
		0,
		token0VirtualAmount, token1VirtualAmount,
		contributions[0].Amplifier(),
	)
	return NewPoolPairStateWithValue(
		*state,
		make(map[string]statedb.Pdexv3ShareState),
		make(map[string]map[string]uint64),
	)
}

func NewPoolPairState() *PoolPairState {
	return &PoolPairState{
		shares:      make(map[string]statedb.Pdexv3ShareState),
		tradingFees: make(map[string]map[string]uint64),
	}
}

func NewPoolPairStateWithValue(
	state statedb.Pdexv3PoolPairState,
	shares map[string]statedb.Pdexv3ShareState,
	tradingFees map[string]map[string]uint64,
) *PoolPairState {
	return &PoolPairState{
		state:       state,
		shares:      shares,
		tradingFees: tradingFees,
	}
}

func (p *PoolPairState) getContributionsByOrder(
	contribution0, contribution1 *statedb.Pdexv3ContributionState,
	metaData0, metaData1 *metadataPdexV3.AddLiquidity,
) (
	statedb.Pdexv3ContributionState, statedb.Pdexv3ContributionState,
	metadataPdexV3.AddLiquidity, metadataPdexV3.AddLiquidity,
) {
	if contribution0.TokenID() == p.state.Token0ID() {
		return *contribution0, *contribution1, *metaData0, *metaData1
	}
	return *contribution1, *contribution0, *metaData1, *metaData0
}

func (p *PoolPairState) computeActualContributedAmounts(
	contribution0, contribution1 *statedb.Pdexv3ContributionState,
) (uint64, uint64, uint64, uint64) {
	contribution0Amount := big.NewInt(0)
	tempAmt := big.NewInt(0)
	tempAmt.Mul(
		new(big.Int).SetUint64(contribution1.Amount()),
		new(big.Int).SetUint64(p.state.Token0RealAmount()),
	)
	tempAmt.Div(
		tempAmt,
		new(big.Int).SetUint64(p.state.Token1RealAmount()),
	)
	if tempAmt.Uint64() > contribution0.Amount() {
		contribution0Amount = new(big.Int).SetUint64(contribution0.Amount())
	} else {
		contribution0Amount = tempAmt
	}
	contribution1Amount := big.NewInt(0)
	contribution1Amount.Mul(
		contribution0Amount,
		new(big.Int).SetUint64(p.state.Token1RealAmount()),
	)
	contribution1Amount.Div(
		contribution1Amount,
		new(big.Int).SetUint64(p.state.Token0RealAmount()),
	)
	actualContribution0Amt := contribution0Amount.Uint64()
	actualContribution1Amt := contribution1Amount.Uint64()
	return actualContribution0Amt, contribution0.Amount() - actualContribution0Amt, actualContribution1Amt, contribution1.Amount() - actualContribution1Amt
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
	p.state.SetToken0RealAmount(p.state.Token0RealAmount() + amount0)
	p.state.SetToken1RealAmount(p.state.Token1RealAmount() + amount1)
	p.state.SetToken0VirtualAmount(p.state.Token0VirtualAmount() + amount0)
	p.state.SetToken1VirtualAmount(p.state.Token1VirtualAmount() + amount1)
	return amount0

}

func (p *PoolPairState) addShare(key string, amount, beaconHeight uint64) string {
	nfctID := genNFT(key, p.state.CurrentContributionID(), beaconHeight)
	share := statedb.NewPdexv3ShareStateWithValue(amount)
	p.shares[nfctID] = *share
	p.state.SetCurrentContributionID(p.state.CurrentContributionID() + 1)
	return nfctID
}

func genNFT(key string, id, beaconHeight uint64) string {
	hash := key + strconv.FormatUint(id, 10) + strconv.FormatUint(beaconHeight, 10)
	return common.HashH([]byte(hash)).String()
}

func (p *PoolPairState) Clone() PoolPairState {
	res := NewPoolPairState()
	res.state = *p.state.Clone()
	for k, v := range p.shares {
		res.shares[k] = v
	}
	for k, v := range p.tradingFees {
		res.tradingFees[k] = make(map[string]uint64)
		for key, value := range v {
			res.tradingFees[k][key] = value
		}
	}
	return *res
}
