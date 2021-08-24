package pdex

import (
	"encoding/json"
	"errors"
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

type StateChange struct {
	poolPairIDs map[string]bool
	shares      map[string]*ShareChange
	orderIDs    map[string]bool
}

func NewStateChange() *StateChange {
	return &StateChange{
		poolPairIDs: make(map[string]bool),
		shares:      make(map[string]*ShareChange),
		orderIDs:    make(map[string]bool),
	}
}

type StakingPoolState struct {
	liquidity uint64
	stakers   map[string]*Staker // nft -> amount staking
}

func NewStakingPoolState() *StakingPoolState {
	return &StakingPoolState{
		stakers: make(map[string]*Staker),
	}
}

func NewStakingPoolStateWithValue(
	liquidity uint64,
	stakers map[string]*Staker,
) *StakingPoolState {
	return &StakingPoolState{
		liquidity: liquidity,
		stakers:   stakers,
	}
}

func (s *StakingPoolState) Clone() *StakingPoolState {
	res := NewStakingPoolState()
	res.liquidity = s.liquidity
	for k, v := range s.stakers {
		res.stakers[k] = v
	}
	return res
}

func (s *StakingPoolState) addLiquidity(nftID string, liquidity, beaconHeight uint64) error {
	staker, found := s.stakers[nftID]
	if !found {
		s.stakers[nftID] = NewStakerWithValue(liquidity, 0, beaconHeight, make(map[string]uint64))
	} else {
		tempLiquidity := staker.liquidity + liquidity
		if tempLiquidity < s.liquidity {
			return errors.New("Staking pool liquidity is out of range")
		}
		staker.liquidity = tempLiquidity
	}
	tempLiquidity := s.liquidity + liquidity
	if tempLiquidity < s.liquidity {
		return errors.New("Staking pool liquidity is out of range")
	}
	return nil
}

type Staker struct {
	liquidity               uint64
	uncollectedReward       uint64
	lastUpdatedBeaconHeight uint64
	tokenIDs                map[string]uint64
}

func NewStaker() *Staker {
	return &Staker{
		tokenIDs: make(map[string]uint64),
	}
}

func NewStakerWithValue(
	liquidity, uncollectedReward, lastUpdatedBeaconHeight uint64,
	tokenIDs map[string]uint64,
) *Staker {
	return &Staker{
		liquidity:               liquidity,
		uncollectedReward:       uncollectedReward,
		lastUpdatedBeaconHeight: lastUpdatedBeaconHeight,
		tokenIDs:                tokenIDs,
	}
}

func (staker *Staker) Clone() *Staker {
	res := NewStaker()
	res.liquidity = staker.liquidity
	res.lastUpdatedBeaconHeight = staker.lastUpdatedBeaconHeight
	res.uncollectedReward = staker.uncollectedReward
	for k, v := range staker.tokenIDs {
		res.tokenIDs[k] = v
	}
	return res
}
