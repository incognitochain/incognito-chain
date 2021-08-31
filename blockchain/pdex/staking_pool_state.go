package pdex

import (
	"encoding/json"
	"errors"
	"reflect"
)

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

func (stakingPoolState *StakingPoolState) Liquidity() uint64 {
	return stakingPoolState.liquidity
}

func (stakingPoolState *StakingPoolState) Stakers() map[string]*Staker {
	res := make(map[string]*Staker)
	for k, v := range stakingPoolState.stakers {
		res[k] = v.Clone()
	}
	return res
}

func (stakingPoolState *StakingPoolState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Liquidity uint64             `json:"Liquidity"`
		Stakers   map[string]*Staker `json:"Stakers"`
	}{
		Liquidity: stakingPoolState.liquidity,
		Stakers:   stakingPoolState.stakers,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (stakingPoolState *StakingPoolState) UnmarshalJSON(data []byte) error {
	temp := struct {
		Liquidity uint64             `json:"Liquidity"`
		Stakers   map[string]*Staker `json:"Stakers"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	stakingPoolState.liquidity = temp.Liquidity
	stakingPoolState.stakers = temp.Stakers
	return nil
}

func (s *StakingPoolState) Clone() *StakingPoolState {
	res := NewStakingPoolState()
	res.liquidity = s.liquidity
	for k, v := range s.stakers {
		res.stakers[k] = v.Clone()
	}
	return res
}

func (s *StakingPoolState) getDiff(
	stakingPoolID string, compareStakingPoolState *StakingPoolState, stateChange *StateChange,
) *StateChange {
	newStateChange := stateChange
	if compareStakingPoolState == nil {
		if newStateChange.stakingPool[stakingPoolID] == nil {
			newStateChange.stakingPool[stakingPoolID] = make(map[string]*StakingChange)
		}
		for nftID, staker := range s.stakers {
			newStateChange = staker.getDiff(stakingPoolID, nftID, nil, newStateChange)
		}
	} else {
		for nftID, staker := range s.stakers {
			if m, ok := compareStakingPoolState.stakers[nftID]; !ok || !reflect.DeepEqual(m, staker) {
				if newStateChange.stakingPool[stakingPoolID] == nil {
					newStateChange.stakingPool[stakingPoolID] = make(map[string]*StakingChange)
				}
				newStateChange = staker.getDiff(stakingPoolID, nftID, m, newStateChange)
			}
		}
	}
	return newStateChange
}

func (s *StakingPoolState) updateLiquidity(nftID string, liquidity, beaconHeight uint64, operator byte) error {
	staker, found := s.stakers[nftID]
	if !found {
		if operator == subOperator {
			return errors.New("remove liquidity from invalid staker")
		}
		s.stakers[nftID] = NewStakerWithValue(liquidity, beaconHeight, make(map[string]uint64))
	} else {
		var tempLiquidity uint64
		switch operator {
		case subOperator:
			tempLiquidity = staker.liquidity - liquidity
			if tempLiquidity >= s.liquidity {
				return errors.New("Staker liquidity is out of range")
			}
		case addOperator:
			tempLiquidity = staker.liquidity + liquidity
			if tempLiquidity < s.liquidity {
				return errors.New("Staker liquidity is out of range")
			}
		}
		staker.liquidity = tempLiquidity
	}
	var tempLiquidity uint64
	switch operator {
	case subOperator:
		tempLiquidity = s.liquidity - liquidity
		if tempLiquidity >= s.liquidity {
			return errors.New("Staking pool liquidity is out of range")
		}
	case addOperator:
		tempLiquidity = s.liquidity + liquidity
		if tempLiquidity < s.liquidity {
			return errors.New("Staking pool liquidity is out of range")
		}
	}
	s.liquidity = tempLiquidity
	return nil
}

func (s *StakingPoolState) withLiquidity(liquidity uint64) {
	s.liquidity = liquidity
}

func (s *StakingPoolState) withStakers(stakers map[string]*Staker) {
	s.stakers = stakers
}

func (s *StakingPoolState) cloneStaker(nftID string) map[string]*Staker {
	res := make(map[string]*Staker)
	for k, v := range s.stakers {
		if k == nftID {
			res[k] = v.cloneState()
		} else {
			res[k] = v
		}
	}
	return res
}
