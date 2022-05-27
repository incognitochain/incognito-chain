package bridgeagg

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type Vault struct {
	statedb.BridgeAggVaultState
	tokenID common.Hash
}

func (v *Vault) TokenID() common.Hash {
	return v.tokenID
}

func NewVault() *Vault {
	return &Vault{}
}

func NewVaultWithValue(state statedb.BridgeAggVaultState, tokenID common.Hash) *Vault {
	return &Vault{
		BridgeAggVaultState: state,
		tokenID:             tokenID,
	}
}

func (v *Vault) Clone() *Vault {
	res := &Vault{
		BridgeAggVaultState: *v.BridgeAggVaultState.Clone(),
	}
	copy(res.tokenID[:], v.tokenID[:])
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
	if v.tokenID.String() != compareVault.tokenID.String() {
		vaultChange.IsChanged = true
	}
	if difVaultState != nil {
		vaultChange.IsReserveChanged = true
	}
	if vaultChange.IsChanged || vaultChange.IsReserveChanged {
		return v.Clone(), vaultChange, nil
	}
	return nil, nil, nil
}

func (v *Vault) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		State   *statedb.BridgeAggVaultState `json:"State"`
		TokenID common.Hash                  `json:"TokenID"`
	}{
		State:   &v.BridgeAggVaultState,
		TokenID: v.tokenID,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (v *Vault) UnmarshalJSON(data []byte) error {
	temp := struct {
		State   *statedb.BridgeAggVaultState `json:"State"`
		TokenID common.Hash                  `json:"TokenID"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	if temp.State != nil {
		v.BridgeAggVaultState = *temp.State
	}
	v.tokenID = temp.TokenID
	return nil
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
