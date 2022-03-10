package bridgeagg

import (
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

func (v *Vault) GetDiff(compareVault *Vault) (*Vault, error) {
	if compareVault == nil {
		return nil, errors.New("Compare vault is nul")
	}
	res := v.Clone()
	difVaultState, err := v.BridgeAggVaultState.GetDiff(&compareVault.BridgeAggVaultState)
	if err != nil {
		return nil, err
	}
	if difVaultState != nil {
		return res, nil
	}
	return nil, nil
}
