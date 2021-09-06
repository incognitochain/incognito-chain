package pdex

import (
	"encoding/json"
	"math/big"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type Share struct {
	amount             uint64
	tradingFees        map[common.Hash]uint64
	lastLPFeesPerShare map[common.Hash]*big.Int
}

func (share *Share) Amount() uint64 {
	return share.amount
}

func (share *Share) LastLPFeesPerShare() map[common.Hash]*big.Int {
	return share.lastLPFeesPerShare
}

func (share *Share) TradingFees() map[common.Hash]uint64 {
	res := make(map[common.Hash]uint64)
	for k, v := range share.tradingFees {
		res[k] = v
	}
	return res
}

func NewShare() *Share {
	return &Share{
		amount:             0,
		tradingFees:        map[common.Hash]uint64{},
		lastLPFeesPerShare: map[common.Hash]*big.Int{},
	}
}

func NewShareWithValue(
	amount uint64,
	tradingFees map[common.Hash]uint64,
	lastLPFeesPerShare map[common.Hash]*big.Int,
) *Share {
	return &Share{
		amount:             amount,
		tradingFees:        tradingFees,
		lastLPFeesPerShare: lastLPFeesPerShare,
	}
}

func (share *Share) Clone() *Share {
	res := NewShare()
	res.amount = share.amount
	res.tradingFees = map[common.Hash]uint64{}
	for k, v := range share.tradingFees {
		res.tradingFees[k] = v
	}
	res.lastLPFeesPerShare = map[common.Hash]*big.Int{}
	for k, v := range share.lastLPFeesPerShare {
		res.lastLPFeesPerShare[k] = new(big.Int).Set(v)
	}
	return res
}

func (share *Share) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Amount             uint64                   `json:"Amount"`
		TradingFees        map[common.Hash]uint64   `json:"TradingFees"`
		LastLPFeesPerShare map[common.Hash]*big.Int `json:"LastLPFeesPerShare"`
	}{
		Amount:             share.amount,
		TradingFees:        share.tradingFees,
		LastLPFeesPerShare: share.lastLPFeesPerShare,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (share *Share) UnmarshalJSON(data []byte) error {
	temp := struct {
		Amount             uint64                   `json:"Amount"`
		TradingFees        map[common.Hash]uint64   `json:"TradingFees"`
		LastLPFeesPerShare map[common.Hash]*big.Int `json:"LastLPFeesPerShare"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	share.amount = temp.Amount
	share.tradingFees = temp.TradingFees
	share.lastLPFeesPerShare = temp.LastLPFeesPerShare
	return nil
}

func (share *Share) getDiff(
	nftID string,
	compareShare *Share,
	stateChange *StateChange,
) *StateChange {
	newStateChange := stateChange
	if compareShare == nil || !reflect.DeepEqual(share, compareShare) {
		newStateChange.shares[nftID] = true
	}
	return newStateChange
}

type StakingChange struct {
	isChanged bool
	tokenIDs  map[string]bool
}

type StateChange struct {
	poolPairIDs map[string]bool
	shares      map[string]bool
	orders      map[string]map[int]bool
	orderIDs    map[string]bool
	stakingPool map[string]map[string]*StakingChange
}

func NewStateChange() *StateChange {
	return &StateChange{
		poolPairIDs: make(map[string]bool),
		shares:      make(map[string]bool),
		orders:      make(map[string]map[int]bool),
		orderIDs:    make(map[string]bool),
		stakingPool: make(map[string]map[string]*StakingChange),
	}
}

type Staker struct {
	liquidity               uint64
	lastUpdatedBeaconHeight uint64
	rewards                 map[string]uint64
}

func (staker *Staker) Liquidity() uint64 {
	return staker.liquidity
}

func (staker *Staker) LastUpdatedBeaconHeight() uint64 {
	return staker.lastUpdatedBeaconHeight
}

func (staker *Staker) Rewards() map[string]uint64 {
	res := make(map[string]uint64)
	for k, v := range staker.rewards {
		res[k] = v
	}
	return res
}

func (staker *Staker) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Liquidity               uint64            `json:"Liquidity"`
		LastUpdatedBeaconHeight uint64            `json:"LastUpdatedBeaconHeight"`
		Rewards                 map[string]uint64 `json:"Rewards"`
	}{
		Liquidity:               staker.liquidity,
		LastUpdatedBeaconHeight: staker.lastUpdatedBeaconHeight,
		Rewards:                 staker.rewards,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (staker *Staker) UnmarshalJSON(data []byte) error {
	temp := struct {
		Liquidity               uint64            `json:"Liquidity"`
		LastUpdatedBeaconHeight uint64            `json:"LastUpdatedBeaconHeight"`
		Rewards                 map[string]uint64 `json:"Rewards"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	staker.liquidity = temp.Liquidity
	staker.lastUpdatedBeaconHeight = temp.LastUpdatedBeaconHeight
	staker.rewards = temp.Rewards
	return nil
}

func NewStaker() *Staker {
	return &Staker{
		rewards: make(map[string]uint64),
	}
}

func NewStakerWithValue(liquidity, lastUpdatedBeaconHeight uint64, rewards map[string]uint64) *Staker {
	return &Staker{
		liquidity:               liquidity,
		lastUpdatedBeaconHeight: lastUpdatedBeaconHeight,
		rewards:                 rewards,
	}
}

func (staker *Staker) Clone() *Staker {
	res := NewStaker()
	res.liquidity = staker.liquidity
	res.lastUpdatedBeaconHeight = staker.lastUpdatedBeaconHeight
	for k, v := range staker.rewards {
		res.rewards[k] = v
	}
	return res
}

func (staker *Staker) getDiff(stakingPoolID, nftID string, compareStaker *Staker, stateChange *StateChange) *StateChange {
	newStateChange := stateChange
	stakingChange := &StakingChange{}
	if compareStaker == nil {
		stakingChange = &StakingChange{
			isChanged: true,
			tokenIDs:  make(map[string]bool),
		}
		newStateChange.stakingPool[stakingPoolID][nftID] = stakingChange
		for tokenID := range staker.rewards {
			newStateChange.stakingPool[stakingPoolID][nftID].tokenIDs[tokenID] = true
		}
	} else {
		if staker.liquidity != compareStaker.liquidity {
			stakingChange.isChanged = true
		}
		newStateChange.stakingPool[stakingPoolID][nftID] = stakingChange
		for tokenID, value := range staker.rewards {
			if v, ok := compareStaker.rewards[nftID]; !ok || !reflect.DeepEqual(v, value) {
				if stakingChange.tokenIDs == nil {
					stakingChange.tokenIDs = make(map[string]bool)
				}
				newStateChange.stakingPool[stakingPoolID][nftID].tokenIDs[tokenID] = true
			}
		}
	}
	return newStateChange
}

func addStakingPoolState(
	stakingPoolStates map[string]*StakingPoolState, stakingPoolIDs map[string]uint,
) map[string]*StakingPoolState {
	for k := range stakingPoolIDs {
		if stakingPoolStates[k] == nil {
			stakingPoolStates[k] = NewStakingPoolState()
		}
	}
	return stakingPoolStates
}
