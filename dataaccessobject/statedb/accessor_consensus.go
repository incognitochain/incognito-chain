package statedb

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

func storeCommittee(stateDB *StateDB, role int, shardID int, committees []incognitokey.CommitteePublicKey, rewardReceiver map[string]string, autoStaking map[string]bool) error {
	for _, committee := range committees {
		key, err := GenerateCommitteeObjectKeyWithRole(role, shardID, committee)
		if err != nil {
			return err
		}
		temp, err := incognitokey.CommitteeKeyListToString([]incognitokey.CommitteePublicKey{committee})
		if err != nil {
			return err
		}
		committeeString := temp[0]
		rewardReceiverPaymentAddress, ok := rewardReceiver[committeeString]
		if !ok {
			return fmt.Errorf("reward receiver of %+v not found", committeeString)
		}
		autoStakingValue, ok := autoStaking[committeeString]
		if !ok {
			return fmt.Errorf("auto staking of %+v not found", committeeString)
		}
		value := NewCommitteeStateWithValue(shardID, role, committee, rewardReceiverPaymentAddress, autoStakingValue)
		err = stateDB.SetStateObject(CommitteeObjectType, key, value)
		if err != nil {
			return err
		}
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

func GetBeaconCommittee(stateDB *StateDB) []incognitokey.CommitteePublicKey {
	m := stateDB.GetAllValidatorCommitteePublicKey(CurrentValidator, []int{BeaconShardID})
	return m[BeaconShardID]
}

func GetBeaconSubstituteValidator(stateDB *StateDB) []incognitokey.CommitteePublicKey {
	m := stateDB.GetAllValidatorCommitteePublicKey(SubstituteValidator, []int{BeaconShardID})
	return m[BeaconShardID]
}

func GetOneShardCommittee(stateDB *StateDB, shardID byte) []incognitokey.CommitteePublicKey {
	m := stateDB.GetAllValidatorCommitteePublicKey(CurrentValidator, []int{int(shardID)})
	return m[int(shardID)]
}

func GetOneShardSubstituteValidator(stateDB *StateDB, shardID byte) []incognitokey.CommitteePublicKey {
	m := stateDB.GetAllValidatorCommitteePublicKey(SubstituteValidator, []int{int(shardID)})
	return m[int(shardID)]
}

func GetAllShardCommittee(stateDB *StateDB, shardIDs []int) map[int][]incognitokey.CommitteePublicKey {
	m := stateDB.GetAllValidatorCommitteePublicKey(CurrentValidator, shardIDs)
	return m
}

func GetAllShardSubstituteValidator(stateDB *StateDB, shardIDs []int) map[int][]incognitokey.CommitteePublicKey {
	m := stateDB.GetAllValidatorCommitteePublicKey(SubstituteValidator, shardIDs)
	return m
}
