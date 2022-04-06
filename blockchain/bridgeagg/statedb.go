package bridgeagg

import (
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

func InitStateFromDB(sDB *statedb.StateDB) (*State, error) {
	unifiedTokenStates, err := statedb.GetBridgeAggUnifiedTokens(sDB)
	if err != nil {
		return nil, err
	}
	unifiedTokenInfos := make(map[common.Hash]map[uint]*Vault)
	for _, unifiedTokenState := range unifiedTokenStates {
		unifiedTokenInfos[unifiedTokenState.TokenID()] = make(map[uint]*Vault)
		convertTokens, err := statedb.GetBridgeAggConvertedTokens(sDB, unifiedTokenState.TokenID())
		if err != nil {
			return nil, err
		}
		for _, convertToken := range convertTokens {
			state, err := statedb.GetBridgeAggVault(sDB, unifiedTokenState.TokenID(), convertToken.TokenID())
			if err != nil {
				state = statedb.NewBridgeAggVaultState()
			}
			vault := NewVaultWithValue(*state, convertToken.TokenID())
			unifiedTokenInfos[unifiedTokenState.TokenID()][convertToken.NetworkID()] = vault
		}
	}
	return NewStateWithValue(unifiedTokenInfos), nil
}

func GetExternalTokenIDByIncTokenID(incTokenID common.Hash, sDB *statedb.StateDB) ([]byte, error) {
	info, has, err := statedb.GetBridgeTokenByType(sDB, incTokenID, false)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not found externalTokenID")
	}
	return info.ExternalTokenID(), nil
}

func GetVaultByUnifiedTokenIDAndNetworkID(unifiedTokenID common.Hash, networkID uint, sDB *statedb.StateDB) (*Vault, error) {
	convertTokens, err := statedb.GetBridgeAggConvertedTokens(sDB, unifiedTokenID)
	if err != nil {
		return nil, err
	}
	for _, convertToken := range convertTokens {
		if convertToken.NetworkID() != networkID {
			continue
		}
		state, err := statedb.GetBridgeAggVault(sDB, unifiedTokenID, convertToken.TokenID())
		if err != nil {
			return nil, err
		}
		state = statedb.NewBridgeAggVaultState()
		return NewVaultWithValue(*state, convertToken.TokenID()), nil
	}
	return nil, errors.New("Not found vault")
}
