package v2utils

type StateChange struct {
	PoolPairs   map[string]*PoolPairChange
	OrderIDs    map[string]bool
	StakingPool map[string]map[string]*StakingPoolChange
}

func NewStateChange() *StateChange {
	return &StateChange{
		PoolPairs:   make(map[string]*PoolPairChange),
		OrderIDs:    make(map[string]bool),
		StakingPool: make(map[string]map[string]*StakingPoolChange),
	}
}

type StakingPoolChange struct {
	IsChanged bool
	TokenIDs  map[string]bool
}

func NewStakingChange() *StakingPoolChange {
	return &StakingPoolChange{
		TokenIDs: make(map[string]bool),
	}
}

type PoolPairChange struct {
	IsChanged       bool
	Shares          map[string]*ShareChange
	LpFeesPerShare  map[string]bool
	ProtocolFees    map[string]bool
	StakingPoolFees map[string]bool
}

func NewPoolPairChange() *PoolPairChange {
	return &PoolPairChange{
		Shares:          make(map[string]*ShareChange),
		LpFeesPerShare:  make(map[string]bool),
		ProtocolFees:    make(map[string]bool),
		StakingPoolFees: make(map[string]bool),
	}
}

type ShareChange struct {
	IsChanged          bool
	TradingFees        map[string]bool
	LastLPFeesPerShare map[string]bool
}

func NewShareChange() *ShareChange {
	return &ShareChange{
		TradingFees:        make(map[string]bool),
		LastLPFeesPerShare: make(map[string]bool),
	}
}
