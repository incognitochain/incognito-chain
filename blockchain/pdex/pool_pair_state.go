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

func (poolPairState *PoolPairState) State() rawdbv2.Pdexv3PoolPair {
	return poolPairState.state
}

func (poolPairState *PoolPairState) Shares() map[string]*Share {
	res := make(map[string]*Share)
	for k, v := range poolPairState.shares {
		res[k] = NewShare()
		*res[k] = *v
	}
	return res
}

func (poolPairState *PoolPairState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		State     *rawdbv2.Pdexv3PoolPair `json:"State"`
		Shares    map[string]*Share       `json:"Shares"`
		Orderbook Orderbook               `json:"Orderbook"`
	}{
		State:     &poolPairState.state,
		Shares:    poolPairState.shares,
		Orderbook: poolPairState.orderbook,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (poolPairState *PoolPairState) UnmarshalJSON(data []byte) error {
	temp := struct {
		State     *rawdbv2.Pdexv3PoolPair `json:"State"`
		Shares    map[string]*Share       `json:"Shares"`
		Orderbook Orderbook               `json:"Orderbook"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	poolPairState.shares = temp.Shares
	if temp.State != nil {
		poolPairState.state = *temp.State
	}
	poolPairState.orderbook = temp.Orderbook
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

func (p *PoolPairState) addReserveDataAndCalculateShare(
	token0ID, token1ID string,
	token0Amount, token1Amount uint64,
) (uint64, error) {
	var amount0, amount1 uint64
	if token0ID < token1ID {
		amount0 = token0Amount
		amount1 = token1Amount
	} else {
		amount0 = token1Amount
		amount1 = token0Amount
	}
	shareAmount := p.calculateShareAmount(amount0, amount1)
	err := p.updateReserveData(amount0, amount1, shareAmount, addOperator)
	return shareAmount, err

}

func (p *PoolPairState) addShare(
	nftID common.Hash,
	amount, beaconHeight uint64,
	txHash string,
) error {
	var shareAmount uint64
	var newBeaconHeight uint64
	if p.shares[nftID.String()] == nil {
		shareAmount = amount
		newBeaconHeight = beaconHeight
	} else {
		shareAmount = p.shares[nftID.String()].amount + amount
		newBeaconHeight = p.shares[nftID.String()].lastUpdatedBeaconHeight
	}
	share := NewShareWithValue(shareAmount, make(map[string]uint64), newBeaconHeight)
	p.shares[nftID.String()] = share
	newShareAmount := p.state.ShareAmount() + amount
	if newShareAmount < p.state.ShareAmount() {
		return fmt.Errorf("Share amount is out of range")
	}
	p.state.SetShareAmount(newShareAmount)
	return nil
}

func (p *PoolPairState) Clone() *PoolPairState {
	res := NewPoolPairState()
	res.state = *p.state.Clone()
	for k, v := range p.shares {
		res.shares[k] = v.Clone()
	}
	res.orderbook = p.orderbook.Clone()
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
		newStateChange = p.orderbook.getDiff(&comparePoolPair.orderbook, newStateChange)
	}
	return newStateChange
}

func (p *PoolPairState) calculateShareAmount(amount0, amount1 uint64) uint64 {
	return CalculateShareAmount(
		p.state.Token0RealAmount(),
		p.state.Token1RealAmount(),
		amount0, amount1, p.state.ShareAmount(),
	)
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
	err := p.updateReserveData(token0Amount.Uint64(), token1Amount.Uint64(), tempShareAmount, subOperator)
	if err != nil {
		return 0, 0, 0, errors.New("shareAmount = 0 or share.amount = 0")
	}
	p.shares[nftID], err = p.updateShareAmount(tempShareAmount, share, subOperator)
	return token0Amount.Uint64(), token1Amount.Uint64(), tempShareAmount, err
}

func (p *PoolPairState) updateShareAmount(shareAmount uint64, share *Share, operator byte) (*Share, error) {
	newShare := share
	var err error
	newShare.amount, err = executeOperationUint64(newShare.amount, shareAmount, operator)
	if err != nil {
		return newShare, errors.New("newShare.amount is out of range")
	}
	poolPairShareAmount, err := executeOperationUint64(p.state.ShareAmount(), shareAmount, operator)
	if err != nil {
		return newShare, errors.New("poolPairShareAmount is out of range")
	}
	p.state.SetShareAmount(poolPairShareAmount)
	return newShare, nil
}

func (p *PoolPairState) updateReserveData(amount0, amount1, shareAmount uint64, operator byte) error {
	err := p.updateSingleTokenAmount(p.state.Token0ID(), amount0, shareAmount, operator)
	if err != nil {
		return err
	}
	err = p.updateSingleTokenAmount(p.state.Token1ID(), amount1, shareAmount, operator)
	if err != nil {
		return err
	}
	return nil
}

func (p *PoolPairState) updateSingleTokenAmount(
	tokenID common.Hash,
	amount, shareAmount uint64,
	operator byte,
) error {
	var realAmount uint64
	virtualAmount := big.NewInt(0)
	switch tokenID.String() {
	case p.state.Token0ID().String():
		realAmount = p.state.Token0RealAmount()
		virtualAmount = p.state.Token0VirtualAmount()
	case p.state.Token1ID().String():
		realAmount = p.state.Token1RealAmount()
		virtualAmount = p.state.Token1VirtualAmount()
	default:
		return errors.New("Can't find tokenID")
	}
	tempShareAmount, err := executeOperationUint64(p.state.ShareAmount(), shareAmount, operator)
	if err != nil {
		return err
	}
	newRealAmount, err := executeOperationUint64(realAmount, amount, operator)
	if err != nil {
		return err
	}
	newVirtualAmount := big.NewInt(0)
	if p.state.Amplifier() != metadataPdexv3.BaseAmplifier {
		tempVirtualAmount := big.NewInt(0).Mul(
			virtualAmount,
			big.NewInt(0).SetUint64(tempShareAmount),
		)
		tempVirtualAmount = tempVirtualAmount.Div(
			tempVirtualAmount,
			big.NewInt(0).SetUint64(p.state.ShareAmount()),
		)
		if tempVirtualAmount.Uint64() > newRealAmount {
			newVirtualAmount = tempVirtualAmount
		} else {
			newVirtualAmount.SetUint64(newRealAmount)
		}
	} else {
		newVirtualAmount, err = executeOperationBigInt(virtualAmount, big.NewInt(0).SetUint64(amount), operator)
	}
	switch tokenID.String() {
	case p.state.Token0ID().String():
		p.state.SetToken0RealAmount(newRealAmount)
		p.state.SetToken0VirtualAmount(newVirtualAmount)
	case p.state.Token1ID().String():
		p.state.SetToken1RealAmount(newRealAmount)
		p.state.SetToken1VirtualAmount(newVirtualAmount)
	}
	return nil
}
