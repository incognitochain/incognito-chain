package statedb

import "fmt"

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
