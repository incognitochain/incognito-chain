package bridgeagg

import (
	"encoding/base64"
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

func InitStateFromDB(sDB *statedb.StateDB) (*State, error) {
	unifiedTokenStates, err := statedb.GetBridgeAggUnifiedTokens(sDB)
	if err != nil {
		return nil, err
	}
	unifiedTokenInfos := make(map[common.Hash]map[common.Hash]*Vault)
	for _, unifiedTokenState := range unifiedTokenStates {
		unifiedTokenInfos[unifiedTokenState.TokenID()] = make(map[common.Hash]*Vault)
		vaults, err := statedb.GetBridgeAggVaults(sDB, unifiedTokenState.TokenID())
		if err != nil {
			return nil, err
		}
		for tokenID, vault := range vaults {
			v := NewVaultWithValue(*vault)
			unifiedTokenInfos[unifiedTokenState.TokenID()][tokenID] = v
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

func GetBridgeTokenIndex(sDB *statedb.StateDB) (map[common.Hash]*rawdbv2.BridgeTokenInfo, map[string]bool, error) {
	bridgeTokenInfoIndex := make(map[common.Hash]*rawdbv2.BridgeTokenInfo)
	externalTokenIDIndex := make(map[string]bool)
	bridgeTokenInfoStates, err := statedb.GetBridgeTokens(sDB)
	if err != nil {
		return nil, nil, err
	}
	for _, bridgeTokenInfoState := range bridgeTokenInfoStates {
		bridgeTokenInfoIndex[*bridgeTokenInfoState.TokenID] = bridgeTokenInfoState
		encodedExternalTokenID := base64.StdEncoding.EncodeToString(bridgeTokenInfoState.ExternalTokenID)
		externalTokenIDIndex[encodedExternalTokenID] = true

	}
	return bridgeTokenInfoIndex, externalTokenIDIndex, nil
}
