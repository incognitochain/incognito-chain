package pdex

import (
	"encoding/json"
	"reflect"
)

type Share struct {
	amount                  uint64
	tradingFees             map[string]uint64
	lastUpdatedBeaconHeight uint64
}

func (share *Share) Amount() uint64 {
	return share.amount
}

func (share *Share) LastUpdatedBeaconHeight() uint64 {
	return share.lastUpdatedBeaconHeight
}

func (share *Share) TradingFees() map[string]uint64 {
	res := make(map[string]uint64)
	for k, v := range share.tradingFees {
		res[k] = v
	}
	return res
}

func NewShare() *Share {
	return &Share{
		tradingFees: make(map[string]uint64),
	}
}

func NewShareWithValue(
	amount uint64,
	tradingFees map[string]uint64,
	lastUpdatedBeaconHeight uint64,
) *Share {
	return &Share{
		amount:                  amount,
		tradingFees:             tradingFees,
		lastUpdatedBeaconHeight: lastUpdatedBeaconHeight,
	}
}

func (share *Share) Clone() *Share {
	res := NewShare()
	res.amount = share.amount
	res.lastUpdatedBeaconHeight = share.lastUpdatedBeaconHeight
	res.tradingFees = make(map[string]uint64)
	for k, v := range share.tradingFees {
		res.tradingFees[k] = v
	}
	return res
}

func (share *Share) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Amount                  uint64            `json:"Amount"`
		TradingFees             map[string]uint64 `json:"TradingFees"`
		LastUpdatedBeaconHeight uint64            `json:"LastUpdatedBeaconHeight"`
	}{
		Amount:                  share.amount,
		TradingFees:             share.tradingFees,
		LastUpdatedBeaconHeight: share.lastUpdatedBeaconHeight,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (share *Share) UnmarshalJSON(data []byte) error {
	temp := struct {
		Amount                  uint64            `json:"Amount"`
		TradingFees             map[string]uint64 `json:"TradingFees"`
		LastUpdatedBeaconHeight uint64            `json:"LastUpdatedBeaconHeight"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	share.amount = temp.Amount
	share.lastUpdatedBeaconHeight = temp.LastUpdatedBeaconHeight
	share.tradingFees = temp.TradingFees
	return nil
}

func (share *Share) getDiff(
	nftID string,
	compareShare *Share,
	stateChange *StateChange,
) *StateChange {
	newStateChange := stateChange
	if compareShare == nil {
		newStateChange.shares[nftID] = &ShareChange{
			isChanged: true,
		}
		for tokenID := range share.tradingFees {
			if newStateChange.shares[nftID].tokenIDs == nil {
				newStateChange.shares[nftID].tokenIDs = make(map[string]bool)
			}
			newStateChange.shares[nftID].tokenIDs[tokenID] = true
		}
	} else {
		if share.amount != compareShare.amount || share.lastUpdatedBeaconHeight != compareShare.lastUpdatedBeaconHeight {
			newStateChange.shares[nftID] = &ShareChange{
				isChanged: true,
			}
		}
		for k, v := range share.tradingFees {
			if m, ok := compareShare.tradingFees[k]; !ok || !reflect.DeepEqual(m, v) {
				if newStateChange.shares[nftID].tokenIDs == nil {
					newStateChange.shares[nftID].tokenIDs = make(map[string]bool)
				}
				newStateChange.shares[nftID].tokenIDs[k] = true
			}
		}
	}
	return newStateChange
}

type ShareChange struct {
	isChanged bool
	tokenIDs  map[string]bool
}

type StakingChange struct {
	isChanged bool
	tokenIDs  map[string]bool
}

type StateChange struct {
	poolPairIDs map[string]bool
	shares      map[string]*ShareChange
	orderIDs    map[string]bool
	stakingPool map[string]map[string]*StakingChange
}

func NewStateChange() *StateChange {
	return &StateChange{
		poolPairIDs: make(map[string]bool),
		shares:      make(map[string]*ShareChange),
		orderIDs:    make(map[string]bool),
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
	var stakingChange *StakingChange
	if compareStaker == nil {
		stakingChange = &StakingChange{
			isChanged: true,
			tokenIDs:  make(map[string]bool),
		}
		for tokenID := range staker.rewards {
			newStateChange.stakingPool[stakingPoolID][nftID].tokenIDs[tokenID] = true
		}
	} else {
		if staker.liquidity != compareStaker.liquidity {
			stakingChange = &StakingChange{
				isChanged: true,
			}
		}
		for tokenID, value := range staker.rewards {
			if v, ok := compareStaker.rewards[nftID]; !ok || !reflect.DeepEqual(v, value) {
				if stakingChange.tokenIDs == nil {
					stakingChange.tokenIDs = make(map[string]bool)
				}
				newStateChange.stakingPool[stakingPoolID][nftID].tokenIDs[tokenID] = true
			}
		}
	}
	if stakingChange != nil {
		newStateChange.stakingPool[stakingPoolID][nftID] = stakingChange
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
