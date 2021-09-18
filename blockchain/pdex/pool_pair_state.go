package pdex

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"sort"

	"github.com/incognitochain/incognito-chain/blockchain/pdex/v2utils"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

type PoolPairState struct {
	state           rawdbv2.Pdexv3PoolPair
	shares          map[string]*Share
	orderbook       Orderbook
	lpFeesPerShare  map[common.Hash]*big.Int
	protocolFees    map[common.Hash]uint64
	stakingPoolFees map[common.Hash]uint64
}

func NewPoolPairState() *PoolPairState {
	return &PoolPairState{
		shares:          make(map[string]*Share),
		state:           *rawdbv2.NewPdexv3PoolPair(),
		orderbook:       Orderbook{[]*Order{}},
		lpFeesPerShare:  make(map[common.Hash]*big.Int),
		protocolFees:    make(map[common.Hash]uint64),
		stakingPoolFees: make(map[common.Hash]uint64),
	}
}

func NewPoolPairStateWithValue(
	state rawdbv2.Pdexv3PoolPair,
	shares map[string]*Share,
	orderbook Orderbook,
	lpFeesPerShare map[common.Hash]*big.Int,
	protocolFees, stakingPoolFees map[common.Hash]uint64,
) *PoolPairState {
	return &PoolPairState{
		state:           state,
		shares:          shares,
		orderbook:       orderbook,
		lpFeesPerShare:  lpFeesPerShare,
		protocolFees:    protocolFees,
		stakingPoolFees: stakingPoolFees,
	}
}

func (poolPairState *PoolPairState) State() rawdbv2.Pdexv3PoolPair {
	return poolPairState.state
}

func (poolPairState *PoolPairState) LpFeesPerShare() map[common.Hash]*big.Int {
	res := make(map[common.Hash]*big.Int)
	for k, v := range poolPairState.lpFeesPerShare {
		res[k] = big.NewInt(0).SetBytes(v.Bytes())
	}
	return res
}

func (poolPairState *PoolPairState) ProtocolFees() map[common.Hash]uint64 {
	res := make(map[common.Hash]uint64)
	for k, v := range poolPairState.protocolFees {
		res[k] = v
	}
	return res
}

func (poolPairState *PoolPairState) StakingPoolFees() map[common.Hash]uint64 {
	res := make(map[common.Hash]uint64)
	for k, v := range poolPairState.stakingPoolFees {
		res[k] = v
	}
	return res
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
		State           *rawdbv2.Pdexv3PoolPair  `json:"State"`
		Shares          map[string]*Share        `json:"Shares"`
		Orderbook       Orderbook                `json:"Orderbook"`
		LpFeesPerShare  map[common.Hash]*big.Int `json:"LpFeesPerShare"`
		ProtocolFees    map[common.Hash]uint64   `json:"ProtocolFees"`
		StakingPoolFees map[common.Hash]uint64   `json:"StakingPoolFees"`
	}{
		State:           &poolPairState.state,
		Shares:          poolPairState.shares,
		Orderbook:       poolPairState.orderbook,
		LpFeesPerShare:  poolPairState.lpFeesPerShare,
		ProtocolFees:    poolPairState.protocolFees,
		StakingPoolFees: poolPairState.stakingPoolFees,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (poolPairState *PoolPairState) UnmarshalJSON(data []byte) error {
	temp := struct {
		State           *rawdbv2.Pdexv3PoolPair  `json:"State"`
		Shares          map[string]*Share        `json:"Shares"`
		Orderbook       Orderbook                `json:"Orderbook"`
		LpFeesPerShare  map[common.Hash]*big.Int `json:"LpFeesPerShare"`
		ProtocolFees    map[common.Hash]uint64   `json:"ProtocolFees"`
		StakingPoolFees map[common.Hash]uint64   `json:"StakingPoolFees"`
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
	poolPairState.lpFeesPerShare = temp.LpFeesPerShare
	poolPairState.protocolFees = temp.ProtocolFees
	poolPairState.stakingPoolFees = temp.StakingPoolFees
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
		make(map[common.Hash]*big.Int),
		make(map[common.Hash]uint64), make(map[common.Hash]uint64),
	)
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
	return p.updateShareValue(amount, beaconHeight, nftID.String(), addOperator)
}

func (p *PoolPairState) Clone() *PoolPairState {
	res := NewPoolPairState()
	res.state = *p.state.Clone()
	for k, v := range p.shares {
		res.shares[k] = v.Clone()
	}
	for k, v := range p.lpFeesPerShare {
		res.lpFeesPerShare[k] = big.NewInt(0).Set(v)
	}
	for k, v := range p.protocolFees {
		res.protocolFees[k] = v
	}
	for k, v := range p.stakingPoolFees {
		res.stakingPoolFees[k] = v
	}

	res.orderbook = p.orderbook.Clone()
	return res
}

func (p *PoolPairState) getDiff(
	poolPairID string, comparePoolPair *PoolPairState,
	poolPairChange *v2utils.PoolPairChange,
	stateChange *v2utils.StateChange,
) (*v2utils.PoolPairChange, *v2utils.StateChange) {
	newPoolPairChange := poolPairChange
	newStateChange := stateChange
	if comparePoolPair == nil {
		newPoolPairChange.IsChanged = true
		for nftID, share := range p.shares {
			shareChange := v2utils.NewShareChange()
			shareChange = share.getDiff(nftID, nil, shareChange)
			poolPairChange.Shares[nftID] = shareChange
		}
		for tokenID := range p.lpFeesPerShare {
			newPoolPairChange.LpFeesPerShare[tokenID.String()] = true
		}
		for tokenID := range p.protocolFees {
			newPoolPairChange.ProtocolFees[tokenID.String()] = true
		}
		for tokenID := range p.stakingPoolFees {
			newPoolPairChange.StakingPoolFees[tokenID.String()] = true
		}
		for _, ord := range p.orderbook.orders {
			newPoolPairChange.OrderIDs[ord.Id()] = true
		}
	} else {
		if !reflect.DeepEqual(p.state, comparePoolPair.state) {
			newPoolPairChange.IsChanged = true
		}
		for nftID, share := range p.shares {
			if m, ok := comparePoolPair.shares[nftID]; !ok || !reflect.DeepEqual(m, share) {
				shareChange := v2utils.NewShareChange()
				shareChange = share.getDiff(nftID, m, shareChange)
				poolPairChange.Shares[nftID] = shareChange
			}
		}
		for tokenID, value := range p.lpFeesPerShare {
			if m, ok := comparePoolPair.lpFeesPerShare[tokenID]; !ok || !reflect.DeepEqual(m, value) {
				newPoolPairChange.LpFeesPerShare[tokenID.String()] = true
			}
		}
		for tokenID, value := range p.protocolFees {
			if m, ok := comparePoolPair.protocolFees[tokenID]; !ok || !reflect.DeepEqual(m, value) {
				newPoolPairChange.ProtocolFees[tokenID.String()] = true
			}
		}
		for tokenID, value := range p.stakingPoolFees {
			if m, ok := comparePoolPair.stakingPoolFees[tokenID]; !ok || !reflect.DeepEqual(m, value) {
				newPoolPairChange.StakingPoolFees[tokenID.String()] = true
			}
		}
		newPoolPairChange = p.orderbook.getDiff(&comparePoolPair.orderbook, newPoolPairChange)
	}
	return newPoolPairChange, newStateChange
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
	shareAmount, beaconHeight uint64,
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
	err = p.updateShareValue(tempShareAmount, beaconHeight, nftID, subOperator)
	return token0Amount.Uint64(), token1Amount.Uint64(), tempShareAmount, err
}

func (p *PoolPairState) updateShareValue(
	shareAmount, beaconHeight uint64, nftID string, operator byte,
) error {
	share, found := p.shares[nftID]
	if !found {
		if operator == subOperator {
			return errors.New("Deduct nil share amount")
		}
		share = NewShare()
	} else {
		nftIDBytes, err := common.Hash{}.NewHashFromStr(nftID)
		if err != nil {
			return fmt.Errorf("Invalid nftID: %s", nftID)
		}
		share.tradingFees, err = p.RecomputeLPFee(*nftIDBytes)
		if err != nil {
			return fmt.Errorf("Error when tracking LP reward: %v\n", err)
		}
	}

	share.lastLPFeesPerShare = map[common.Hash]*big.Int{}
	for tokenID, value := range p.lpFeesPerShare {
		share.lastLPFeesPerShare[tokenID] = new(big.Int).Set(value)
	}

	var err error
	share.amount, err = executeOperationUint64(share.amount, shareAmount, operator)
	if err != nil {
		return errors.New("newShare.amount is out of range")
	}

	poolPairShareAmount, err := executeOperationUint64(p.state.ShareAmount(), shareAmount, operator)
	if err != nil {
		return errors.New("poolPairShareAmount is out of range")
	}
	p.state.SetShareAmount(poolPairShareAmount)

	p.shares[nftID] = share
	return nil
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

func (p *PoolPairState) RecomputeLPFee(
	nftID common.Hash,
) (map[common.Hash]uint64, error) {
	result := map[common.Hash]uint64{}

	curShare, ok := p.shares[nftID.String()]
	if !ok {
		return nil, fmt.Errorf("Share not found")
	}

	curLPFeesPerShare := p.lpFeesPerShare
	oldLPFeesPerShare := curShare.lastLPFeesPerShare

	for tokenID := range curLPFeesPerShare {
		tradingFee, isExisted := curShare.tradingFees[tokenID]
		if !isExisted {
			tradingFee = 0
		}
		oldFees, isExisted := oldLPFeesPerShare[tokenID]
		if !isExisted {
			oldFees = big.NewInt(0)
		}
		newFees := curLPFeesPerShare[tokenID]

		reward := new(big.Int).Mul(new(big.Int).Sub(newFees, oldFees), new(big.Int).SetUint64(curShare.amount))
		reward = new(big.Int).Div(reward, BaseLPFeesPerShare)
		reward = new(big.Int).Add(reward, new(big.Int).SetUint64(tradingFee))

		if !reward.IsUint64() {
			return nil, fmt.Errorf("Reward of token %v is out of range", tokenID)
		}
		if reward.Uint64() > 0 {
			result[tokenID] = reward.Uint64()
		}
	}
	return result, nil
}

func (p *PoolPairState) withState(state rawdbv2.Pdexv3PoolPair) {
	p.state = state
}

func (p *PoolPairState) withShares(shares map[string]*Share) {
	p.shares = shares
}

func (p *PoolPairState) withOrderBook(orderbook Orderbook) {
	p.orderbook = orderbook
}

func (p *PoolPairState) withLpFeesPerShare(lpFeesPerShare map[common.Hash]*big.Int) {
	p.lpFeesPerShare = lpFeesPerShare
}

func (p *PoolPairState) withProtocolFees(protocolFees map[common.Hash]uint64) {
	p.protocolFees = protocolFees
}

func (p *PoolPairState) withStakingPoolFees(stakingPoolFees map[common.Hash]uint64) {
	p.stakingPoolFees = stakingPoolFees
}

func (p *PoolPairState) cloneShare(nftID string) map[string]*Share {
	res := make(map[string]*Share)
	for k, v := range p.shares {
		if k == nftID {
			res[k] = v.Clone()
		} else {
			res[k] = v
		}
	}
	return res
}

func (p *PoolPairState) updateToDB(
	env StateEnvironment, poolPairID string, poolPairChange *v2utils.PoolPairChange,
) error {
	var err error
	if poolPairChange.IsChanged {
		err = statedb.StorePdexv3PoolPair(env.StateDB(), poolPairID, p.state)
		if err != nil {
			return err
		}
	}
	for nftID, share := range p.shares {
		shareChange, found := poolPairChange.Shares[nftID]
		if !found || shareChange == nil {
			continue
		}
		err := share.updateToDB(env, poolPairID, nftID, shareChange)
		if err != nil {
			return err
		}
	}
	for tokenID, value := range p.lpFeesPerShare {
		if poolPairChange.LpFeesPerShare[tokenID.String()] {
			statedb.StorePdexv3PoolPairLpFeePerShare(
				env.StateDB(), poolPairID,
				statedb.NewPdexv3PoolPairLpFeePerShareStateWithValue(tokenID, value),
			)
		}
	}
	for tokenID, value := range p.protocolFees {
		if poolPairChange.ProtocolFees[tokenID.String()] {
			statedb.StorePdexv3PoolPairProtocolFee(
				env.StateDB(), poolPairID,
				statedb.NewPdexv3PoolPairProtocolFeeStateWithValue(tokenID, value),
			)
		}
	}
	for tokenID, value := range p.stakingPoolFees {
		if poolPairChange.StakingPoolFees[tokenID.String()] {
			statedb.StorePdexv3PoolPairStakingPoolFee(
				env.StateDB(), poolPairID,
				statedb.NewPdexv3PoolPairStakingPoolFeeStateWithValue(tokenID, value),
			)
		}
	}
	return nil
}

func initPoolPairStatesFromDB(stateDB *statedb.StateDB) (map[string]*PoolPairState, error) {
	poolPairsStates, err := statedb.GetPdexv3PoolPairs(stateDB)
	if err != nil {
		return nil, err
	}
	res := make(map[string]*PoolPairState)
	for poolPairID, poolPairState := range poolPairsStates {
		lpFeesPerShare, err := statedb.GetPdexv3PoolPairLpFeesPerShares(stateDB, poolPairID)
		if err != nil {
			return nil, err
		}
		protocolFees, err := statedb.GetPdexv3PoolPairProtocolFees(stateDB, poolPairID)
		if err != nil {
			return nil, err
		}
		stakingPoolFees, err := statedb.GetPdexv3PoolPairStakingPoolFees(stateDB, poolPairID)
		if err != nil {
			return nil, err
		}
		shares, err := initShares(poolPairID, stateDB)
		if err != nil {
			return nil, err
		}

		orderbook := &Orderbook{[]*Order{}}
		orderMap, err := statedb.GetPdexv3Orders(stateDB, poolPairState.PoolPairID())
		if err != nil {
			return nil, err
		}
		for _, item := range orderMap {
			v := item.Value()
			orderbook.InsertOrder(&v)
		}
		poolPair := NewPoolPairStateWithValue(
			poolPairState.Value(), shares, *orderbook,
			lpFeesPerShare, protocolFees, stakingPoolFees,
		)
		res[poolPairID] = poolPair
	}
	return res, nil
}
