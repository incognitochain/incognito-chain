package pdex

import (
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

func (s *StakingPoolState) Clone() *StakingPoolState {
	res := NewStakingPoolState()
	res.liquidity = s.liquidity
	for k, v := range s.stakers {
		res.stakers[k] = v
	}
	return res
}

func (s *StakingPoolState) getDiff(
	stakingPoolID string, compareStakingPoolState *StakingPoolState, stateChange *StateChange,
) *StateChange {
	newStateChange := stateChange
	if compareStakingPoolState == nil {
		for nftID, staker := range s.stakers {
			newStateChange = staker.getDiff(stakingPoolID, nftID, nil, newStateChange)
		}
	} else {
		for nftID, staker := range s.stakers {
			if m, ok := compareStakingPoolState.stakers[nftID]; !ok || !reflect.DeepEqual(m, staker) {
				newStateChange = staker.getDiff(stakingPoolID, nftID, m, newStateChange)
			}
		}
	}
	return newStateChange
}

func (s *StakingPoolState) addLiquidity(nftID string, liquidity, beaconHeight uint64) error {
	staker, found := s.stakers[nftID]
	if !found {
		s.stakers[nftID] = NewStakerWithValue(liquidity, beaconHeight, make(map[string]uint64))
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
