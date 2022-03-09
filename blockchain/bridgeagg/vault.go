package bridgeagg

import "github.com/incognitochain/incognito-chain/dataaccessobject/statedb"

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
