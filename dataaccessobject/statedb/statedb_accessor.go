package statedb

import (
	"encoding/json"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/trie"
)

// ================================= Serial Number OBJECT =======================================
func (stateDB *StateDB) GetSerialNumber(key common.Hash) ([]byte, error) {
	serialNumberObject, err := stateDB.getStateObject(SerialNumberObjectType, key)
	if err != nil {
		return []byte{}, err
	}
	if serialNumberObject != nil {
		return serialNumberObject.GetValueBytes(), nil
	}
	return []byte{}, nil
}

func (stateDB *StateDB) GetAllSerialNumberKeyValueList() ([]common.Hash, [][]byte) {
	temp := stateDB.trie.NodeIterator(GetSerialNumberPrefix())
	it := trie.NewIterator(temp)
	keys := []common.Hash{}
	values := [][]byte{}
	for it.Next() {
		key := stateDB.trie.GetKey(it.Key)
		newKey := make([]byte, len(key))
		copy(newKey, key)
		value := it.Value
		newValue := make([]byte, len(value))
		copy(newValue, value)
		keys = append(keys, common.BytesToHash(key))
		values = append(values, value)
	}
	return keys, values
}
func (stateDB *StateDB) GetAllSerialNumberValueList() [][]byte {
	temp := stateDB.trie.NodeIterator(GetSerialNumberPrefix())
	it := trie.NewIterator(temp)
	values := [][]byte{}
	for it.Next() {
		value := it.Value
		newValue := make([]byte, len(value))
		copy(newValue, value)
		values = append(values, value)
	}
	return values
}

// ================================= Committee OBJECT =======================================
func (stateDB *StateDB) GetCommitteeState(key common.Hash) (*CommitteeState, bool, error) {
	committeeStateObject, err := stateDB.getStateObject(CommitteeObjectType, key)
	if err != nil {
		return nil, false, err
	}
	if committeeStateObject != nil {
		return committeeStateObject.GetValue().(*CommitteeState), true, nil
	}
	return NewCommitteeState(), false, nil
}
func (stateDB *StateDB) GetAllValidatorCommitteePublicKey(role int, ids []int) map[int][]incognitokey.CommitteePublicKey {
	if role != CurrentValidator && role != SubstituteValidator {
		panic("wrong expected role " + strconv.Itoa(role))
	}
	m := make(map[int][]incognitokey.CommitteePublicKey)
	for _, id := range ids {
		prefix := GetCommitteePrefixWithRole(role, id)
		temp := stateDB.trie.NodeIterator(prefix)
		it := trie.NewIterator(temp)
		for it.Next() {
			value := it.Value
			newValue := make([]byte, len(value))
			copy(newValue, value)
			committeeState := NewCommitteeState()
			err := json.Unmarshal(newValue, committeeState)
			if err != nil {
				panic("wrong value type")
			}
			m[committeeState.shardID] = append(m[committeeState.shardID], committeeState.committeePublicKey)
		}
	}
	return m
}

func (stateDB *StateDB) GetAllCandidateCommitteePublicKey(role int) []incognitokey.CommitteePublicKey {
	if role != CurrentEpochCandidate && role != NextEpochCandidate {
		panic("wrong expected role " + strconv.Itoa(role))
	}
	list := []incognitokey.CommitteePublicKey{}
	prefix := GetCommitteePrefixWithRole(role, CandidateShardID)
	temp := stateDB.trie.NodeIterator(prefix)
	it := trie.NewIterator(temp)
	for it.Next() {
		value := it.Value
		newValue := make([]byte, len(value))
		copy(newValue, value)
		committeeState := NewCommitteeState()
		err := committeeState.UnmarshalJSON(newValue)
		if err != nil {
			panic("wrong value type")
		}
		list = append(list, committeeState.committeePublicKey)
	}
	return list
}

func (stateDB *StateDB) GetByShardIDCurrentValidatorState(shardID int) []incognitokey.CommitteePublicKey {
	committees := []incognitokey.CommitteePublicKey{}
	prefix := GetCommitteePrefixWithRole(CurrentValidator, shardID)
	temp := stateDB.trie.NodeIterator(prefix)
	it := trie.NewIterator(temp)
	for it.Next() {
		value := it.Value
		newValue := make([]byte, len(value))
		copy(newValue, value)
		committeeState := NewCommitteeState()
		err := json.Unmarshal(newValue, committeeState)
		if err != nil {
			panic("wrong value type")
		}
		if committeeState.ShardID() != shardID {
			panic("wrong expected shard id")
		}
		committees = append(committees, committeeState.CommitteePublicKey())
	}
	return committees
}

func (stateDB *StateDB) GetByShardIDSubstituteValidatorState(shardID int) []incognitokey.CommitteePublicKey {
	committees := []incognitokey.CommitteePublicKey{}
	prefix := GetCommitteePrefixWithRole(SubstituteValidator, shardID)
	temp := stateDB.trie.NodeIterator(prefix)
	it := trie.NewIterator(temp)
	for it.Next() {
		value := it.Value
		newValue := make([]byte, len(value))
		copy(newValue, value)
		committeeState := NewCommitteeState()
		err := json.Unmarshal(newValue, committeeState)
		if err != nil {
			panic("wrong value type")
		}
		if committeeState.ShardID() != shardID {
			panic("wrong expected shard id")
		}
		committees = append(committees, committeeState.CommitteePublicKey())
	}
	return committees
}

// GetAllCommitteeState return all data related to all committee roles
// return params #1: current validator
// return params #2: substitute validator
// return params #3: next epoch candidate
// return params #4: current epoch candidate
// return params #5: reward receiver map
// return params #6: auto staking map
func (stateDB *StateDB) GetAllCommitteeState(ids []int) (map[int][]incognitokey.CommitteePublicKey, map[int][]incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, map[string]string, map[string]bool) {
	currentValidator := make(map[int][]incognitokey.CommitteePublicKey)
	substituteValidator := make(map[int][]incognitokey.CommitteePublicKey)
	nextEpochCandidate := []incognitokey.CommitteePublicKey{}
	currentEpochCandidate := []incognitokey.CommitteePublicKey{}
	rewardReceivers := make(map[string]string)
	autoStaking := make(map[string]bool)
	for _, shardID := range ids {
		// Current Validator
		prefixCurrentValidator := GetCommitteePrefixWithRole(CurrentValidator, shardID)
		resCurrentValidator := stateDB.iterateWithCommitteeState(prefixCurrentValidator)
		tempCurrentValidator := []incognitokey.CommitteePublicKey{}
		for _, v := range resCurrentValidator {
			tempCurrentValidator = append(tempCurrentValidator, v.committeePublicKey)
			tempCurrentValidatorString, err := incognitokey.CommitteeKeyListToString([]incognitokey.CommitteePublicKey{v.committeePublicKey})
			if err != nil {
				panic(err)
			}
			rewardReceivers[tempCurrentValidatorString[0]] = v.rewardReceiver
			autoStaking[tempCurrentValidatorString[0]] = v.autoStaking
		}
		currentValidator[shardID] = tempCurrentValidator
		// Substitute Validator
		prefixSubstituteValidator := GetCommitteePrefixWithRole(SubstituteValidator, shardID)
		resSubstituteValidator := stateDB.iterateWithCommitteeState(prefixSubstituteValidator)
		tempSubstituteValidator := []incognitokey.CommitteePublicKey{}
		for _, v := range resSubstituteValidator {
			tempSubstituteValidator = append(tempSubstituteValidator, v.committeePublicKey)
			tempSubstituteValidatorString, err := incognitokey.CommitteeKeyListToString([]incognitokey.CommitteePublicKey{v.committeePublicKey})
			if err != nil {
				panic(err)
			}
			rewardReceivers[tempSubstituteValidatorString[0]] = v.rewardReceiver
			autoStaking[tempSubstituteValidatorString[0]] = v.autoStaking
		}
		substituteValidator[shardID] = tempSubstituteValidator
	}
	// next epoch candidate
	prefixNextEpochCandidate := GetCommitteePrefixWithRole(NextEpochCandidate, -2)
	resNextEpochCandidate := stateDB.iterateWithCommitteeState(prefixNextEpochCandidate)
	for _, v := range resNextEpochCandidate {
		nextEpochCandidate = append(nextEpochCandidate, v.committeePublicKey)
		tempNextEpochCandidateString, err := incognitokey.CommitteeKeyListToString([]incognitokey.CommitteePublicKey{v.committeePublicKey})
		if err != nil {
			panic(err)
		}
		rewardReceivers[tempNextEpochCandidateString[0]] = v.rewardReceiver
		autoStaking[tempNextEpochCandidateString[0]] = v.autoStaking
	}
	// current epoch candidate
	prefixCurrentEpochCandidate := GetCommitteePrefixWithRole(CurrentEpochCandidate, -2)
	resCurrentEpochCandidate := stateDB.iterateWithCommitteeState(prefixCurrentEpochCandidate)
	for _, v := range resCurrentEpochCandidate {
		currentEpochCandidate = append(currentEpochCandidate, v.committeePublicKey)
		tempCurrentEpochCandidateString, err := incognitokey.CommitteeKeyListToString([]incognitokey.CommitteePublicKey{v.committeePublicKey})
		if err != nil {
			panic(err)
		}
		rewardReceivers[tempCurrentEpochCandidateString[0]] = v.rewardReceiver
		autoStaking[tempCurrentEpochCandidateString[0]] = v.autoStaking
	}
	return currentValidator, substituteValidator, nextEpochCandidate, currentEpochCandidate, rewardReceivers, autoStaking
}
func (stateDB *StateDB) iterateWithCommitteeState(prefix []byte) []*CommitteeState {
	m := []*CommitteeState{}
	temp := stateDB.trie.NodeIterator(prefix)
	it := trie.NewIterator(temp)
	for it.Next() {
		value := it.Value
		newValue := make([]byte, len(value))
		copy(newValue, value)
		committeeState := NewCommitteeState()
		err := json.Unmarshal(newValue, committeeState)
		if err != nil {
			panic("wrong value type")
		}
		m = append(m, committeeState)
	}
	return m
}

// ================================= Committee Reward OBJECT =======================================
func (stateDB *StateDB) GetCommitteeRewardState(key common.Hash) (*CommitteeRewardState, bool, error) {
	rewardReceiverObject, err := stateDB.getStateObject(CommitteeRewardObjectType, key)
	if err != nil {
		return nil, false, err
	}
	if rewardReceiverObject != nil {
		return rewardReceiverObject.GetValue().(*CommitteeRewardState), true, nil
	}
	return NewCommitteeRewardState(), false, nil
}

func (stateDB *StateDB) GetCommitteeRewardAmount(key common.Hash) (map[common.Hash]int, bool, error) {
	m := make(map[common.Hash]int)
	committeeRewardObject, err := stateDB.getStateObject(CommitteeRewardObjectType, key)
	if err != nil {
		return nil, false, err
	}
	if committeeRewardObject != nil {
		temp := committeeRewardObject.GetValue().(*CommitteeRewardState)
		m = temp.reward
		return m, true, nil
	}
	return m, false, nil
}

func (stateDB *StateDB) GetAllCommitteeReward() map[string]map[common.Hash]int {
	m := make(map[string]map[common.Hash]int)
	prefix := GetCommitteeRewardPrefix()
	temp := stateDB.trie.NodeIterator(prefix)
	it := trie.NewIterator(temp)
	for it.Next() {
		value := it.Value
		newValue := make([]byte, len(value))
		copy(newValue, value)
		committeeRewardState := NewCommitteeRewardState()
		err := json.Unmarshal(newValue, committeeRewardState)
		if err != nil {
			panic("wrong value type")
		}
		m[committeeRewardState.incognitoPublicKey] = committeeRewardState.reward
	}
	return m
}
