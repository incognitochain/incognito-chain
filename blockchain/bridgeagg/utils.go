package bridgeagg

import (
	"errors"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
)

type ShieldStatus struct {
	Status    byte `json:"Status"`
	ErrorCode uint `json:"ErrorCode,omitempty"`
}

type ModifyListTokenStatus struct {
	Status    byte `json:"Status"`
	ErrorCode uint `json:"ErrorCode,omitempty"`
}

type ConvertStatus struct {
	Status    byte `json:"Status"`
	ErrorCode uint `json:"ErrorCode,omitempty"`
}

type VaultChange struct {
	IsChanged        bool
	IsReserveChanged bool
}

func NewVaultChange() *VaultChange {
	return &VaultChange{}
}

type StateChange struct {
	unifiedTokenID map[common.Hash]bool
	vaultChange    map[common.Hash]map[uint]VaultChange
}

func NewStateChange() *StateChange {
	return &StateChange{
		unifiedTokenID: make(map[common.Hash]bool),
		vaultChange:    make(map[common.Hash]map[uint]VaultChange),
	}
}

func CalculateRewardByAmount(x, y, deltaX uint64, operator byte) (uint64, error) {
	k := big.NewInt(0).Mul(big.NewInt(0).SetUint64(x), big.NewInt(0).SetUint64(y))
	newX := big.NewInt(0) // x'
	actualAmount := big.NewInt(0)
	switch operator {
	case AddOperator:
		newX.Add(big.NewInt(0).SetUint64(x), big.NewInt(0).SetUint64(deltaX))
		newY := big.NewInt(0).Div(k, newX) // y'
		reward := big.NewInt(0).Sub(big.NewInt(0).SetUint64(y), newY)
		actualAmount = big.NewInt(0).Add(big.NewInt(0).SetUint64(deltaX), reward)
		if actualAmount.Cmp(big.NewInt(0).SetUint64(deltaX)) < 0 {
			return 0, errors.New("actualAmount < deltaX")
		}
	case SubOperator:
		newX.Sub(big.NewInt(0).SetUint64(x), big.NewInt(0).SetUint64(deltaX))
		newY := big.NewInt(0).Div(k, newX) // y'
		fee := big.NewInt(0).Sub(newY, big.NewInt(0).SetUint64(y))
		actualAmount = big.NewInt(0).Sub(big.NewInt(0).SetUint64(deltaX), fee)
		if actualAmount.Cmp(big.NewInt(0).SetUint64(deltaX)) > 0 {
			return 0, errors.New("actualAmount > deltaX")
		}
	default:
		return 0, errors.New("Cannot recognize operator")
	}
	if !actualAmount.IsUint64() {
		return 0, errors.New("Actual amount is not uint64")
	}
	return actualAmount.Uint64(), nil
}

func EstimateActualAmountByBurntAmount(x, y, burntAmount uint64) (uint64, error) {
	return 0, nil
}
