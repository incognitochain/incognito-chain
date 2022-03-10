package statedb

import "errors"

type BridgeAggVaultState struct {
	reserve                  uint64
	lastUpdatedRewardReserve uint64
	currentRewardReserve     uint64
}

func NewBridgeAggVaultState() *BridgeAggVaultState {
	return &BridgeAggVaultState{}
}

func (b *BridgeAggVaultState) Reserve() uint64 {
	return b.reserve
}

func (b *BridgeAggVaultState) LastUpdatedRewardReserve() uint64 {
	return b.lastUpdatedRewardReserve
}

func (b *BridgeAggVaultState) CurrentRewardReserve() uint64 {
	return b.currentRewardReserve
}

func (b *BridgeAggVaultState) SetReserve(reserve uint64) {
	b.reserve = reserve
}

func (b *BridgeAggVaultState) SetLastUpdatedRewardReserve(lastUpdatedRewardReserve uint64) {
	b.lastUpdatedRewardReserve = lastUpdatedRewardReserve
}

func (b *BridgeAggVaultState) SetCurrentRewardReserve(currentRewardReserve uint64) {
	b.currentRewardReserve = currentRewardReserve
}

func (b *BridgeAggVaultState) Clone() *BridgeAggVaultState {
	return &BridgeAggVaultState{
		reserve:                  b.reserve,
		lastUpdatedRewardReserve: b.lastUpdatedRewardReserve,
		currentRewardReserve:     b.currentRewardReserve,
	}
}

func (b *BridgeAggVaultState) GetDiff(compareState *BridgeAggVaultState) (*BridgeAggVaultState, error) {
	if compareState == nil {
		return nil, errors.New("compareState is nil")
	}
	res := b.Clone()
	if b.reserve != compareState.reserve ||
		b.currentRewardReserve != compareState.currentRewardReserve ||
		b.lastUpdatedRewardReserve != compareState.lastUpdatedRewardReserve {
		return res, nil
	}
	return nil, nil
}
