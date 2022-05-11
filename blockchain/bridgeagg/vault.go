package bridgeagg

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type Vault struct {
	statedb.BridgeAggVaultState
}

func NewVault() *Vault {
	return &Vault{}
}

func NewVaultWithValue(state statedb.BridgeAggVaultState) *Vault {
	return &Vault{
		BridgeAggVaultState: state,
	}
}

func (v *Vault) Clone() *Vault {
	res := &Vault{
		BridgeAggVaultState: *v.BridgeAggVaultState.Clone(),
	}
	return res
}

func (v *Vault) GetDiff(compareVault *Vault) (*Vault, *VaultChange, error) {
	vaultChange := NewVaultChange()
	if compareVault == nil {
		return nil, nil, errors.New("Compare vault is nul")
	}
	difVaultState, err := v.BridgeAggVaultState.GetDiff(&compareVault.BridgeAggVaultState)
	if err != nil {
		return nil, nil, err
	}
	if difVaultState != nil {
		vaultChange.IsChanged = true
	}
	if vaultChange.IsChanged {
		return v.Clone(), vaultChange, nil
	}
	return nil, nil, nil
}

func (v *Vault) decreaseCurrentRewardReserve(amount uint64) error {
	temp := v.CurrentRewardReserve() - amount
	if temp > v.CurrentRewardReserve() {
		return errors.New("decrease out of range uint64")
	}
	v.SetCurrentRewardReserve(temp)
	return nil
}

func (v *Vault) decreaseReserve(amount uint64) error {
	temp := v.Reserve() - amount
	if temp > v.Reserve() {
		return errors.New("decrease out of range uint64")
	}
	v.SetReserve(temp)
	return nil
}

func (v *Vault) increaseCurrentRewardReserve(amount uint64) error {
	temp := v.CurrentRewardReserve() + amount
	if temp < v.CurrentRewardReserve() {
		return errors.New("increase out of range uint64")
	}
	v.SetCurrentRewardReserve(temp)
	return nil
}

func (v *Vault) increaseReserve(amount uint64) error {
	temp := v.Reserve() + amount
	if temp < v.Reserve() {
		return errors.New("increase out of range uint64")
	}
	v.SetReserve(temp)
	return nil
}

func (v *Vault) convert(amount uint64) (uint64, error) {
	decimal := CalculateIncDecimal(v.Decimal(), config.Param().BridgeAggParam.BaseDecimal)
	tmpAmount, err := CalculateAmountByDecimal(*big.NewInt(0).SetUint64(amount), decimal, AddOperator)
	if err != nil {
		return 0, err
	}
	if tmpAmount.Cmp(big.NewInt(0)) == 0 {
		return 0, fmt.Errorf("amount %d is not enough for converting", amount)
	}
	return tmpAmount.Uint64(), v.increaseReserve(tmpAmount.Uint64())
}

func (v *Vault) shield(amount uint64) (uint64, error) {
	actualAmount, err := CalculateShieldActualAmount(v.Reserve(), v.CurrentRewardReserve(), amount, v.IsPaused())
	if err != nil {
		return 0, err
	}
	err = v.decreaseCurrentRewardReserve(actualAmount - amount)
	if err != nil {
		return 0, err
	}
	err = v.increaseReserve(amount)
	if err != nil {
		return 0, err
	}
	return actualAmount, nil
}

func (v *Vault) unshield(amount, expectedAmount uint64) (uint64, error) {
	actualAmount, err := EstimateActualAmountByBurntAmount(v.Reserve(), v.CurrentRewardReserve(), amount, v.IsPaused())
	if err != nil {
		return 0, err
	}
	if actualAmount < expectedAmount {
		return 0, fmt.Errorf("actual amount %v < expected amount %v", actualAmount, expectedAmount)
	}
	err = v.increaseCurrentRewardReserve(amount - actualAmount)
	if err != nil {
		return 0, err
	}
	err = v.decreaseReserve(actualAmount)
	if err != nil {
		return 0, err
	}
	return actualAmount, nil
}

func (v *Vault) updateRewardReserve(newRewardReserve uint64, isPaused bool) error {
	newLastUpdatedRewardReserve, newCurrentRewardReserve, err := updateRewardReserve(v.LastUpdatedRewardReserve(), v.CurrentRewardReserve(), newRewardReserve)
	if err != nil {
		return err
	}
	v.SetLastUpdatedRewardReserve(newLastUpdatedRewardReserve)
	v.SetCurrentRewardReserve(newCurrentRewardReserve)
	v.SetIsPaused(isPaused)
	return nil
}
