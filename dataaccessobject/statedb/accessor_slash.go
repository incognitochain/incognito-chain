package statedb

func GetProducersBlackList(stateDB *StateDB, beaconHeight uint64) map[string]uint8 {
	return stateDB.GetAllProducerBlackList()
}

func StoreProducersBlackList(stateDB *StateDB, beaconHeight uint64, producersBlackList map[string]uint8) error {
	for producerKey, punishedEpoches := range producersBlackList {
		key := GenerateBlackListProducerObjectKey(producerKey)
		value := NewBlackListProducerStateWithValue(producerKey, punishedEpoches, beaconHeight)
		err := stateDB.SetStateObject(BlackListProducerObjectType, key, value)
		if err != nil {
			return NewStatedbError(StoreBlackListProducersError, err)
		}
	}
	return nil
}

func RemoveProducerBlackList(stateDB *StateDB, blackListProducerKeyList []string) {
	for _, blackListProducerKey := range blackListProducerKeyList {
		key := GenerateBlackListProducerObjectKey(blackListProducerKey)
		stateDB.MarkDeleteStateObject(BlackListProducerObjectType, key)
	}
}

func RemoveAllEmptyProducerBlackList(stateDB *StateDB, blackListProducerKey string) {
	m := stateDB.GetAllProducerBlackListState()
	for key, value := range m {
		if value == 0 {
			stateDB.MarkDeleteStateObject(BlackListProducerObjectType, key)
		}
	}
}
