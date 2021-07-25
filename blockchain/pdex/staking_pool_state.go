package pdex

type StakingPoolState struct {
	liquidity        uint64
	stakers          map[string]uint64 // nfst -> amount staking
	currentStakingID uint64
}

func NewStakingPoolState() *StakingPoolState {
	return &StakingPoolState{
		stakers: make(map[string]uint64),
	}
}

func NewStakingPoolStateWithValue(
	liquidity uint64,
	stakers map[string]uint64,
	currentStakingID uint64,
) *StakingPoolState {
	return &StakingPoolState{
		liquidity:        liquidity,
		stakers:          stakers,
		currentStakingID: currentStakingID,
	}
}

func (s *StakingPoolState) Clone() StakingPoolState {
	res := NewStakingPoolState()
	return *res
}
