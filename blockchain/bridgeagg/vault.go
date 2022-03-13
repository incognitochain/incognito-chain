package bridgeagg

import (
	"encoding/json"
	"errors"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type Vault struct {
	statedb.BridgeAggVaultState
	externalTokenID []byte
}

func NewVault() *Vault {
	return &Vault{}
}

func NewVaultWithValue(state statedb.BridgeAggVaultState, externalTokenID []byte) *Vault {
	return &Vault{
		BridgeAggVaultState: state,
		externalTokenID:     externalTokenID,
	}
}

func (v *Vault) Clone() *Vault {
	res := &Vault{
		BridgeAggVaultState: *v.BridgeAggVaultState.Clone(),
	}
	copy(v.externalTokenID, res.externalTokenID)
	return res
}

func (v *Vault) GetDiff(compareVault *Vault) (*Vault, *VaultChange, error) {
	vaultChange := NewVaultChange()
	if compareVault == nil {
		return nil, nil, errors.New("Compare vault is nul")
	}
	res := v.Clone()
	difVaultState, err := v.BridgeAggVaultState.GetDiff(&compareVault.BridgeAggVaultState)
	if err != nil {
		return nil, nil, err
	}
	if difVaultState != nil {
		vaultChange.IsReserveChanged = true
		return res, vaultChange, nil
	}
	return nil, nil, nil
}

func (v *Vault) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		statedb.BridgeAggVaultState
		ExternalTokenID []byte `json:"ExternalTokenID"`
	}{
		BridgeAggVaultState: v.BridgeAggVaultState,
		ExternalTokenID:     v.externalTokenID,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (v *Vault) UnmarshalJSON(data []byte) error {
	temp := struct {
		statedb.BridgeAggVaultState
		ExternalTokenID []byte `json:"ExternalTokenID"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	v.externalTokenID = temp.ExternalTokenID
	v.BridgeAggVaultState = temp.BridgeAggVaultState
	return nil
}

func (v *Vault) Convert(amount uint64) {
	v.SetReserve(v.Reserve() + amount)
}
