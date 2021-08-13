package pdex

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"sort"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

type PoolPairState struct {
	state     rawdbv2.Pdexv3PoolPair
	shares    map[string]*Share
	orderbook Orderbook
}

func (poolPairState *PoolPairState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		State  *rawdbv2.Pdexv3PoolPair `json:"State"`
		Shares map[string]*Share       `json:"Shares"`
	}{
		State:  &poolPairState.state,
		Shares: poolPairState.shares,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (poolPairState *PoolPairState) UnmarshalJSON(data []byte) error {
	temp := struct {
		State  *rawdbv2.Pdexv3PoolPair `json:"State"`
		Shares map[string]*Share       `json:"Shares"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	poolPairState.shares = temp.Shares
	if temp.State != nil {
		poolPairState.state = *temp.State
	}
	return nil
}

func initPoolPairState(contribution0, contribution1 rawdbv2.Pdexv3Contribution) *PoolPairState {
	cloneContribution0 := contribution0.Clone()
	cloneContribution1 := contribution1.Clone()

	contributions := []rawdbv2.Pdexv3Contribution{*cloneContribution0, *cloneContribution1}
	sort.Slice(contributions, func(i, j int) bool {
		return contributions[i].TokenID().String() < contributions[j].TokenID().String()
	})
	token0VirtualAmount, token1VirtualAmount := calculateVirtualAmount(
		contributions[0].Amount(),
		contributions[1].Amount(),
		contributions[0].Amplifier(),
	)
	poolPairState := rawdbv2.NewPdexv3PoolPairWithValue(
		contributions[0].TokenID(), contributions[1].TokenID(),
		0, contributions[0].Amount(), contributions[1].Amount(),
		token0VirtualAmount, token1VirtualAmount,
		contributions[0].Amplifier(),
	)
	return NewPoolPairStateWithValue(
		*poolPairState,
		make(map[string]*Share),
		Orderbook{[]*Order{}},
	)
}

func NewPoolPairState() *PoolPairState {
	return &PoolPairState{
		shares:    make(map[string]*Share),
		state:     *rawdbv2.NewPdexv3PoolPair(),
		orderbook: Orderbook{[]*Order{}},
	}
}

func NewPoolPairStateWithValue(
	state rawdbv2.Pdexv3PoolPair,
	shares map[string]*Share,
	orderbook Orderbook,
) *PoolPairState {
	return &PoolPairState{
		state:     state,
		shares:    shares,
		orderbook: orderbook,
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

//update both real and virtual amount
func (p *PoolPairState) addReserveData(amount0, amount1, shareAmount uint64) error {
	newToken0RealAmount := p.state.Token0RealAmount() + amount0
	newToken1RealAmount := p.state.Token1RealAmount() + amount1
	if newToken0RealAmount < p.state.Token0RealAmount() {
		return fmt.Errorf("newToken0RealAmount is out of range")
	}
	if newToken1RealAmount < p.state.Token1RealAmount() {
		return fmt.Errorf("newToken1RealAmount is out of range")
	}
	p.state.SetToken0RealAmount(newToken0RealAmount)
	p.state.SetToken1RealAmount(newToken1RealAmount)

	newToken0VirtualAmount := big.NewInt(0)
	newToken1VirtualAmount := big.NewInt(0)

	if p.state.Amplifier() != metadataPdexv3.BaseAmplifier {
		tempShareAmount := p.state.ShareAmount() + shareAmount
		if tempShareAmount < p.state.ShareAmount() {
			return fmt.Errorf("tempShareAmount is out of range")
		}
		token0VirtualAmount := p.state.Token0VirtualAmount()
		token1VirtualAmount := p.state.Token1VirtualAmount()
		tempToken0VirtualAmount := big.NewInt(0).Mul(
			token0VirtualAmount,
			big.NewInt(0).SetUint64(tempShareAmount),
		)
		tempToken0VirtualAmount = tempToken0VirtualAmount.Div(
			tempToken0VirtualAmount,
			big.NewInt(0).SetUint64(p.state.ShareAmount()),
		)
		tempToken1VirtualAmount := big.NewInt(0).Mul(
			token1VirtualAmount,
			big.NewInt(0).SetUint64(tempShareAmount),
		)
		tempToken1VirtualAmount = tempToken1VirtualAmount.Div(
			tempToken1VirtualAmount,
			big.NewInt(0).SetUint64(p.state.ShareAmount()),
		)
		if tempToken0VirtualAmount.Uint64() > p.state.Token0RealAmount() {
			newToken0VirtualAmount = tempToken0VirtualAmount
		} else {
			newToken0VirtualAmount.SetUint64(p.state.Token0RealAmount())
		}
		if tempToken1VirtualAmount.Uint64() > p.state.Token1RealAmount() {
			newToken1VirtualAmount = tempToken1VirtualAmount
		} else {
			newToken1VirtualAmount.SetUint64(p.state.Token1RealAmount())
		}
	} else {
		oldToken0VirtualAmount := p.state.Token0VirtualAmount()
		newToken0VirtualAmount = big.NewInt(0).Add(
			oldToken0VirtualAmount,
			big.NewInt(0).SetUint64(amount0),
		)

		oldToken1VirtualAmount := p.state.Token1VirtualAmount()
		newToken1VirtualAmount = big.NewInt(0)
		newToken1VirtualAmount.Add(
			oldToken1VirtualAmount,
			big.NewInt(0).SetUint64(amount1),
		)
	}
	p.state.SetToken0VirtualAmount(newToken0VirtualAmount)
	p.state.SetToken1VirtualAmount(newToken1VirtualAmount)
	return nil
}

func (p *PoolPairState) addReserveDataAndCalculateShare(
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
	p.addReserveData(amount0, amount1, shareAmount)
	return shareAmount

}

func (p *PoolPairState) addShare(
	nftID common.Hash, nftIDs map[string]bool,
	amount, beaconHeight uint64,
	txHash string,
) (common.Hash, map[string]bool, error) {
	newNftID := genNFT(nftID, nftIDs, beaconHeight)
	nftIDStr := chooseNftStr(nftID, newNftID)
	nftIDs[nftIDStr] = true
	var shareAmount uint64
	var newBeaconHeight uint64
	if p.shares[nftIDStr] == nil {
		shareAmount = amount
		newBeaconHeight = beaconHeight
	} else {
		shareAmount = p.shares[nftIDStr].amount + amount
		newBeaconHeight = p.shares[nftIDStr].lastUpdatedBeaconHeight
	}
	share := NewShareWithValue(shareAmount, make(map[string]uint64), newBeaconHeight)
	p.shares[nftIDStr] = share
	newShareAmount := p.state.ShareAmount() + amount
	if newShareAmount < p.state.ShareAmount() {
		return newNftID, nftIDs, fmt.Errorf("Share amount is out of range")
	}
	p.state.SetShareAmount(newShareAmount)
	return newNftID, nftIDs, nil
}

func (p *PoolPairState) Clone() *PoolPairState {
	res := NewPoolPairState()
	res.state = *p.state.Clone()
	for k, v := range p.shares {
		res.shares[k] = v.Clone()
	}
	//TODO: clone order book here
	return res
}

func (p *PoolPairState) getDiff(
	poolPairID string,
	comparePoolPair *PoolPairState,
	stateChange *StateChange,
) *StateChange {
	newStateChange := stateChange
	if comparePoolPair == nil {
		newStateChange.poolPairIDs[poolPairID] = true
		for nftID, share := range p.shares {
			newStateChange = share.getDiff(nftID, nil, newStateChange)
		}
	} else {
		if !reflect.DeepEqual(p.state, comparePoolPair.state) {
			newStateChange.poolPairIDs[poolPairID] = true
		}
		for nftID, share := range p.shares {
			if m, ok := comparePoolPair.shares[nftID]; !ok || !reflect.DeepEqual(m, share) {
				newStateChange = share.getDiff(nftID, m, newStateChange)
			}
		}
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

func (p *PoolPairState) deductReserveData(amount0, amount1, shareAmount uint64) error {
	newToken0RealAmount := p.state.Token0RealAmount() - amount0
	newToken1RealAmount := p.state.Token1RealAmount() - amount1
	if newToken0RealAmount >= p.state.Token0RealAmount() {
		return fmt.Errorf("newToken0RealAmount is out of range")
	}
	if newToken1RealAmount >= p.state.Token1RealAmount() {
		return fmt.Errorf("newToken1RealAmount is out of range")
	}
	p.state.SetToken0RealAmount(newToken0RealAmount)
	p.state.SetToken1RealAmount(newToken1RealAmount)
	newToken0VirtualAmount := big.NewInt(0)
	newToken1VirtualAmount := big.NewInt(0)

	if p.state.Amplifier() != metadataPdexv3.BaseAmplifier {
		tempShareAmount := p.state.ShareAmount() - shareAmount
		if tempShareAmount >= p.state.ShareAmount() {
			return fmt.Errorf("tempShareAmount is out of range")
		}
		token0VirtualAmount := p.state.Token0VirtualAmount()
		token1VirtualAmount := p.state.Token1VirtualAmount()
		tempToken0VirtualAmount := big.NewInt(0).Mul(
			token0VirtualAmount,
			big.NewInt(0).SetUint64(tempShareAmount),
		)
		tempToken0VirtualAmount = tempToken0VirtualAmount.Div(
			tempToken0VirtualAmount,
			big.NewInt(0).SetUint64(p.state.ShareAmount()),
		)
		tempToken1VirtualAmount := big.NewInt(0).Mul(
			token1VirtualAmount,
			big.NewInt(0).SetUint64(tempShareAmount),
		)
		tempToken1VirtualAmount = tempToken1VirtualAmount.Div(
			tempToken1VirtualAmount,
			big.NewInt(0).SetUint64(p.state.ShareAmount()),
		)
		if tempToken0VirtualAmount.Uint64() > p.state.Token0RealAmount() {
			newToken0VirtualAmount = tempToken0VirtualAmount
		} else {
			newToken0VirtualAmount.SetUint64(p.state.Token0RealAmount())
		}
		if tempToken1VirtualAmount.Uint64() > p.state.Token1RealAmount() {
			newToken1VirtualAmount = tempToken1VirtualAmount
		} else {
			newToken1VirtualAmount.SetUint64(p.state.Token1RealAmount())
		}
	} else {
		oldToken0VirtualAmount := p.state.Token0VirtualAmount()
		newToken0VirtualAmount = big.NewInt(0).Sub(
			oldToken0VirtualAmount,
			big.NewInt(0).SetUint64(amount0),
		)

		oldToken1VirtualAmount := p.state.Token1VirtualAmount()
		newToken1VirtualAmount = big.NewInt(0)
		newToken1VirtualAmount.Sub(
			oldToken1VirtualAmount,
			big.NewInt(0).SetUint64(amount1),
		)
	}
	p.state.SetToken0VirtualAmount(newToken0VirtualAmount)
	p.state.SetToken1VirtualAmount(newToken1VirtualAmount)
	return nil
}

func (p *PoolPairState) deductShare(
	nftID string,
	shareAmount uint64,
) (uint64, uint64, uint64, error) {
	share := p.shares[nftID]
	if shareAmount == 0 || share.amount == 0 {
		return 0, 0, 0, errors.New("shareAmount = 0 or share.amount = 0")
	}
	tempShareAmount := shareAmount
	if share.amount < shareAmount {
		tempShareAmount = share.amount
	}

	token0Amount := big.NewInt(0)
	token0Amount = token0Amount.Mul(
		big.NewInt(0).SetUint64(p.state.Token0RealAmount()),
		big.NewInt(0).SetUint64(tempShareAmount),
	)
	token0Amount = token0Amount.Div(token0Amount, big.NewInt(0).SetUint64(p.state.ShareAmount()))
	token1Amount := big.NewInt(0)
	token1Amount = token1Amount.Mul(
		big.NewInt(0).SetUint64(p.state.Token1RealAmount()),
		big.NewInt(0).SetUint64(tempShareAmount),
	)
	token1Amount = token1Amount.Div(token1Amount, big.NewInt(0).SetUint64(p.state.ShareAmount()))
	err := p.deductReserveData(token0Amount.Uint64(), token1Amount.Uint64(), tempShareAmount)
	if err != nil {
		return 0, 0, 0, errors.New("shareAmount = 0 or share.amount = 0")
	}
	share.amount -= tempShareAmount
	if share.amount >= tempShareAmount {
		return token0Amount.Uint64(), token1Amount.Uint64(), tempShareAmount, errors.New("share.amount is our of range")
	}
	return token0Amount.Uint64(), token1Amount.Uint64(), tempShareAmount, nil
}
