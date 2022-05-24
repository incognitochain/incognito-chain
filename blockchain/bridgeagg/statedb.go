package bridgeagg

import (
	"encoding/base64"
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

func InitManager(sDB *statedb.StateDB) (*Manager, error) {
	state, err := InitStateFromDB(sDB)
	if err != nil {
		return nil, err
	}
	return NewManagerWithValue(state), nil
}

func InitStateFromDB(sDB *statedb.StateDB) (*State, error) {
	// load list all unified tokens
	unifiedTokenStates, err := statedb.GetBridgeAggUnifiedTokens(sDB)
	if err != nil {
		return nil, err
	}

	// load unified token infos
	unifiedTokenInfos := make(map[common.Hash]map[common.Hash]*statedb.BridgeAggVaultState)
	for _, unifiedTokenState := range unifiedTokenStates {
		unifiedTokenInfos[unifiedTokenState.TokenID()] = make(map[common.Hash]*statedb.BridgeAggVaultState)
		vaults, err := statedb.GetBridgeAggVaults(sDB, unifiedTokenState.TokenID())
		if err != nil {
			return nil, err
		}
		for tokenID, vault := range vaults {
			unifiedTokenInfos[unifiedTokenState.TokenID()][tokenID] = vault
		}
	}

	// load waiting unshield reqs
	waitingUnshieldReqs := make(map[common.Hash][]*statedb.BridgeAggWaitingUnshieldReq)
	for _, unifiedTokenState := range unifiedTokenStates {
		unifiedTokenID := unifiedTokenState.TokenID()
		reqs, err := statedb.GetBridgeAggWaitingUnshieldReqs(sDB, unifiedTokenID)
		if err != nil {
			return nil, err
		}
		if len(reqs) > 0 {
			waitingUnshieldReqs[unifiedTokenID] = reqs
		}
	}

	return NewStateWithValue(
		unifiedTokenInfos,
		waitingUnshieldReqs,
		map[common.Hash][]*statedb.BridgeAggWaitingUnshieldReq{},
		[]common.Hash{}), nil
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
