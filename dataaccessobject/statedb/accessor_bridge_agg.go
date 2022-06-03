package statedb

import (
	"errors"
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

func StoreBridgeAggConvertedToken(stateDB *StateDB, unifiedTokenID, tokenID common.Hash, state *BridgeAggConvertedTokenState) error {
	key := GenerateBridgeAggConvertedTokenObjectKey(unifiedTokenID, tokenID)
	return stateDB.SetStateObject(BridgeAggConvertedTokenObjectType, key, state)
}

func StoreBridgeAggVault(stateDB *StateDB, unifiedTokenID, tokenID common.Hash, state *BridgeAggVaultState) error {
	key := GenerateBridgeAggVaultObjectKey(unifiedTokenID, tokenID)
	return stateDB.SetStateObject(BridgeAggVaultObjectType, key, state)
}

func GetBridgeAggUnifiedTokens(stateDB *StateDB) ([]*BridgeAggUnifiedTokenState, error) {
	prefixHash := generateBridgeAggUnifiedTokenObjectPrefix()
	return stateDB.iterateWithBridgeAggUnifiedTokens(prefixHash)
}

func GetBridgeAggConvertedTokens(stateDB *StateDB, unifiedTokenID common.Hash) ([]*BridgeAggConvertedTokenState, error) {
	prefixHash := generateBridgeAggConvertedTokenObjectPrefix(unifiedTokenID)
	return stateDB.iterateWithBridgeAggConvertedTokens(prefixHash)
}

func GetBridgeAggVault(stateDB *StateDB, unifiedTokenID, tokenID common.Hash) (*BridgeAggVaultState, error) {
	key := GenerateBridgeAggVaultObjectKey(unifiedTokenID, tokenID)
	state, ok, err := stateDB.getBridgeAggVault(key)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("Can't find bridge agg vault")
	}
	return state, nil
}
