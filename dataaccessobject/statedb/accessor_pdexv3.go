package statedb

import "fmt"

func TrackPDexV3Status(stateDB *StateDB, statusType []byte, statusSuffix []byte, statusContent []byte) error {
	key := GeneratePDexV3StatusObjectKey(statusType, statusSuffix)
	value := NewPDexV3StatusStateWithValue(statusType, statusSuffix, statusContent)
	err := stateDB.SetStateObject(PDexV3StatusObjectType, key, value)
	if err != nil {
		return NewStatedbError(StorePDexV3StatusError, err)
	}
	return nil
}

func GetPDexV3Status(stateDB *StateDB, statusType []byte, statusSuffix []byte) ([]byte, error) {
	key := GeneratePDexV3StatusObjectKey(statusType, statusSuffix)
	s, has, err := stateDB.getPDexV3StatusByKey(key)
	if err != nil {
		return []byte{}, NewStatedbError(GetPDexV3StatusError, err)
	}
	if !has {
		return []byte{}, NewStatedbError(GetPDexV3StatusNotFoundError, fmt.Errorf("status %+v with prefix %+v not found", string(statusType), string(statusSuffix)))
	}
	return s.statusContent, nil
}
