package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

type LockedCollateralState struct {
	totalLockedCollateralForRewards uint64            // amountInUSDT
	lockedCollateralDetail          map[string]uint64 // custodianAddress : amountInUSDT
}

func (lcs LockedCollateralState) GetTotalLockedCollateralForRewards() uint64 {
	return lcs.totalLockedCollateralForRewards
}

func (lcs *LockedCollateralState) SetTotalLockedCollateralForRewards(amount uint64) {
	lcs.totalLockedCollateralForRewards = amount
}

func (lcs LockedCollateralState) GetLockedCollateralDetail() map[string]uint64 {
	if lcs.lockedCollateralDetail == nil {
		return map[string]uint64{}
	}
	return lcs.lockedCollateralDetail
}

func (lcs *LockedCollateralState) SetLockedCollateralDetail(lockedCollateralDetail map[string]uint64) {
	lcs.lockedCollateralDetail = lockedCollateralDetail
}

func (lcs *LockedCollateralState) Reset() {
	lcs.lockedCollateralDetail = nil
	lcs.totalLockedCollateralForRewards = 0
}

func (lcs LockedCollateralState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		TotalLockedCollateralForRewards uint64
		LockedCollateralDetail          map[string]uint64
	}{
		TotalLockedCollateralForRewards: lcs.totalLockedCollateralForRewards,
		LockedCollateralDetail:          lcs.lockedCollateralDetail,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (lcs *LockedCollateralState) UnmarshalJSON(data []byte) error {
	temp := struct {
		TotalLockedCollateralForRewards uint64
		LockedCollateralDetail          map[string]uint64
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	lcs.totalLockedCollateralForRewards = temp.TotalLockedCollateralForRewards
	lcs.lockedCollateralDetail = temp.LockedCollateralDetail
	return nil
}

func NewLockedCollateralState() *LockedCollateralState {
	return &LockedCollateralState{
		lockedCollateralDetail: map[string]uint64{},
	}
}

func NewLockedCollateralStateWithValue(
	totalLockedCollateralInEpoch uint64,
	lockedCollateralDetail map[string]uint64,
) *LockedCollateralState {
	return &LockedCollateralState{
		totalLockedCollateralForRewards: totalLockedCollateralInEpoch,
		lockedCollateralDetail:          lockedCollateralDetail,
	}
}

type LockedCollateralStateObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version                   int
	lockedCollateralStateHash common.Hash
	lockedCollateralState     *LockedCollateralState
	objectType                int
	deleted                   bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newLockedCollateralStateObject(db *StateDB, hash common.Hash) *LockedCollateralStateObject {
	return &LockedCollateralStateObject{
		version:                   defaultVersion,
		db:                        db,
		lockedCollateralStateHash: hash,
		lockedCollateralState:     NewLockedCollateralState(),
		objectType:                LockedCollateralStateObjectType,
		deleted:                   false,
	}
}

func newLockedCollateralStateObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*LockedCollateralStateObject, error) {
	var lockedCollateralState = NewLockedCollateralState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, lockedCollateralState)
		if err != nil {
			return nil, err
		}
	} else {
		lockedCollateralState, ok = data.(*LockedCollateralState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPortalLockedCollateralStateType, reflect.TypeOf(data))
		}
	}
	return &LockedCollateralStateObject{
		version:                   defaultVersion,
		lockedCollateralStateHash: key,
		lockedCollateralState:     lockedCollateralState,
		db:                        db,
		objectType:                LockedCollateralStateObjectType,
		deleted:                   false,
	}, nil
}

func GenerateLockedCollateralStateObjectKey() common.Hash {
	prefixHash := GetLockedCollateralStatePrefix()
	return common.BytesToHash(prefixHash)
}

func (t LockedCollateralStateObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *LockedCollateralStateObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t LockedCollateralStateObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *LockedCollateralStateObject) SetValue(data interface{}) error {
	lockedCollateralState, ok := data.(*LockedCollateralState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPortalLockedCollateralStateType, reflect.TypeOf(data))
	}
	t.lockedCollateralState = lockedCollateralState
	return nil
}

func (t LockedCollateralStateObject) GetValue() interface{} {
	return t.lockedCollateralState
}

func (t LockedCollateralStateObject) GetValueBytes() []byte {
	lockedCollateralState, ok := t.GetValue().(*LockedCollateralState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(lockedCollateralState)
	if err != nil {
		panic("failed to marshal locked collateral state")
	}
	return value
}

func (t LockedCollateralStateObject) GetHash() common.Hash {
	return t.lockedCollateralStateHash
}

func (t LockedCollateralStateObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *LockedCollateralStateObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *LockedCollateralStateObject) Reset() bool {
	t.lockedCollateralState = NewLockedCollateralState()
	return true
}

func (t LockedCollateralStateObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t LockedCollateralStateObject) IsEmpty() bool {
	temp := NewLockedCollateralState()
	return reflect.DeepEqual(temp, t.lockedCollateralState) || t.lockedCollateralState == nil
}
