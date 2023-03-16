package statedb

import (
	"fmt"
)

// Get bridge hub param from statedb
func GetBridgeHubParam(stateDB *StateDB) (*BridgeHubParamState, error) {
	key := GenerateBridgeHubParamObjectKey()
	param, has, err := stateDB.getBridgeHubParamByKey(key)
	if err != nil {
		return nil, NewStatedbError(GetBridgeHubStatusError, err)
	}
	if !has {
		return nil, nil
	}
	return param, nil
}

func StoreBridgeHubParam(stateDB *StateDB, param *BridgeHubParamState) error {
	key := GenerateBridgeHubParamObjectKey()
	return stateDB.SetStateObject(BridgeHubParamObjectType, key, param)
}

func TrackBridgeHubStatus(stateDB *StateDB, statusType []byte, statusSuffix []byte, statusContent []byte) error {
	key := GenerateBridgeHubStatusObjectKey(statusType, statusSuffix)
	value := NewBridgeHubStatusStateWithValue(statusType, statusSuffix, statusContent)
	err := stateDB.SetStateObject(BridgeHubStatusObjectType, key, value)
	if err != nil {
		return NewStatedbError(StoreBridgeHubStatusError, err)
	}
	return nil
}

func GetBridgeHubStatus(stateDB *StateDB, statusType []byte, statusSuffix []byte) ([]byte, error) {
	key := GenerateBridgeHubStatusObjectKey(statusType, statusSuffix)
	s, has, err := stateDB.getBridgeHubStatusByKey(key)
	if err != nil {
		return []byte{}, NewStatedbError(GetBridgeHubStatusError, err)
	}
	if !has {
		return []byte{}, NewStatedbError(GetBridgeHubStatusError, fmt.Errorf("status %+v with prefix %+v not found", string(statusType), statusSuffix))
	}
	return s.statusContent, nil
}

func StoreBridgeHubBridgeInfo(stateDB *StateDB, bridgeID string, state *BridgeInfoState) error {
	key := GenerateBridgeHubBridgeInfoObjectKey(bridgeID)
	return stateDB.SetStateObject(BridgeHubBridgeInfoObjectType, key, state)
}

func GetBridgeHubBridgeInfo(stateDB *StateDB) ([]*BridgeInfoState, error) {
	prefixHash := GetBridgeHubBridgeInfoPrefix()
	return stateDB.iterateBridgeHubBridgeInfos(prefixHash)
}

func StoreBridgeHubPTokenAmount(stateDB *StateDB, bridgeID string, pTokenID string, state *BridgeHubPTokenState) error {
	key := GenerateBridgeHubPTokenObjectKey(bridgeID, pTokenID)
	return stateDB.SetStateObject(BridgeHubPTokenObjectType, key, state)
}

func GetBridgeHubPTokenByBridgeID(stateDB *StateDB, bridgeID string) ([]*BridgeHubPTokenState, error) {
	prefixHash := GetBridgeHubPTokenPrefix([]byte(bridgeID))
	return stateDB.iterateBridgeHubPTokens(prefixHash)
}

func StoreBridgeHubStaking(stateDB *StateDB, validator string, state *BridgeStakingInfoState) error {
	key := GenerateBridgeHubStakingInfoObjectKey(validator)
	return stateDB.SetStateObject(BridgeStakingTxObjectType, key, state)
}

func GetBridgeStakingInfo(stateDB *StateDB) ([]*BridgeStakingInfoState, error) {
	prefixHash := GetBridgeHubStakingTxPrefix()
	return stateDB.iterateBridgeHubBridgeStakingInfos(prefixHash)
}
