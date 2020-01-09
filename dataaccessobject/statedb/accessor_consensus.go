package statedb

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"time"
)

func storeCommittee(stateDB *StateDB, shardID int, role int, committees []incognitokey.CommitteePublicKey, rewardReceiver map[string]string, autoStaking map[string]bool) error {
	enterTime := time.Now().UnixNano()
	for _, committee := range committees {
		key, err := GenerateCommitteeObjectKeyWithRole(role, shardID, committee)
		if err != nil {
			return err
		}
		incPublicKey := incognitokey.CommitteeKeyListToMapString([]incognitokey.CommitteePublicKey{committee})
		temp, err := incognitokey.CommitteeKeyListToString([]incognitokey.CommitteePublicKey{committee})
		if err != nil {
			return err
		}
		committeeString := temp[0]
		//TODO: Change to committee public key string in the future, now is incognito public key string
		rewardReceiverPaymentAddress, ok := rewardReceiver[incPublicKey[0].IncPubKey]
		if !ok {
			return fmt.Errorf("reward receiver of %+v not found", committeeString)
		}
		autoStakingValue, ok := autoStaking[committeeString]
		if !ok {
			return fmt.Errorf("auto staking of %+v not found", committeeString)
		}
		value := NewCommitteeStateWithValueAndTime(shardID, role, committee, rewardReceiverPaymentAddress, autoStakingValue, enterTime)
		err = stateDB.SetStateObject(CommitteeObjectType, key, value)
		if err != nil {
			return err
		}
		enterTime++
	}
	return nil
}

func StoreBeaconCommittee(stateDB *StateDB, beaconCommittees []incognitokey.CommitteePublicKey, rewardReceiver map[string]string, autoStaking map[string]bool) error {
	err := storeCommittee(stateDB, BeaconShardID, CurrentValidator, beaconCommittees, rewardReceiver, autoStaking)
	if err != nil {
		return NewStatedbError(StoreBeaconCommitteeError, err)
	}
	return nil
}

func StoreOneShardCommittee(stateDB *StateDB, shardID byte, shardCommittees []incognitokey.CommitteePublicKey, rewardReceiver map[string]string, autoStaking map[string]bool) error {
	err := storeCommittee(stateDB, int(shardID), CurrentValidator, shardCommittees, rewardReceiver, autoStaking)
	if err != nil {
		return NewStatedbError(StoreShardCommitteeError, err)
	}
	return nil
}
func StoreAllShardCommittee(stateDB *StateDB, allShardCommittees map[byte][]incognitokey.CommitteePublicKey, rewardReceiver map[string]string, autoStaking map[string]bool) error {
	for shardID, committee := range allShardCommittees {
		err := storeCommittee(stateDB, int(shardID), CurrentValidator, committee, rewardReceiver, autoStaking)
		if err != nil {
			return NewStatedbError(StoreAllShardCommitteeError, err)
		}
	}
	return nil
}

func StoreNextEpochShardCandidate(stateDB *StateDB, candidate []incognitokey.CommitteePublicKey, rewardReceiver map[string]string, autoStaking map[string]bool) error {
	err := storeCommittee(stateDB, CandidateShardID, NextEpochShardCandidate, candidate, rewardReceiver, autoStaking)
	if err != nil {
		return NewStatedbError(StoreNextEpochCandidateError, err)
	}
	return nil
}

func StoreCurrentEpochShardCandidate(stateDB *StateDB, candidate []incognitokey.CommitteePublicKey, rewardReceiver map[string]string, autoStaking map[string]bool) error {
	err := storeCommittee(stateDB, CandidateShardID, CurrentEpochShardCandidate, candidate, rewardReceiver, autoStaking)
	if err != nil {
		return NewStatedbError(StoreCurrentEpochCandidateError, err)
	}
	return nil
}

func StoreNextEpochBeaconCandidate(stateDB *StateDB, candidate []incognitokey.CommitteePublicKey, rewardReceiver map[string]string, autoStaking map[string]bool) error {
	err := storeCommittee(stateDB, BeaconShardID, NextEpochBeaconCandidate, candidate, rewardReceiver, autoStaking)
	if err != nil {
		return NewStatedbError(StoreNextEpochCandidateError, err)
	}
	return nil
}

func StoreCurrentEpochBeaconCandidate(stateDB *StateDB, candidate []incognitokey.CommitteePublicKey, rewardReceiver map[string]string, autoStaking map[string]bool) error {
	err := storeCommittee(stateDB, BeaconShardID, CurrentEpochBeaconCandidate, candidate, rewardReceiver, autoStaking)
	if err != nil {
		return NewStatedbError(StoreCurrentEpochCandidateError, err)
	}
	return nil
}

func StoreAllShardSubstitutesValidator(stateDB *StateDB, allShardSubstitutes map[byte][]incognitokey.CommitteePublicKey, rewardReceiver map[string]string, autoStaking map[string]bool) error {
	for shardID, committee := range allShardSubstitutes {
		err := storeCommittee(stateDB, int(shardID), SubstituteValidator, committee, rewardReceiver, autoStaking)
		if err != nil {
			return NewStatedbError(StoreNextEpochCandidateError, err)
		}
	}
	return nil
}

func StoreOneShardSubstitutesValidator(stateDB *StateDB, shardID byte, shardSubstitutes []incognitokey.CommitteePublicKey, rewardReceiver map[string]string, autoStaking map[string]bool) error {
	err := storeCommittee(stateDB, int(shardID), SubstituteValidator, shardSubstitutes, rewardReceiver, autoStaking)
	if err != nil {
		return NewStatedbError(StoreNextEpochCandidateError, err)
	}
	return nil
}

func StoreBeaconSubstituteValidator(stateDB *StateDB, beaconSubstitute []incognitokey.CommitteePublicKey, rewardReceiver map[string]string, autoStaking map[string]bool) error {
	err := storeCommittee(stateDB, BeaconShardID, SubstituteValidator, beaconSubstitute, rewardReceiver, autoStaking)
	if err != nil {
		return NewStatedbError(StoreCurrentEpochCandidateError, err)
	}
	return nil
}

func GetBeaconCommittee(stateDB *StateDB) []incognitokey.CommitteePublicKey {
	m := stateDB.GetAllValidatorCommitteePublicKey(CurrentValidator, []int{BeaconShardID})
	return m[BeaconShardID]
}

func GetBeaconSubstituteValidator(stateDB *StateDB) []incognitokey.CommitteePublicKey {
	m := stateDB.GetAllValidatorCommitteePublicKey(SubstituteValidator, []int{BeaconShardID})
	return m[BeaconShardID]
}

func GetOneShardCommittee(stateDB *StateDB, shardID byte) []incognitokey.CommitteePublicKey {
	return stateDB.GetByShardIDCurrentValidatorState(int(shardID))
}

func GetOneShardSubstituteValidator(stateDB *StateDB, shardID byte) []incognitokey.CommitteePublicKey {
	return stateDB.GetByShardIDSubstituteValidatorState(int(shardID))
}

func GetAllShardCommittee(stateDB *StateDB, shardIDs []int) map[int][]incognitokey.CommitteePublicKey {
	m := stateDB.GetAllValidatorCommitteePublicKey(CurrentValidator, shardIDs)
	return m
}

func GetAllShardSubstituteValidator(stateDB *StateDB, shardIDs []int) map[int][]incognitokey.CommitteePublicKey {
	m := stateDB.GetAllValidatorCommitteePublicKey(SubstituteValidator, shardIDs)
	return m
}

func GetNextEpochCandidate(stateDB *StateDB) []incognitokey.CommitteePublicKey {
	return stateDB.GetAllCandidateCommitteePublicKey(NextEpochShardCandidate)
}

func GetCurrentEpochCandidate(stateDB *StateDB) []incognitokey.CommitteePublicKey {
	return stateDB.GetAllCandidateCommitteePublicKey(CurrentEpochShardCandidate)
}

func GetAllCommitteeSubstituteCandidateState(stateDB *StateDB, shardIDs []int) (map[int][]incognitokey.CommitteePublicKey, map[int][]incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, map[string]string, map[string]bool) {
	return stateDB.GetAllCommitteeState(shardIDs)
}

func deleteCommittee(stateDB *StateDB, shardID int, role int, committees []incognitokey.CommitteePublicKey) error {
	for _, committee := range committees {
		key, err := GenerateCommitteeObjectKeyWithRole(role, shardID, committee)
		if err != nil {
			return err
		}
		stateDB.MarkDeleteStateObject(CommitmentObjectType, key)
	}
	return nil
}

func DeleteBeaconCommittee(stateDB *StateDB, beaconCommittees []incognitokey.CommitteePublicKey) error {
	err := deleteCommittee(stateDB, BeaconShardID, CurrentValidator, beaconCommittees)
	if err != nil {
		return NewStatedbError(DeleteBeaconCommitteeError, err)
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
	err := deleteCommittee(stateDB, CandidateShardID, NextEpochShardCandidate, candidate)
	if err != nil {
		return NewStatedbError(DeleteNextEpochShardCandidateError, err)
	}
	return nil
}

func DeleteCurrentEpochShardCandidate(stateDB *StateDB, candidate []incognitokey.CommitteePublicKey) error {
	err := deleteCommittee(stateDB, CandidateShardID, CurrentEpochShardCandidate, candidate)
	if err != nil {
		return NewStatedbError(DeleteCurrentEpochShardCandidateError, err)
	}
	return nil
}

func DeleteNextEpochBeaconCandidate(stateDB *StateDB, candidate []incognitokey.CommitteePublicKey) error {
	err := deleteCommittee(stateDB, BeaconShardID, NextEpochBeaconCandidate, candidate)
	if err != nil {
		return NewStatedbError(DeleteNextEpochBeaconCandidateError, err)
	}
	return nil
}

func DeleteCurrentEpochBeaconCandidate(stateDB *StateDB, candidate []incognitokey.CommitteePublicKey) error {
	err := deleteCommittee(stateDB, BeaconShardID, CurrentEpochBeaconCandidate, candidate)
	if err != nil {
		return NewStatedbError(DeleteCurrentEpochBeaconCandidateError, err)
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

func DeleteOneShardSubstitutesValidator(stateDB *StateDB, shardID byte, shardSubstitutes []incognitokey.CommitteePublicKey) error {
	err := deleteCommittee(stateDB, int(shardID), SubstituteValidator, shardSubstitutes)
	if err != nil {
		return NewStatedbError(DeleteAllShardSubstitutesValidatorError, err)
	}
	return nil
}

func DeleteBeaconSubstituteValidator(stateDB *StateDB, beaconSubstitute []incognitokey.CommitteePublicKey) error {
	err := deleteCommittee(stateDB, BeaconShardID, SubstituteValidator, beaconSubstitute)
	if err != nil {
		return NewStatedbError(DeleteBeaconSubstituteValidatorError, err)
	}
	return nil
}
