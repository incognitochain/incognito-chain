package bridgeagg

import (
	"encoding/json"
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type Vault struct {
	statedb.BridgeAggVaultState
	externalTokenID []byte
	tokenID         common.Hash
}

func NewVault() *Vault {
	return &Vault{}
}

func NewVaultWithValue(state statedb.BridgeAggVaultState, externalTokenID []byte, tokenID common.Hash) *Vault {
	return &Vault{
		BridgeAggVaultState: state,
		externalTokenID:     externalTokenID,
		tokenID:             tokenID,
	}
}

func (v *Vault) Clone() *Vault {
	res := &Vault{
		BridgeAggVaultState: *v.BridgeAggVaultState.Clone(),
	}
	copy(res.externalTokenID, v.externalTokenID)
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
		ReserveState    *statedb.BridgeAggVaultState `json:"ReserveState"`
		ExternalTokenID []byte                       `json:"ExternalTokenID"`
		TokenID         common.Hash                  `json:"TokenID"`
	}{
		ReserveState:    &v.BridgeAggVaultState,
		ExternalTokenID: v.externalTokenID,
		TokenID:         v.tokenID,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (v *Vault) UnmarshalJSON(data []byte) error {
	temp := struct {
		ReserveState    *statedb.BridgeAggVaultState `json:"ReserveState"`
		ExternalTokenID []byte                       `json:"ExternalTokenID"`
		TokenID         common.Hash                  `json:"TokenID"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	v.externalTokenID = temp.ExternalTokenID
	if temp.ReserveState != nil {
		v.BridgeAggVaultState = *temp.ReserveState
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

func (v *Vault) convert(amount uint64) error {
	return v.increaseReserve(amount)
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

func (v *Vault) unshield(amount uint64) error {
	return nil
}
