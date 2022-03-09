package statedb

type BridgeAggVaultState struct {
	reserve                  uint64
	lastUpdatedRewardReserve uint64
	currentRewardReserve     uint64
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
