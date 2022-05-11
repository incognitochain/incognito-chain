package statedb

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
)

func TrackBridgeAggStatus(stateDB *StateDB, statusType []byte, statusSuffix []byte, statusContent []byte) error {
	key := GenerateBridgeAggStatusObjectKey(statusType, statusSuffix)
	value := NewBridgeAggStatusStateWithValue(statusType, statusSuffix, statusContent)
	err := stateDB.SetStateObject(BridgeAggStatusObjectType, key, value)
	if err != nil {
		return NewStatedbError(StoreBridgeAggStatusError, err)
	}
	return nil
}

func GetBridgeAggStatus(stateDB *StateDB, statusType []byte, statusSuffix []byte) ([]byte, error) {
	key := GenerateBridgeAggStatusObjectKey(statusType, statusSuffix)
	s, has, err := stateDB.getBridgeAggStatusByKey(key)
	if err != nil {
		return []byte{}, NewStatedbError(GetBridgeAggStatusError, err)
	}
	if !has {
		return []byte{}, NewStatedbError(GetBridgeAggStatusError, fmt.Errorf("status %+v with prefix %+v not found", string(statusType), statusSuffix))
	}
	return s.statusContent, nil
}

func StoreBridgeAggUnifiedToken(stateDB *StateDB, unifiedTokenID common.Hash, state *BridgeAggUnifiedTokenState) error {
	key := GenerateBridgeAggUnifiedTokenObjectKey(unifiedTokenID)
	return stateDB.SetStateObject(BridgeAggUnifiedTokenObjectType, key, state)
}

func StoreBridgeAggVault(stateDB *StateDB, unifiedTokenID, tokenID common.Hash, state *BridgeAggVaultState) error {
	key := GenerateBridgeAggVaultObjectKey(unifiedTokenID, tokenID)
	return stateDB.SetStateObject(BridgeAggVaultObjectType, key, state)
}

func GetBridgeAggUnifiedTokens(stateDB *StateDB) ([]*BridgeAggUnifiedTokenState, error) {
	prefixHash := generateBridgeAggUnifiedTokenObjectPrefix()
	return stateDB.iterateBridgeAggUnifiedTokens(prefixHash)
}

func GetBridgeAggVaults(stateDB *StateDB, unifiedTokenID common.Hash) (map[common.Hash]*BridgeAggVaultState, error) {
	prefixHash := generateBridgeAggVaultObjectPrefix(unifiedTokenID)
	return stateDB.iterateBridgeAggVaults(prefixHash)
}
