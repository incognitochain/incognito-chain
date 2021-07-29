package pdex

import (
	"math/big"
	"reflect"
	"sort"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

type PoolPairState struct {
	state  rawdbv2.Pdexv3PoolPair
	shares map[string]Share
}

func initPoolPairState(contribution0, contribution1 rawdbv2.Pdexv3Contribution) *PoolPairState {
	contributions := []rawdbv2.Pdexv3Contribution{contribution0, contribution1}
	sort.Slice(contributions, func(i, j int) bool {
		return contribution0.TokenID().String() < contribution1.TokenID().String()
	})
	token0VirtualAmount, token1VirtualAmount := calculateVirtualAmount(
		contributions[0].Amount(),
		contributions[1].Amount(),
		contributions[0].Amplifier(),
	)
	poolPairState := rawdbv2.NewPdexv3PoolPairWithValue(
		contributions[0].TokenID(), contributions[1].TokenID(),
		0, contributions[0].Amount(), contributions[1].Amount(),
		0,
		*token0VirtualAmount, *token1VirtualAmount,
		contributions[0].Amplifier(),
	)

	return NewPoolPairStateWithValue(
		*poolPairState,
		make(map[string]Share),
	)
}

func NewPoolPairState() *PoolPairState {
	return &PoolPairState{
		shares: make(map[string]Share),
		state:  *rawdbv2.NewPdexv3PoolPair(),
	}
}

func NewPoolPairStateWithValue(
	state rawdbv2.Pdexv3PoolPair,
	shares map[string]Share,
) *PoolPairState {
	return &PoolPairState{
		state:  state,
		shares: shares,
	}
}

func (p *PoolPairState) getContributionsByOrder(
	contribution0, contribution1 *rawdbv2.Pdexv3Contribution,
) (
	rawdbv2.Pdexv3Contribution, rawdbv2.Pdexv3Contribution,
) {
	if contribution0.TokenID() == p.state.Token0ID() {
		return *contribution0, *contribution1
	}
	return *contribution1, *contribution0
}

func (p *PoolPairState) computeActualContributedAmounts(
	contribution0, contribution1 *rawdbv2.Pdexv3Contribution,
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

func (p *PoolPairState) updateReserveData(amount0, amount1, shareAmount uint64) {
	if p.state.Amplifier() != metadataPdexv3.BaseAmplifier {
		tempShareAmount := p.state.ShareAmount() + shareAmount
		state := p.state
		token0VirtualAmount := state.Token0VirtualAmount()
		token1VirtualAmount := state.Token1VirtualAmount()
		tempToken0VirtualAmount := big.NewInt(0).Mul(
			&token0VirtualAmount,
			big.NewInt(0).SetUint64(tempShareAmount),
		)
		tempToken0VirtualAmount = tempToken0VirtualAmount.Div(
			tempToken0VirtualAmount,
			big.NewInt(0).SetUint64(state.ShareAmount()),
		)
		tempToken1VirtualAmount := big.NewInt(0).Mul(
			&token1VirtualAmount,
			big.NewInt(0).SetUint64(tempShareAmount),
		)
		tempToken1VirtualAmount = tempToken1VirtualAmount.Div(
			tempToken1VirtualAmount,
			big.NewInt(0).SetUint64(state.ShareAmount()),
		)
		if tempToken0VirtualAmount.Uint64() > amount0 {
			amount0 = tempToken0VirtualAmount.Uint64()
		}
		if tempToken1VirtualAmount.Uint64() > amount0 {
			amount1 = tempToken1VirtualAmount.Uint64()
		}
	}
	oldToken0VirtualAmount := p.state.Token0VirtualAmount()
	newToken0VirtualAmount := big.NewInt(0).Add(
		&oldToken0VirtualAmount,
		big.NewInt(0).SetUint64(amount0),
	)

	oldToken1VirtualAmount := p.state.Token1VirtualAmount()
	newToken1VirtualAmount := big.NewInt(0)
	newToken1VirtualAmount.Add(
		&oldToken1VirtualAmount,
		big.NewInt(0).SetUint64(amount1),
	)
	p.state.SetToken0VirtualAmount(*newToken0VirtualAmount)
	p.state.SetToken1VirtualAmount(*newToken1VirtualAmount)
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

	shareAmount := p.calculateShareAmount(amount0, amount1)
	p.state.SetToken0RealAmount(p.state.Token0RealAmount() + amount0)
	p.state.SetToken1RealAmount(p.state.Token1RealAmount() + amount1)
	p.updateReserveData(amount0, amount1, shareAmount)
	return shareAmount

}

func (p *PoolPairState) addShare(key string, amount, beaconHeight uint64) common.Hash {
	nfctID := genNFT(key, p.state.CurrentContributionID(), beaconHeight)
	share := NewShareWithValue(amount, make(map[string]uint64), beaconHeight)
	p.shares[nfctID.String()] = *share
	p.state.SetCurrentContributionID(p.state.CurrentContributionID() + 1)
	p.state.SetShareAmount(p.state.ShareAmount() + amount)
	return nfctID
}

func genNFT(key string, id, beaconHeight uint64) common.Hash {
	hash := append([]byte(key), append(common.Uint64ToBytes(id), common.Uint64ToBytes(beaconHeight)...)...)
	return common.HashH(hash)
}

func (p *PoolPairState) Clone() PoolPairState {
	res := NewPoolPairState()
	res.state = *p.state.Clone()
	for k, v := range p.shares {
		res.shares[k] = *v.Clone()
	}
	return *res
}

func (p *PoolPairState) getDiff(poolPairID string, comparePoolPair *PoolPairState, stateChange *StateChange) *StateChange {
	newStateChange := stateChange
	for k, v := range p.shares {
		if m, ok := comparePoolPair.shares[k]; !ok || !reflect.DeepEqual(m, v) {
			newStateChange = v.getDiff(k, &m, newStateChange)
		}
	}
	if !reflect.DeepEqual(p.state, comparePoolPair.state) {
		newStateChange.poolPairIDs[poolPairID] = true
	}
	return newStateChange
}

func (p *PoolPairState) calculateShareAmount(amount0, amount1 uint64) uint64 {
	state := p.state
	liquidityToken0 := big.NewInt(0).Mul(
		big.NewInt(0).SetUint64(amount0),
		big.NewInt(0).SetUint64(state.ShareAmount()),
	)
	liquidityToken0 = liquidityToken0.Div(
		liquidityToken0,
		big.NewInt(0).SetUint64(state.Token0RealAmount()),
	)
	liquidityToken1 := big.NewInt(0).Mul(
		big.NewInt(0).SetUint64(amount1),
		big.NewInt(0).SetUint64(state.ShareAmount()),
	)
	liquidityToken1 = liquidityToken1.Div(
		liquidityToken1,
		big.NewInt(0).SetUint64(state.Token1RealAmount()),
	)
	if liquidityToken0.Uint64() < liquidityToken1.Uint64() {
		return liquidityToken0.Uint64()
	}
	return liquidityToken1.Uint64()
}
