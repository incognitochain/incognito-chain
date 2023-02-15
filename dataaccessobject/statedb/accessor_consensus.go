package statedb

import (
	"sort"
	"time"

	"github.com/incognitochain/incognito-chain/privacy/key"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

var defaultEnterTime = []int64{}

func storeCommittee(stateDB *StateDB, shardID int, role int, committees []incognitokey.CommitteePublicKey, newEnterTime []int64) error {
	enterTime := time.Now().UnixNano()
	for id, committee := range committees {
		if (len(newEnterTime) != 0) && (id < len(newEnterTime)) {
			enterTime = newEnterTime[id]
		}
		key, err := GenerateCommitteeObjectKeyWithRole(role, shardID, committee)
		if err != nil {
			return err
		}
		value := NewCommitteeState()
		has := false
		value, has, err = stateDB.getCommitteeState(key)
		if err != nil {
			return err
		}
		if !has {
			value = NewCommitteeStateWithValueAndTime(shardID, role, committee, enterTime)
		}
		err = stateDB.SetStateObject(CommitteeObjectType, key, value)
		if err != nil {
			return err
		}
		enterTime++
	}
	return nil
}

func storeCommitteeWithNewValue(stateDB *StateDB, shardID int, role int, committees []incognitokey.CommitteePublicKey, newEnterTime []int64) error {
	enterTime := time.Now().UnixNano()
	for id, committee := range committees {
		if (len(newEnterTime) != 0) && (id < len(newEnterTime)) {
			enterTime = newEnterTime[id]
		}
		key, err := GenerateCommitteeObjectKeyWithRole(role, shardID, committee)
		if err != nil {
			return err
		}
		value := NewCommitteeStateWithValueAndTime(shardID, role, committee, enterTime)
		err = stateDB.SetStateObject(CommitteeObjectType, key, value)
		if err != nil {
			return err
		}
		enterTime++
	}
	return nil
}

func StoreSyncingValidators(stateDB *StateDB, syncingValidators map[byte][]incognitokey.CommitteePublicKey) error {
	for shardID, singleChainSyncingValidators := range syncingValidators {
		err := storeCommittee(stateDB, int(shardID), SyncingValidators, singleChainSyncingValidators, defaultEnterTime)
		if err != nil {
			return NewStatedbError(StoreSyncingValidatorsError, err)
		}
	}
	return nil
}

func StoreOneShardCommittee(stateDB *StateDB, shardID byte, shardCommittees []incognitokey.CommitteePublicKey) error {
	err := storeCommittee(stateDB, int(shardID), CurrentValidator, shardCommittees, defaultEnterTime)
	if err != nil {
		return NewStatedbError(StoreShardCommitteeError, err)
	}
	return nil
}

func StoreAllShardCommittee(stateDB *StateDB, allShardCommittees map[byte][]incognitokey.CommitteePublicKey) error {
	for shardID, committee := range allShardCommittees {
		err := storeCommittee(stateDB, int(shardID), CurrentValidator, committee, defaultEnterTime)
		if err != nil {
			return NewStatedbError(StoreAllShardCommitteeError, err)
		}
	}
	return nil
}

func ReplaceOneShardCommittee(stateDB *StateDB, shardID byte, shardCommittee [2][]incognitokey.CommitteePublicKey) error {
	if len(shardCommittee[common.REPLACE_IN]) == 0 {
		return nil
	}
	newEnterTime := GetOneShardCommitteeEnterTime(stateDB, shardID)
	err := storeCommittee(stateDB, int(shardID), CurrentValidator, shardCommittee[common.REPLACE_IN], newEnterTime)
	if err != nil {
		return NewStatedbError(StoreAllShardCommitteeError, err)
	}
	err = deleteCommittee(stateDB, int(shardID), CurrentValidator, shardCommittee[common.REPLACE_OUT])
	if err != nil {
		return NewStatedbError(DeleteOneShardCommitteeError, err)
	}
	return nil
}

func ReplaceAllShardCommittee(stateDB *StateDB, allShardCommittees map[byte][2][]incognitokey.CommitteePublicKey) error {
	for shardID, committee := range allShardCommittees {
		if len(committee[common.REPLACE_IN]) == 0 {
			continue
		}
		newEnterTime := GetOneShardCommitteeEnterTime(stateDB, shardID)
		err := storeCommittee(stateDB, int(shardID), CurrentValidator, committee[common.REPLACE_IN], newEnterTime)
		if err != nil {
			return NewStatedbError(StoreAllShardCommitteeError, err)
		}
		err = deleteCommittee(stateDB, int(shardID), CurrentValidator, committee[common.REPLACE_OUT])
		if err != nil {
			return NewStatedbError(DeleteOneShardCommitteeError, err)
		}
	}
	return nil
}

func IsInShardCandidateForNextEpoch(
	stateDB *StateDB,
	committee incognitokey.CommitteePublicKey,
) (*CommitteeState, bool, error) {
	key, err := GenerateCommitteeObjectKeyWithRole(NextEpochShardCandidate, CandidateChainID, committee)
	if err != nil {
		return nil, false, err
	}

	return stateDB.getCommitteeState(key)

}

func IsInShardCandidateForCurrentEpoch(
	stateDB *StateDB,
	committee incognitokey.CommitteePublicKey,
) (*CommitteeState, bool, error) {
	key, err := GenerateCommitteeObjectKeyWithRole(CurrentEpochShardCandidate, CandidateChainID, committee)
	if err != nil {
		return nil, false, err
	}

	return stateDB.getCommitteeState(key)

}

func StoreNextEpochShardCandidate(
	stateDB *StateDB,
	candidate []incognitokey.CommitteePublicKey,
	rewardReceiver map[string]key.PaymentAddress,
	// funderAddress map[string]key.PaymentAddress,
	autoStaking map[string]bool,
	stakingTx map[string]common.Hash,
	// amountStaking map[string]uint64,
) error {
	err1 := storeCommittee(stateDB, CandidateChainID, NextEpochShardCandidate, candidate, defaultEnterTime)
	if err1 != nil {
		return NewStatedbError(StoreNextEpochCandidateError, err1)
	}
	return nil
}

func StoreMembersAtCommonShardPool(
	stateDB *StateDB,
	members []incognitokey.CommitteePublicKey,
) error {
	err := storeCommittee(stateDB, CandidateChainID, CommonShardPool, members, defaultEnterTime)
	if err != nil {
		return NewStatedbError(StoreMemberCommonShardPoolError, err)
	}
	return nil
}

func StoreMembersAtShardPool(
	stateDB *StateDB,
	shardID byte,
	members []incognitokey.CommitteePublicKey,
) error {
	err := storeCommittee(stateDB, int(shardID), ShardPool, members, defaultEnterTime)
	if err != nil {
		return NewStatedbError(StoreMemberShardPoolError, err)
	}
	return nil
}

func StoreCurrentEpochShardCandidate(stateDB *StateDB, candidate []incognitokey.CommitteePublicKey) error {
	err := storeCommittee(stateDB, CandidateChainID, CurrentEpochShardCandidate, candidate, defaultEnterTime)
	if err != nil {
		return NewStatedbError(StoreCurrentEpochCandidateError, err)
	}
	return nil
}

func StoreNextEpochBeaconCandidate(
	stateDB *StateDB,
	candidate []incognitokey.CommitteePublicKey,
	rewardReceiver map[string]key.PaymentAddress,
	autoStaking map[string]bool,
	stakingTx map[string]common.Hash,
) error {
	err1 := storeCommittee(stateDB, BeaconChainID, NextEpochBeaconCandidate, candidate, defaultEnterTime)
	if err1 != nil {
		return NewStatedbError(StoreNextEpochCandidateError, err1)
	}
	return nil
}

func StoreOneShardSubstitutesValidator(stateDB *StateDB, shardID byte, shardSubstitutes []incognitokey.CommitteePublicKey) error {
	err := storeCommittee(stateDB, int(shardID), SubstituteValidator, shardSubstitutes, defaultEnterTime)
	if err != nil {
		return NewStatedbError(StoreOneShardSubstitutesValidatorError, err)
	}
	return nil
}

func StoreOneShardSubstitutesValidatorV3(stateDB *StateDB, shardID byte, shardSubstitutes []incognitokey.CommitteePublicKey) error {
	err := storeCommitteeWithNewValue(stateDB, int(shardID), SubstituteValidator, shardSubstitutes, defaultEnterTime)
	if err != nil {
		return NewStatedbError(StoreOneShardSubstitutesValidatorError, err)
	}
	return nil
}

func GetOneShardCommittee(stateDB *StateDB, shardID byte) []incognitokey.CommitteePublicKey {
	tempShardCommitteeStates := stateDB.getByShardIDCurrentValidatorState(int(shardID))
	sort.Slice(tempShardCommitteeStates, func(i, j int) bool {
		return tempShardCommitteeStates[i].EnterTime() < tempShardCommitteeStates[j].EnterTime()
	})
	list := []incognitokey.CommitteePublicKey{}
	for _, tempShardCommitteeState := range tempShardCommitteeStates {
		list = append(list, tempShardCommitteeState.CommitteePublicKey())
	}
	return list
}

func GetOneShardSubstituteValidator(stateDB *StateDB, shardID byte) []incognitokey.CommitteePublicKey {
	tempShardCommitteeStates := stateDB.getByShardIDSubstituteValidatorState(int(shardID))
	sort.Slice(tempShardCommitteeStates, func(i, j int) bool {
		return tempShardCommitteeStates[i].EnterTime() < tempShardCommitteeStates[j].EnterTime()
	})
	list := []incognitokey.CommitteePublicKey{}
	for _, tempShardCommitteeState := range tempShardCommitteeStates {
		list = append(list, tempShardCommitteeState.CommitteePublicKey())
	}
	return list
}

func GetAllShardCommittee(stateDB *StateDB, shardIDs []int) map[int][]incognitokey.CommitteePublicKey {
	tempM := stateDB.getAllValidatorCommitteePublicKey(CurrentValidator, shardIDs)
	m := make(map[int][]incognitokey.CommitteePublicKey)
	for shardID, tempShardCommitteeStates := range tempM {
		sort.Slice(tempShardCommitteeStates, func(i, j int) bool {
			return tempShardCommitteeStates[i].EnterTime() < tempShardCommitteeStates[j].EnterTime()
		})
		list := []incognitokey.CommitteePublicKey{}
		for _, tempShardCommitteeState := range tempShardCommitteeStates {
			list = append(list, tempShardCommitteeState.CommitteePublicKey())
		}
		m[shardID] = list
	}
	return m
}

func GetAllShardSubstituteValidator(stateDB *StateDB, shardIDs []int) map[int][]incognitokey.CommitteePublicKey {
	tempM := stateDB.getAllValidatorCommitteePublicKey(SubstituteValidator, shardIDs)
	m := make(map[int][]incognitokey.CommitteePublicKey)
	for shardID, tempShardCommitteeStates := range tempM {
		sort.Slice(tempShardCommitteeStates, func(i, j int) bool {
			return tempShardCommitteeStates[i].EnterTime() < tempShardCommitteeStates[j].EnterTime()
		})
		list := []incognitokey.CommitteePublicKey{}
		for _, tempShardCommitteeState := range tempShardCommitteeStates {
			list = append(list, tempShardCommitteeState.CommitteePublicKey())
		}
		m[shardID] = list
	}
	return m
}

func GetNextEpochCandidate(stateDB *StateDB) []incognitokey.CommitteePublicKey {
	tempNextEpochShardCandidateStates := stateDB.getAllCandidateCommitteePublicKey(NextEpochShardCandidate)
	sort.Slice(tempNextEpochShardCandidateStates, func(i, j int) bool {
		return tempNextEpochShardCandidateStates[i].EnterTime() < tempNextEpochShardCandidateStates[j].EnterTime()
	})
	list := []incognitokey.CommitteePublicKey{}
	for _, tempNextEpochShardCandidateStates := range tempNextEpochShardCandidateStates {
		list = append(list, tempNextEpochShardCandidateStates.CommitteePublicKey())
	}
	return list
}

func GetCurrentEpochCandidate(stateDB *StateDB) []incognitokey.CommitteePublicKey {
	tempCurrentEpochShardCandidateStates := stateDB.getAllCandidateCommitteePublicKey(CurrentEpochShardCandidate)
	sort.Slice(tempCurrentEpochShardCandidateStates, func(i, j int) bool {
		return tempCurrentEpochShardCandidateStates[i].EnterTime() < tempCurrentEpochShardCandidateStates[j].EnterTime()
	})
	list := []incognitokey.CommitteePublicKey{}
	for _, tempCurrentEpochShardCandidateState := range tempCurrentEpochShardCandidateStates {
		list = append(list, tempCurrentEpochShardCandidateState.CommitteePublicKey())
	}
	return list
}

// staking flow v4 will not worl with shardIDs contains -1
func GetAllCandidateSubstituteCommittee(stateDB *StateDB, shardIDs []int) (
	map[int][]incognitokey.CommitteePublicKey,
	map[int][]incognitokey.CommitteePublicKey,
	[]incognitokey.CommitteePublicKey,
	[]incognitokey.CommitteePublicKey,
	[]incognitokey.CommitteePublicKey,
	[]incognitokey.CommitteePublicKey,
	map[byte][]incognitokey.CommitteePublicKey,
	map[string]key.PaymentAddress,
	map[string]bool,
	map[string]common.Hash,
	map[string][]string,
) {
	tempCurrentValidator, tempSubstituteValidator, tempNextEpochShardCandidate,
		tempCurrentEpochShardCandidate, tempNextEpochBeaconCandidate, tempCurrentEpochBeaconCandidate,
		tempSyncingValidators,
		rewardReceivers, autoStaking, stakingTx, beaconDelegate := stateDB.getAllCommitteeState(shardIDs)
	currentValidator := make(map[int][]incognitokey.CommitteePublicKey)
	substituteValidator := make(map[int][]incognitokey.CommitteePublicKey)
	nextEpochShardCandidate := []incognitokey.CommitteePublicKey{}
	currentEpochShardCandidate := []incognitokey.CommitteePublicKey{}
	nextEpochBeaconCandidate := []incognitokey.CommitteePublicKey{}
	currentEpochBeaconCandidate := []incognitokey.CommitteePublicKey{}
	syncingValidators := make(map[byte][]incognitokey.CommitteePublicKey)

	for shardID, tempShardCommitteeStates := range tempCurrentValidator {
		sort.Slice(tempShardCommitteeStates, func(i, j int) bool {
			return tempShardCommitteeStates[i].EnterTime() < tempShardCommitteeStates[j].EnterTime()
		})
		list := []incognitokey.CommitteePublicKey{}
		for _, tempShardCommitteeState := range tempShardCommitteeStates {
			list = append(list, tempShardCommitteeState.CommitteePublicKey())
		}
		currentValidator[shardID] = list
	}

	for shardID, tempShardSubstituteStates := range tempSubstituteValidator {
		sort.Slice(tempShardSubstituteStates, func(i, j int) bool {
			return tempShardSubstituteStates[i].EnterTime() < tempShardSubstituteStates[j].EnterTime()
		})
		list := []incognitokey.CommitteePublicKey{}
		for _, tempShardCommitteeState := range tempShardSubstituteStates {
			list = append(list, tempShardCommitteeState.CommitteePublicKey())
		}
		substituteValidator[shardID] = list
	}

	for shardID, singleChainSyncingValidators := range tempSyncingValidators {
		sort.Slice(singleChainSyncingValidators, func(i, j int) bool {
			return singleChainSyncingValidators[i].EnterTime() < singleChainSyncingValidators[j].EnterTime()
		})
		list := []incognitokey.CommitteePublicKey{}
		for _, syncingValidator := range singleChainSyncingValidators {
			list = append(list, syncingValidator.CommitteePublicKey())
		}
		syncingValidators[byte(shardID)] = list
	}

	sort.Slice(tempNextEpochShardCandidate, func(i, j int) bool {
		return tempNextEpochShardCandidate[i].EnterTime() < tempNextEpochShardCandidate[j].EnterTime()
	})

	for _, candidate := range tempNextEpochShardCandidate {
		nextEpochShardCandidate = append(nextEpochShardCandidate, candidate.CommitteePublicKey())
	}

	sort.Slice(tempCurrentEpochShardCandidate, func(i, j int) bool {
		return tempCurrentEpochShardCandidate[i].EnterTime() < tempCurrentEpochShardCandidate[j].EnterTime()
	})

	for _, candidate := range tempCurrentEpochShardCandidate {
		currentEpochShardCandidate = append(currentEpochShardCandidate, candidate.CommitteePublicKey())
	}

	sort.Slice(tempNextEpochBeaconCandidate, func(i, j int) bool {
		return tempNextEpochBeaconCandidate[i].EnterTime() < tempNextEpochBeaconCandidate[j].EnterTime()
	})

	for _, candidate := range tempNextEpochBeaconCandidate {
		nextEpochBeaconCandidate = append(nextEpochBeaconCandidate, candidate.CommitteePublicKey())
	}

	sort.Slice(tempCurrentEpochBeaconCandidate, func(i, j int) bool {
		return tempCurrentEpochBeaconCandidate[i].EnterTime() < tempCurrentEpochBeaconCandidate[j].EnterTime()
	})

	for _, candidate := range tempCurrentEpochBeaconCandidate {
		currentEpochBeaconCandidate = append(currentEpochBeaconCandidate, candidate.CommitteePublicKey())
	}

	return currentValidator, substituteValidator, nextEpochShardCandidate,
		currentEpochShardCandidate, nextEpochBeaconCandidate, currentEpochBeaconCandidate,
		syncingValidators, rewardReceivers, autoStaking, stakingTx, beaconDelegate
}

func GetAllCommitteeState(stateDB *StateDB, shardIDs []int) map[int][]*CommitteeState {
	return stateDB.getShardsCommitteeState(shardIDs)
}

func GetAllCommitteeStakeInfo(stateDB *StateDB, allShardCommittee map[int][]*CommitteeState) map[int][]*StakerInfo {
	return stateDB.getShardsCommitteeInfo(allShardCommittee)
}

func GetAllCommitteeStakeInfoSlashingVersion(stateDB *StateDB, allShardCommittee map[int][]*CommitteeState) map[int][]*StakerInfoSlashingVersion {
	return stateDB.getShardsCommitteeInfoV2(allShardCommittee)
}

func GetStakingInfo(bcDB *StateDB, shardIDs []int) map[string]bool {
	mapAutoStaking, err := bcDB.getMapAutoStaking(shardIDs)
	if err != nil {
		panic(err)
	}
	return mapAutoStaking
}

func GetStakerInfo(stateDB *StateDB, stakerPubkey string) (*StakerInfo, bool, error) {
	pubKey := incognitokey.NewCommitteePublicKey()
	err := pubKey.FromString(stakerPubkey)
	if err != nil {
		return nil, false, err
	}
	pubKeyBytes, _ := pubKey.RawBytes()
	key := GetStakerInfoKey(pubKeyBytes)
	return stateDB.getStakerInfo(key)
}

func deleteCommittee(stateDB *StateDB, shardID int, role int, committees []incognitokey.CommitteePublicKey) error {
	for _, committee := range committees {
		key, err := GenerateCommitteeObjectKeyWithRole(role, shardID, committee)
		if err != nil {
			return err
		}
		stateDB.MarkDeleteStateObject(CommitteeObjectType, key)
	}
	return nil
}

func DeleteOneShardCommittee(stateDB *StateDB, shardID byte, shardCommittees []incognitokey.CommitteePublicKey) error {
	err := deleteCommittee(stateDB, int(shardID), CurrentValidator, shardCommittees)
	if err != nil {
		return NewStatedbError(DeleteOneShardCommitteeError, err)
	}
	return nil
}

func DeleteAllShardCommittee(stateDB *StateDB, allShardCommittees map[byte][]incognitokey.CommitteePublicKey) error {
	for shardID, committee := range allShardCommittees {
		err := deleteCommittee(stateDB, int(shardID), CurrentValidator, committee)
		if err != nil {
			return NewStatedbError(DeleteAllShardCommitteeError, err)
		}
	}
	return nil
}

func DeleteNextEpochShardCandidate(stateDB *StateDB, candidate []incognitokey.CommitteePublicKey) error {
	err := deleteCommittee(stateDB, CandidateChainID, NextEpochShardCandidate, candidate)
	if err != nil {
		return NewStatedbError(DeleteNextEpochShardCandidateError, err)
	}
	return nil
}

func DeleteCurrentEpochShardCandidate(stateDB *StateDB, candidate []incognitokey.CommitteePublicKey) error {
	err := deleteCommittee(stateDB, CandidateChainID, CurrentEpochShardCandidate, candidate)
	if err != nil {
		return NewStatedbError(DeleteCurrentEpochShardCandidateError, err)
	}
	return nil
}

func DeleteNextEpochBeaconCandidate(stateDB *StateDB, candidate []incognitokey.CommitteePublicKey) error {
	err := deleteCommittee(stateDB, BeaconChainID, NextEpochBeaconCandidate, candidate)
	if err != nil {
		return NewStatedbError(DeleteNextEpochBeaconCandidateError, err)
	}
	return nil
}

func DeleteCurrentEpochBeaconCandidate(stateDB *StateDB, candidate []incognitokey.CommitteePublicKey) error {
	err := deleteCommittee(stateDB, BeaconChainID, CurrentEpochBeaconCandidate, candidate)
	if err != nil {
		return NewStatedbError(DeleteCurrentEpochBeaconCandidateError, err)
	}
	return nil
}

func DeleteOneShardSubstitutesValidator(stateDB *StateDB, shardID byte, shardSubstitutes []incognitokey.CommitteePublicKey) error {
	err := deleteCommittee(stateDB, int(shardID), SubstituteValidator, shardSubstitutes)
	if err != nil {
		return NewStatedbError(DeleteAllShardSubstitutesValidatorError, err)
	}
	return nil
}

func storeStakerInfo(
	stateDB *StateDB,
	committees []incognitokey.CommitteePublicKey,
	rewardReceiver map[string]key.PaymentAddress,
	autoStaking map[string]bool,
	stakingTx map[string]common.Hash,
	beaconConfirmHeight uint64,
) error {
	for _, committee := range committees {
		keyBytes, err := committee.RawBytes()
		if err != nil {
			return err
		}
		key := GetStakerInfoKey(keyBytes)
		committeeString, err := committee.ToBase58()
		value := NewStakerInfoWithValue(
			rewardReceiver[committee.GetIncKeyBase58()],
			autoStaking[committeeString],
			stakingTx[committeeString],
			beaconConfirmHeight,
		)
		err = stateDB.SetStateObject(ShardStakerObjectType, key, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func SaveStopAutoStakerInfo(
	stateDB *StateDB,
	committees []incognitokey.CommitteePublicKey,
	autoStaking map[string]bool,
) error {
	for _, stakerPubkey := range committees {
		pubKeyBytes, _ := stakerPubkey.RawBytes()
		key := GetStakerInfoKey(pubKeyBytes)
		stakerInfo, exist, err := stateDB.getStakerInfo(key)
		if err != nil || !exist {
			return NewStatedbError(SaveStopAutoStakerInfoError, err)
		}

		//assign auto stake value from autostaking map
		committeeString, err := stakerPubkey.ToBase58()
		stakerInfo.autoStaking = autoStaking[committeeString]

		err = stateDB.SetStateObject(ShardStakerObjectType, key, stakerInfo)
		if err != nil {
			return err
		}
	}

	return nil
}

func StoreStakerInfo(
	stateDB *StateDB,
	committees []incognitokey.CommitteePublicKey,
	rewardReceiver map[string]key.PaymentAddress,
	autoStaking map[string]bool,
	stakingTx map[string]common.Hash,
	beaconConfirmHeight uint64,

) error {

	return storeStakerInfo(stateDB, committees, rewardReceiver, autoStaking, stakingTx, beaconConfirmHeight)
}

func StoreStakerInfoV2(
	stateDB *StateDB,
	committee incognitokey.CommitteePublicKey,
	value *StakerInfo,
) error {
	keyBytes, err := committee.RawBytes()
	if err != nil {
		return err
	}
	key := GetStakerInfoKey(keyBytes)
	err = stateDB.SetStateObject(ShardStakerObjectType, key, value)
	if err != nil {
		return err
	}
	return nil
}

func GetOneShardCommitteeEnterTime(
	stateDB *StateDB,
	shardID byte,
) []int64 {
	tempShardCommitteeStates := stateDB.getByShardIDCurrentValidatorState(int(shardID))
	sort.Slice(tempShardCommitteeStates, func(i, j int) bool {
		return tempShardCommitteeStates[i].EnterTime() < tempShardCommitteeStates[j].EnterTime()
	})
	list := []int64{}
	for _, tempShardCommitteeState := range tempShardCommitteeStates {
		list = append(list, tempShardCommitteeState.EnterTime())
	}
	return list
}

func GetAllStaker(stateDB *StateDB, shardIDs []int) int {
	return stateDB.getAllStaker(shardIDs)
}

// DeleteStakerInfo :
func DeleteStakerInfo(stateDB *StateDB, stakers []incognitokey.CommitteePublicKey) error {
	return deleteStakerInfo(stateDB, stakers)
}

func deleteStakerInfo(stateDB *StateDB, stakers []incognitokey.CommitteePublicKey) error {
	for _, staker := range stakers {
		keyBytes, err := staker.RawBytes()
		if err != nil {
			return err
		}
		key := GetStakerInfoKey(keyBytes)
		stateDB.MarkDeleteStateObject(ShardStakerObjectType, key)
	}
	return nil
}

func StoreSlashingCommittee(stateDB *StateDB, epoch uint64, slashingCommittees map[byte][]string) error {
	for shardID, committees := range slashingCommittees {
		key := GenerateSlashingCommitteeObjectKey(shardID, epoch)
		value := NewSlashingCommitteeStateWithValue(shardID, epoch, committees)
		err := stateDB.SetStateObject(SlashingCommitteeObjectType, key, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetSlashingCommittee(stateDB *StateDB, epoch uint64) map[byte][]string {
	return stateDB.getAllSlashingCommittee(epoch)
}

func DeleteSyncingValidators(stateDB *StateDB, syncingValidators map[byte][]incognitokey.CommitteePublicKey) error {
	for shardID, singleChainSyncingValidators := range syncingValidators {
		err := deleteSyncingValidators(stateDB, shardID, SyncingValidators, singleChainSyncingValidators)
		if err != nil {
			return NewStatedbError(DeleteAllShardCommitteeError, err)
		}
	}
	return nil
}

func deleteSyncingValidators(stateDB *StateDB, shardID byte, role int, committees []incognitokey.CommitteePublicKey) error {
	for _, committee := range committees {
		key, err := GenerateCommitteeObjectKeyWithRole(role, int(shardID), committee)
		if err != nil {
			return err
		}
		stateDB.MarkDeleteStateObject(CommitteeObjectType, key)
	}
	return nil
}

func StoreAllShardSubstitutesValidator(stateDB *StateDB, allShardSubstitutes map[byte][]incognitokey.CommitteePublicKey) error {
	for shardID, committee := range allShardSubstitutes {
		err := storeCommittee(stateDB, int(shardID), SubstituteValidator, committee, defaultEnterTime)
		if err != nil {
			return NewStatedbError(StoreNextEpochCandidateError, err)
		}
	}
	return nil
}

func DeleteAllShardSubstitutesValidator(stateDB *StateDB, allShardSubstitutes map[byte][]incognitokey.CommitteePublicKey) error {
	for shardID, committee := range allShardSubstitutes {
		err := deleteCommittee(stateDB, int(shardID), SubstituteValidator, committee)
		if err != nil {
			return NewStatedbError(DeleteAllShardSubstitutesValidatorError, err)
		}
	}
	return nil
}
