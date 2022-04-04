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
	tokenID         common.Hash
	externalTokenID []byte
}

func (v *Vault) TokenID() common.Hash {
	return v.tokenID
}

func (v *Vault) ExternalTokenID() []byte {
	return v.externalTokenID
}

func NewVault() *Vault {
	return &Vault{}
}

func NewVaultWithValue(state statedb.BridgeAggVaultState, tokenID common.Hash, externalTokenID []byte) *Vault {
	return &Vault{
		BridgeAggVaultState: state,
		tokenID:             tokenID,
		externalTokenID:     externalTokenID,
	}
}

func (v *Vault) Clone() *Vault {
	res := &Vault{
		BridgeAggVaultState: *v.BridgeAggVaultState.Clone(),
	}
	copy(res.tokenID[:], v.tokenID[:])
	copy(res.externalTokenID[:], v.externalTokenID[:])
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
		State           *statedb.BridgeAggVaultState `json:"State"`
		TokenID         common.Hash                  `json:"TokenID"`
		ExternalTokenID []byte                       `json:"ExternalTokenID,omitempty"`
	}{
		State:           &v.BridgeAggVaultState,
		TokenID:         v.tokenID,
		ExternalTokenID: v.externalTokenID,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (v *Vault) UnmarshalJSON(data []byte) error {
	temp := struct {
		State           *statedb.BridgeAggVaultState `json:"State"`
		TokenID         common.Hash                  `json:"TokenID"`
		ExternalTokenID []byte                       `json:"ExternalTokenID,omitempty"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	if temp.State != nil {
		v.BridgeAggVaultState = *temp.State
	}
	v.tokenID = temp.TokenID
	v.externalTokenID = temp.ExternalTokenID
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

func (v *Vault) convert(amount uint64, prefix string) error {
	tmpAmount, err := CalculateAmountByDecimal(
		*big.NewInt(0).SetUint64(amount), config.Param().BridgeAggParam.BaseDecimal, AddOperator, prefix, 0, v.externalTokenID,
	)
	if err != nil {
		return err
	}
	return v.increaseReserve(tmpAmount.Uint64())
}

func (v *Vault) shield(amount uint64) (uint64, error) {
	actualAmount, err := CalculateActualAmount(v.Reserve(), v.CurrentRewardReserve(), amount, AddOperator)
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
	actualAmount, err := EstimateActualAmountByBurntAmount(v.Reserve(), v.CurrentRewardReserve(), amount)
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
