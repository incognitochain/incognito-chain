package bridgeagg

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

func CheckTokenIDExisted(sDBs map[int]*statedb.StateDB, tokenID common.Hash) error {
	for _, sDB := range sDBs {
		if statedb.PrivacyTokenIDExisted(sDB, tokenID) {
			return nil
		}
	}
	return fmt.Errorf("Cannot find tokenID %s in network", tokenID.String())
}

func InitStateFromDB(sDB *statedb.StateDB) (*State, error) {
	unifiedTokenStates, err := statedb.GetBridgeAggUnifiedTokens(sDB)
	if err != nil {
		return nil, err
	}
	unifiedTokenInfos := make(map[common.Hash]map[common.Hash]*Vault)
	for _, unifiedTokenState := range unifiedTokenStates {
		unifiedTokenInfos[unifiedTokenState.TokenID()] = make(map[common.Hash]*Vault)
		convertTokens, err := statedb.GetBridgeAggConvertedTokens(sDB, unifiedTokenState.TokenID())
		if err != nil {
			return nil, err
		}
		for _, convertToken := range convertTokens {
			unifiedTokenInfos[unifiedTokenState.TokenID()][convertToken.TokenID()] = NewVault()
			//TODO: Get vault state from db here
		}
	}
	return NewStateWithValue(unifiedTokenInfos), nil
}
