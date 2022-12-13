package statedb

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"sort"
)

func StoreBeaconStakerInfo(
	stateDB *StateDB,
	committee incognitokey.CommitteePublicKey,
	info BeaconStakerInfo,
) error {
	keyBytes, err := committee.RawBytes()
	if err != nil {
		return err
	}
	key := GetBeaconStakerInfoKey(keyBytes)
	err = stateDB.SetStateObject(ShardStakerObjectType, key, info)
	if err != nil {
		return err
	}
	return nil
}

func GetBeaconStakerInfo(stateDB *StateDB, beaconStakerPubkey string) (*BeaconStakerInfo, bool, error) {
	pubKey := incognitokey.NewCommitteePublicKey()
	err := pubKey.FromString(beaconStakerPubkey)
	if err != nil {
		return nil, false, err
	}
	pubKeyBytes, _ := pubKey.RawBytes()
	key := GetBeaconStakerInfoKey(pubKeyBytes)
	return stateDB.getBeaconStakerInfo(key)
}

func DeleteBeaconStakerInfo(stateDB *StateDB, stakers []incognitokey.CommitteePublicKey) error {
	return deleteBeaconStakerInfo(stateDB, stakers)
}

func deleteBeaconStakerInfo(stateDB *StateDB, stakers []incognitokey.CommitteePublicKey) error {
	for _, staker := range stakers {
		keyBytes, err := staker.RawBytes()
		if err != nil {
			return err
		}
		key := GetBeaconStakerInfoKey(keyBytes)
		stateDB.MarkDeleteStateObject(ShardStakerObjectType, key)
	}
	return nil
}

func GetBeaconCommitteeEnterTime(stateDB *StateDB) []int64 {
	m := stateDB.getAllValidatorCommitteePublicKey(CurrentValidator, []int{BeaconChainID})
	tempBeaconCommitteeStates := m[BeaconChainID]
	sort.Slice(tempBeaconCommitteeStates, func(i, j int) bool {
		return tempBeaconCommitteeStates[i].EnterTime() < tempBeaconCommitteeStates[j].EnterTime()
	})
	list := []int64{}
	for _, tempBeaconCommitteeState := range tempBeaconCommitteeStates {
		list = append(list, tempBeaconCommitteeState.EnterTime())
	}
	return list
}

func ReplaceBeaconCommittee(stateDB *StateDB, beaconCommittees [2][]incognitokey.CommitteePublicKey) error {
	if len(beaconCommittees[common.REPLACE_IN]) == 0 {
		return nil
	}
	// for beaconCommittees
	newEnterTime := GetBeaconCommitteeEnterTime(stateDB)
	err := storeCommittee(stateDB, BeaconChainID, CurrentValidator, beaconCommittees[common.REPLACE_IN], newEnterTime)
	if err != nil {
		return NewStatedbError(StoreBeaconCommitteeError, err)
	}
	err = deleteCommittee(stateDB, BeaconChainID, CurrentValidator, beaconCommittees[common.REPLACE_OUT])
	if err != nil {
		return NewStatedbError(DeleteBeaconCommitteeError, err)
	}
	return nil
}

func StoreCurrentEpochBeaconCandidate(stateDB *StateDB, candidate []incognitokey.CommitteePublicKey) error {
	err := storeCommittee(stateDB, BeaconChainID, CurrentEpochBeaconCandidate, candidate, defaultEnterTime)
	if err != nil {
		return NewStatedbError(StoreCurrentEpochCandidateError, err)
	}
	return nil
}

/*
Beacon Committee
*/
func StoreBeaconCommittee(stateDB *StateDB, beaconCommittees []incognitokey.CommitteePublicKey) error {
	err := storeCommittee(stateDB, BeaconChainID, CurrentValidator, beaconCommittees, defaultEnterTime)
	if err != nil {
		return NewStatedbError(StoreBeaconCommitteeError, err)
	}
	return nil
}
func GetBeaconCommittee(stateDB *StateDB) []incognitokey.CommitteePublicKey {
	m := stateDB.getAllValidatorCommitteePublicKey(CurrentValidator, []int{BeaconChainID})
	tempBeaconCommitteeStates := m[BeaconChainID]
	sort.Slice(tempBeaconCommitteeStates, func(i, j int) bool {
		return tempBeaconCommitteeStates[i].EnterTime() < tempBeaconCommitteeStates[j].EnterTime()
	})
	list := []incognitokey.CommitteePublicKey{}
	for _, tempBeaconCommitteeState := range tempBeaconCommitteeStates {
		list = append(list, tempBeaconCommitteeState.CommitteePublicKey())
	}
	return list
}
func DeleteBeaconCommittee(stateDB *StateDB, beaconCommittees []incognitokey.CommitteePublicKey) error {
	err := deleteCommittee(stateDB, BeaconChainID, CurrentValidator, beaconCommittees)
	if err != nil {
		return NewStatedbError(DeleteBeaconCommitteeError, err)
	}
	return nil
}

/*
Beacon Pending
*/
func StoreBeaconSubstituteValidator(stateDB *StateDB, beaconSubstitute []incognitokey.CommitteePublicKey) error {
	err := storeCommittee(stateDB, BeaconChainID, SubstituteValidator, beaconSubstitute, defaultEnterTime)
	if err != nil {
		return NewStatedbError(StoreBeaconSubstitutesValidatorError, err)
	}
	return nil
}
func GetBeaconSubstituteValidator(stateDB *StateDB) []incognitokey.CommitteePublicKey {
	m := stateDB.getAllValidatorCommitteePublicKey(SubstituteValidator, []int{BeaconChainID})
	tempBeaconCommitteeStates := m[BeaconChainID]
	sort.Slice(tempBeaconCommitteeStates, func(i, j int) bool {
		return tempBeaconCommitteeStates[i].EnterTime() < tempBeaconCommitteeStates[j].EnterTime()
	})
	list := []incognitokey.CommitteePublicKey{}
	for _, tempBeaconCommitteeState := range tempBeaconCommitteeStates {
		list = append(list, tempBeaconCommitteeState.CommitteePublicKey())
	}
	return list
}
func DeleteBeaconSubstituteValidator(stateDB *StateDB, beaconSubstitute []incognitokey.CommitteePublicKey) error {
	err := deleteCommittee(stateDB, BeaconChainID, SubstituteValidator, beaconSubstitute)
	if err != nil {
		return NewStatedbError(DeleteBeaconSubstituteValidatorError, err)
	}
	return nil
}

/*
Beacon Waiting
*/
func StoreBeaconWaiting(stateDB *StateDB, members []incognitokey.CommitteePublicKey) error {
	err := storeCommittee(stateDB, BeaconChainID, BeaconWaitingPool, members, defaultEnterTime)
	if err != nil {
		return NewStatedbError(StoreBeaconWaitingError, err)
	}
	return nil
}
func GetBeaconWaiting(stateDB *StateDB) []incognitokey.CommitteePublicKey {
	m := stateDB.getAllValidatorCommitteePublicKey(BeaconWaitingPool, []int{BeaconChainID})
	tempBeaconCommitteeStates := m[BeaconChainID]
	sort.Slice(tempBeaconCommitteeStates, func(i, j int) bool {
		return tempBeaconCommitteeStates[i].EnterTime() < tempBeaconCommitteeStates[j].EnterTime()
	})
	list := []incognitokey.CommitteePublicKey{}
	for _, tempBeaconCommitteeState := range tempBeaconCommitteeStates {
		list = append(list, tempBeaconCommitteeState.CommitteePublicKey())
	}
	return list
}
func DeleteBeaconWaiting(stateDB *StateDB, beaconWaiting []incognitokey.CommitteePublicKey) error {
	err := deleteCommittee(stateDB, BeaconChainID, BeaconWaitingPool, beaconWaiting)
	if err != nil {
		return NewStatedbError(DeleteBeaconWaitingError, err)
	}
	return nil
}

/*
Beacon Locking
*/
func StoreBeaconLocking(stateDB *StateDB, members []incognitokey.CommitteePublicKey) error {
	err := storeCommittee(stateDB, BeaconChainID, BeaconLockingPool, members, defaultEnterTime)
	if err != nil {
		return NewStatedbError(StoreBeaconLockingError, err)
	}
	return nil
}
func GetBeaconLocking(stateDB *StateDB) []incognitokey.CommitteePublicKey {
	m := stateDB.getAllValidatorCommitteePublicKey(BeaconLockingPool, []int{BeaconChainID})
	tempBeaconCommitteeStates := m[BeaconChainID]
	sort.Slice(tempBeaconCommitteeStates, func(i, j int) bool {
		return tempBeaconCommitteeStates[i].EnterTime() < tempBeaconCommitteeStates[j].EnterTime()
	})
	list := []incognitokey.CommitteePublicKey{}
	for _, tempBeaconCommitteeState := range tempBeaconCommitteeStates {
		list = append(list, tempBeaconCommitteeState.CommitteePublicKey())
	}
	return list
}
func DeleteBeaconLocking(stateDB *StateDB, beaconLocking []incognitokey.CommitteePublicKey) error {
	err := deleteCommittee(stateDB, BeaconChainID, BeaconLockingPool, beaconLocking)
	if err != nil {
		return NewStatedbError(DeleteBeaconLockingError, err)
	}
	return nil
}
