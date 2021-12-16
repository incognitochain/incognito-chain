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
	"github.com/incognitochain/incognito-chain/utils"
)

type PoolPairState struct {
	makingVolume    map[common.Hash]*MakingVolume // tokenID -> MakingVolume
	state           rawdbv2.Pdexv3PoolPair
	shares          map[string]*Share
	orderRewards    map[string]*OrderReward // nftID -> orderReward
	orderbook       Orderbook
	lpFeesPerShare  map[common.Hash]*big.Int
	protocolFees    map[common.Hash]uint64
	stakingPoolFees map[common.Hash]uint64
}

func NewPoolPairState() *PoolPairState {
	return &PoolPairState{
		makingVolume:    make(map[common.Hash]*MakingVolume),
		orderRewards:    make(map[string]*OrderReward),
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
	makingVolume map[common.Hash]*MakingVolume,
	orderRewards map[string]*OrderReward,
) *PoolPairState {
	return &PoolPairState{
		makingVolume:    makingVolume,
		orderRewards:    orderRewards,
		state:           state,
		shares:          shares,
		orderbook:       orderbook,
		lpFeesPerShare:  lpFeesPerShare,
		protocolFees:    protocolFees,
		stakingPoolFees: stakingPoolFees,
	}
}

func (poolPairState *PoolPairState) isEmpty() bool {
	if poolPairState.state.Token0RealAmount() == 0 || poolPairState.state.Token1RealAmount() == 0 {
		return true
	}
	if poolPairState.state.Token0VirtualAmount().Cmp(big.NewInt(0)) <= 0 || poolPairState.state.Token1VirtualAmount().Cmp(big.NewInt(0)) <= 0 {
		return true
	}
	if poolPairState.state.ShareAmount() == 0 {
		return true
	}
	return false
}

func (poolPairState *PoolPairState) State() rawdbv2.Pdexv3PoolPair {
	return poolPairState.state
}

func (poolPairState *PoolPairState) LpFeesPerShare() map[common.Hash]*big.Int {
	res := make(map[common.Hash]*big.Int)
	for k, v := range poolPairState.lpFeesPerShare {
		res[k] = big.NewInt(0).Set(v)
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
		res[k] = v.Clone()
	}
	return res
}

func (poolPairState *PoolPairState) OrderRewards() map[string]*OrderReward {
	res := make(map[string]*OrderReward)
	for k, v := range poolPairState.orderRewards {
		res[k] = v.Clone()
	}
	return res
}

func (poolPairState *PoolPairState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		State           *rawdbv2.Pdexv3PoolPair       `json:"State"`
		Shares          map[string]*Share             `json:"Shares,omitempty"`
		Orderbook       Orderbook                     `json:"Orderbook,omitempty"`
		LpFeesPerShare  map[common.Hash]*big.Int      `json:"LpFeesPerShare"`
		ProtocolFees    map[common.Hash]uint64        `json:"ProtocolFees"`
		StakingPoolFees map[common.Hash]uint64        `json:"StakingPoolFees"`
		OrderRewards    map[string]*OrderReward       `json:"OrderRewards"`
		MakingVolume    map[common.Hash]*MakingVolume `json:"MakingVolume"`
	}{
		State:           &poolPairState.state,
		Shares:          poolPairState.shares,
		Orderbook:       poolPairState.orderbook,
		LpFeesPerShare:  poolPairState.lpFeesPerShare,
		ProtocolFees:    poolPairState.protocolFees,
		StakingPoolFees: poolPairState.stakingPoolFees,
		OrderRewards:    poolPairState.orderRewards,
		MakingVolume:    poolPairState.makingVolume,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (poolPairState *PoolPairState) UnmarshalJSON(data []byte) error {
	temp := struct {
		State           *rawdbv2.Pdexv3PoolPair       `json:"State"`
		Shares          map[string]*Share             `json:"Shares"`
		Orderbook       Orderbook                     `json:"Orderbook"`
		LpFeesPerShare  map[common.Hash]*big.Int      `json:"LpFeesPerShare"`
		ProtocolFees    map[common.Hash]uint64        `json:"ProtocolFees"`
		StakingPoolFees map[common.Hash]uint64        `json:"StakingPoolFees"`
		OrderRewards    map[string]*OrderReward       `json:"OrderRewards"`
		MakingVolume    map[common.Hash]*MakingVolume `json:"MakingVolume"`
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
	poolPairState.orderRewards = temp.OrderRewards
	poolPairState.makingVolume = temp.MakingVolume
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
		make(map[common.Hash]*MakingVolume), make(map[string]*OrderReward),
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
) (uint64, uint64, uint64, uint64, error) {
	if p.isEmpty() {
		return 0, 0, 0, 0, errors.New("Pool is invalid to contribute")
	}
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
	if tempAmt.Cmp(big.NewInt(0).SetUint64(contribution0.Amount())) > 0 {
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
	if !contribution0Amount.IsUint64() || !contribution1Amount.IsUint64() {
		return 0, 0, 0, 0, errors.New("contributed amount is not uint64")
	}
	actualContribution0Amt := contribution0Amount.Uint64()
	actualContribution1Amt := contribution1Amount.Uint64()
	return actualContribution0Amt, contribution0.Amount() - actualContribution0Amt, actualContribution1Amt, contribution1.Amount() - actualContribution1Amt, nil
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
	txHash, accessOTA string,
) (string, error) {
	return p.updateShareValue(amount, beaconHeight, nftID.String(), accessOTA, addOperator)
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
	for k, v := range p.orderRewards {
		res.orderRewards[k] = v.Clone()
	}
	for k, v := range p.makingVolume {
		res.makingVolume[k] = v.Clone()
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
		for nftID, orderReward := range p.orderRewards {
			orderRewardChange := v2utils.NewOrderRewardChange()
			orderRewardChange = orderReward.getDiff(nftID, nil, orderRewardChange)
			poolPairChange.OrderRewards[nftID] = orderRewardChange
		}
		for tokenID, makingVolume := range p.makingVolume {
			makingVolumeChange := v2utils.NewMakingVolumeChange()
			makingVolumeChange = makingVolume.getDiff(tokenID.String(), nil, makingVolumeChange)
			poolPairChange.MakingVolume[tokenID.String()] = makingVolumeChange
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
		newPoolPairChange.LpFeesPerShare = v2utils.GetChangedElementsFromMapBigInt(p.lpFeesPerShare, comparePoolPair.lpFeesPerShare)
		newPoolPairChange.ProtocolFees = v2utils.GetChangedElementsFromMapUint64(p.protocolFees, comparePoolPair.protocolFees)
		newPoolPairChange.StakingPoolFees = v2utils.GetChangedElementsFromMapUint64(p.stakingPoolFees, comparePoolPair.stakingPoolFees)
		for nftID, orderReward := range p.orderRewards {
			if m, ok := comparePoolPair.orderRewards[nftID]; !ok || !reflect.DeepEqual(m, orderReward) {
				orderRewardChange := v2utils.NewOrderRewardChange()
				orderRewardChange = orderReward.getDiff(nftID, m, orderRewardChange)
				poolPairChange.OrderRewards[nftID] = orderRewardChange
			}
		}
		for tokenID, makingVolume := range p.makingVolume {
			if m, ok := comparePoolPair.makingVolume[tokenID]; !ok || !reflect.DeepEqual(m, makingVolume) {
				makingVolumeChange := v2utils.NewMakingVolumeChange()
				makingVolumeChange = makingVolume.getDiff(tokenID.String(), m, makingVolumeChange)
				poolPairChange.MakingVolume[tokenID.String()] = makingVolumeChange
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
	accessOption metadataPdexv3.AccessOption,
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
	_, err = p.updateShareValue(tempShareAmount, beaconHeight, nftID, utils.EmptyString, subOperator) //TODO: @tin fix here
	return token0Amount.Uint64(), token1Amount.Uint64(), tempShareAmount, err
}

func (p *PoolPairState) updateShareValue(
	shareAmount, beaconHeight uint64, nftID, accessOTA string, operator byte,
) (string, error) {
	share, found := p.shares[nftID]
	if !found {
		if operator == subOperator {
			return utils.EmptyString, errors.New("Deduct nil share amount")
		}
		share = NewShare()
	} else {
		nftIDBytes, err := common.Hash{}.NewHashFromStr(nftID)
		if err != nil {
			return utils.EmptyString, fmt.Errorf("Invalid nftID: %s", nftID)
		}
		share.tradingFees, err = p.RecomputeLPFee(*nftIDBytes)
		if err != nil {
			return utils.EmptyString, fmt.Errorf("Error when tracking LP reward: %v\n", err)
		}
		accessOTA = share.accessOTA
	}

	share.lastLPFeesPerShare = map[common.Hash]*big.Int{}
	for tokenID, value := range p.lpFeesPerShare {
		share.lastLPFeesPerShare[tokenID] = new(big.Int).Set(value)
	}

	var err error
	share.amount, err = executeOperationUint64(share.amount, shareAmount, operator)
	if err != nil {
		return utils.EmptyString, errors.New("newShare.amount is out of range")
	}
	if accessOTA == utils.EmptyString {
		accessOTA = share.accessOTA
	}
	share.accessOTA = accessOTA

	poolPairShareAmount, err := executeOperationUint64(p.state.ShareAmount(), shareAmount, operator)
	if err != nil {
		return utils.EmptyString, errors.New("poolPairShareAmount is out of range")
	}
	p.state.SetShareAmount(poolPairShareAmount)

	p.shares[nftID] = share
	return accessOTA, nil
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
		if tempVirtualAmount.Cmp(big.NewInt(0).SetUint64(newRealAmount)) > 0 {
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
	for nftID, shareChange := range poolPairChange.Shares {
		if shareChange.IsChanged {
			if share, found := p.shares[nftID]; found {
				err := share.updateToDB(env, poolPairID, nftID, shareChange)
				if err != nil {
					return err
				}
			} else {
				err := share.deleteFromDB(env, poolPairID, nftID, shareChange)
				if err != nil {
					return err
				}
			}
		}
	}

	for tokenID, value := range p.lpFeesPerShare {
		if poolPairChange.LpFeesPerShare[tokenID.String()] {
			err = statedb.StorePdexv3PoolPairLpFeePerShare(
				env.StateDB(), poolPairID,
				statedb.NewPdexv3PoolPairLpFeePerShareStateWithValue(tokenID, value),
			)
			if err != nil {
				return err
			}
		}
	}
	for tokenID, value := range p.protocolFees {
		if poolPairChange.ProtocolFees[tokenID.String()] {
			err = statedb.StorePdexv3PoolPairProtocolFee(
				env.StateDB(), poolPairID,
				statedb.NewPdexv3PoolPairProtocolFeeStateWithValue(tokenID, value),
			)
			if err != nil {
				return err
			}
		}
	}
	for tokenID, value := range p.stakingPoolFees {
		if poolPairChange.StakingPoolFees[tokenID.String()] {
			err = statedb.StorePdexv3PoolPairStakingPoolFee(
				env.StateDB(), poolPairID,
				statedb.NewPdexv3PoolPairStakingPoolFeeStateWithValue(tokenID, value),
			)
			if err != nil {
				return err
			}
		}
	}

	for tokenID, makingVolume := range poolPairChange.MakingVolume {
		tokenHash, _ := common.Hash{}.NewHashFromStr(tokenID)
		for nftID, isChanged := range makingVolume.Volume {
			if isChanged {
				if volume, found := p.makingVolume[*tokenHash].volume[nftID]; found {
					err = statedb.StorePdexv3PoolPairMakingVolume(
						env.StateDB(), poolPairID,
						statedb.NewPdexv3PoolPairMakingVolumeStateWithValue(
							nftID, *tokenHash, volume,
						),
					)
					if err != nil {
						return err
					}
				} else {
					err = statedb.DeletePdexv3PoolPairMakingVolume(env.StateDB(), poolPairID, nftID, *tokenHash)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	for nftID, orderReward := range poolPairChange.OrderRewards {
		for tokenID, isChanged := range orderReward.UncollectedReward {
			if isChanged {
				tokenHash, _ := common.Hash{}.NewHashFromStr(tokenID)
				if reward, found := p.orderRewards[nftID].uncollectedRewards[*tokenHash]; found {
					err = statedb.StorePdexv3PoolPairOrderReward(
						env.StateDB(), poolPairID,
						statedb.NewPdexv3PoolPairOrderRewardStateWithValue(
							*tokenHash, nftID, reward,
						),
					)
					if err != nil {
						return err
					}
				} else {
					nftHash, _ := common.Hash{}.NewHashFromStr(nftID)
					err = statedb.DeletePdexv3PoolPairOrderReward(env.StateDB(), poolPairID, tokenID, *nftHash)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	// store / delete orders
	ordersByID := make(map[string]*Order)
	for _, ord := range p.orderbook.orders {
		ordersByID[ord.Id()] = ord
	}
	for orderID, changed := range poolPairChange.OrderIDs {
		if changed {
			if order, exists := ordersByID[orderID]; exists {
				// update order in db
				orderState := statedb.NewPdexv3OrderStateWithValue(poolPairID, *order)
				err = statedb.StorePdexv3Order(env.StateDB(), *orderState)
				if err != nil {
					return err
				}
			} else {
				// delete order from db
				err = statedb.DeletePdexv3Order(env.StateDB(), poolPairID, orderID)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
