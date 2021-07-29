package pdex

import "reflect"

type Share struct {
	amount                  uint64
	tradingFees             map[string]uint64
	lastUpdatedBeaconHeight uint64
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

func (share *Share) getDiff(nfctID string, compareShare *Share, stateChange *StateChange) *StateChange {
	newStateChange := stateChange
	for k, v := range share.tradingFees {
		if m, ok := compareShare.tradingFees[k]; !ok || !reflect.DeepEqual(m, v) {
			newStateChange.tokenIDs[k] = true
		}
	}
	if share.amount != compareShare.amount || share.lastUpdatedBeaconHeight != compareShare.lastUpdatedBeaconHeight {
		newStateChange.nfctIDs[nfctID] = true
	}
	return newStateChange
}

type StakingInfo struct {
	amount                  uint64
	uncollectedReward       uint64
	lastUpdatedBeaconHeight uint64
}

type StateChange struct {
	poolPairIDs map[string]bool
	nfctIDs     map[string]bool
	tokenIDs    map[string]bool
}

type StakingPoolState struct {
	liquidity        uint64
	stakers          map[string]StakingInfo // nfst -> amount staking
	currentStakingID uint64
}

func NewStakingPoolState() *StakingPoolState {
	return &StakingPoolState{
		stakers: make(map[string]StakingInfo),
	}
}

func NewStakingPoolStateWithValue(
	liquidity uint64,
	stakers map[string]StakingInfo,
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
