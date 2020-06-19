package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

type RewardFeatureState struct {
	totalRewards map[string]uint64
}

func (rfs RewardFeatureState) GetTotalRewards() map[string]uint64 {
	return rfs.totalRewards
}

func (rfs *RewardFeatureState) SetTotalRewards(totalRewards map[string]uint64) {
	rfs.totalRewards = totalRewards
}

func (rfs *RewardFeatureState) AddTotalRewards(tokenID string, amount uint64) {
	if rfs.totalRewards == nil {
		rfs.totalRewards = make(map[string]uint64, 0)
		rfs.totalRewards[tokenID] = amount
	} else {
		rfs.totalRewards[tokenID] += amount
	}
}

func (rfs *RewardFeatureState) ResetTotalRewardByTokenID(tokenID string) {
	rfs.totalRewards[tokenID] = 0
}

func (rfs RewardFeatureState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		TotalRewards map[string]uint64
	}{
		TotalRewards: rfs.totalRewards,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (rfs *RewardFeatureState) UnmarshalJSON(data []byte) error {
	temp := struct {
		TotalRewards map[string]uint64
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	rfs.totalRewards = temp.TotalRewards
	return nil
}

func NewRewardFeatureState() *RewardFeatureState {
	return &RewardFeatureState{}
}

func NewRewardFeatureStateWithValue(
	totalRewards map[string]uint64,
) *RewardFeatureState {
	return &RewardFeatureState{
		totalRewards: totalRewards,
	}
}

type RewardFeatureStateObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version                int
	rewardFeatureStateHash common.Hash
	rewardFeatureState     *RewardFeatureState
	objectType             int
	deleted                bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newRewardFeatureStateObject(db *StateDB, hash common.Hash) *RewardFeatureStateObject {
	return &RewardFeatureStateObject{
		version:                defaultVersion,
		db:                     db,
		rewardFeatureStateHash: hash,
		rewardFeatureState:     NewRewardFeatureState(),
		objectType:             RewardFeatureStateObjectType,
		deleted:                false,
	}
}

func newRewardFeatureStateObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*RewardFeatureStateObject, error) {
	var totalCustodianRewardState = NewRewardFeatureState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, totalCustodianRewardState)
		if err != nil {
			return nil, err
		}
	} else {
		totalCustodianRewardState, ok = data.(*RewardFeatureState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidRewardFeatureStateType, reflect.TypeOf(data))
		}
	}
	return &RewardFeatureStateObject{
		version:                defaultVersion,
		rewardFeatureStateHash: key,
		rewardFeatureState:     totalCustodianRewardState,
		db:                     db,
		objectType:             RewardFeatureStateObjectType,
		deleted:                false,
	}, nil
}

func GenerateRewardFeatureStateObjectKey(featureName string, epoch uint64) common.Hash {
	prefixHash := GetRewardFeatureStatePrefix(epoch)
	valueHash := common.HashH([]byte(featureName))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (t RewardFeatureStateObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *RewardFeatureStateObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t RewardFeatureStateObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *RewardFeatureStateObject) SetValue(data interface{}) error {
	rewardFeatureState, ok := data.(*RewardFeatureState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidRewardFeatureStateType, reflect.TypeOf(data))
	}
	t.rewardFeatureState = rewardFeatureState
	return nil
}

func (t RewardFeatureStateObject) GetValue() interface{} {
	return t.rewardFeatureState
}

func (t RewardFeatureStateObject) GetValueBytes() []byte {
	rewardFeatureState, ok := t.GetValue().(*RewardFeatureState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(rewardFeatureState)
	if err != nil {
		panic("failed to marshal reward feature state")
	}
	return value
}

func (t RewardFeatureStateObject) GetHash() common.Hash {
	return t.rewardFeatureStateHash
}

func (t RewardFeatureStateObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *RewardFeatureStateObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *RewardFeatureStateObject) Reset() bool {
	t.rewardFeatureState = NewRewardFeatureState()
	return true
}

func (t RewardFeatureStateObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t RewardFeatureStateObject) IsEmpty() bool {
	temp := NewRewardFeatureState()
	return reflect.DeepEqual(temp, t.rewardFeatureState) || t.rewardFeatureState == nil
}
