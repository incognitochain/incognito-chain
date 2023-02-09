package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type ReDelegateInfo struct {
	Old    string
	New    string
	OldUID string
	NewUID string
}

type BeaconDelegateState struct {
	NextEpochDelegate map[string]ReDelegateInfo
}

func NewBeaconDelegateState() *BeaconDelegateState {
	return &BeaconDelegateState{
		NextEpochDelegate: map[string]ReDelegateInfo{},
	}
}

func NewBeaconDelegateStateWithValue(reward map[common.Hash]uint64, incognitoPublicKey string) *BeaconDelegateState {
	return &BeaconDelegateState{}
}

func (bs BeaconDelegateState) Data() map[string]ReDelegateInfo {
	return bs.NextEpochDelegate
}

func (bs *BeaconDelegateState) SetData(data map[string]ReDelegateInfo) {
	bs.NextEpochDelegate = data
}

func (bs BeaconDelegateState) AddReDelegateInfo(delegator string, redelegateInfo ReDelegateInfo) {
	bs.NextEpochDelegate[delegator] = redelegateInfo
}

func (bs BeaconDelegateState) GetReDelegateInfo(delegator string) *ReDelegateInfo {
	if info, ok := bs.NextEpochDelegate[delegator]; ok {
		return &info
	}
	return nil
}

func (c BeaconDelegateState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(c.NextEpochDelegate)
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (c *BeaconDelegateState) UnmarshalJSON(data []byte) error {

	err := json.Unmarshal(data, &c.NextEpochDelegate)
	if err != nil {
		return err
	}
	return nil
}

type BeaconReDelegateStateObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version                int
	committeePublicKeyHash common.Hash
	reDelegateState        *BeaconDelegateState
	objectType             int
	deleted                bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBeaconReDelegateStateObject(db *StateDB, hash common.Hash) *BeaconReDelegateStateObject {
	return &BeaconReDelegateStateObject{
		version:                defaultVersion,
		db:                     db,
		committeePublicKeyHash: hash,
		reDelegateState:        NewBeaconDelegateState(),
		objectType:             BeaconReDelegateStateObjectType,
		deleted:                false,
	}
}

func newBeaconReDelegateStateObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BeaconReDelegateStateObject, error) {
	var newBeaconDelegateState = NewBeaconDelegateState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBeaconDelegateState)
		if err != nil {
			return nil, err
		}
	} else {
		newBeaconDelegateState, ok = data.(*BeaconDelegateState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBeaconDelegateStateType, reflect.TypeOf(data))
		}
	}

	return &BeaconReDelegateStateObject{
		version:                defaultVersion,
		committeePublicKeyHash: key,
		reDelegateState:        newBeaconDelegateState,
		db:                     db,
		objectType:             BeaconReDelegateStateObjectType,
		deleted:                false,
	}, nil
}

func (bsObj BeaconReDelegateStateObject) GetVersion() int {
	return bsObj.version
}

// setError remembers the first non-nil error it is called with.
func (bsObj *BeaconReDelegateStateObject) SetError(err error) {
	if bsObj.dbErr == nil {
		bsObj.dbErr = err
	}
}

func (bsObj BeaconReDelegateStateObject) GetTrie(db DatabaseAccessWarper) Trie {
	return bsObj.trie
}

func (bsObj *BeaconReDelegateStateObject) SetValue(data interface{}) error {
	var newBeaconDelegateState = NewBeaconDelegateState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBeaconDelegateState)
		if err != nil {
			return err
		}
	} else {
		newBeaconDelegateState, ok = data.(*BeaconDelegateState)
		if !ok {
			return fmt.Errorf("%+v, got type %+v", ErrInvalidBeaconDelegateStateType, reflect.TypeOf(data))
		}
	}
	bsObj.reDelegateState = newBeaconDelegateState
	return nil
}

func (bsObj BeaconReDelegateStateObject) GetValue() interface{} {
	return bsObj.reDelegateState
}

func (bsObj BeaconReDelegateStateObject) GetValueBytes() []byte {
	data := bsObj.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal committee reward state")
	}
	return []byte(value)
}

func (bsObj BeaconReDelegateStateObject) GetHash() common.Hash {
	return bsObj.committeePublicKeyHash
}

func (bsObj BeaconReDelegateStateObject) GetType() int {
	return bsObj.objectType
}

// MarkDelete will delete an object in trie
func (bsObj *BeaconReDelegateStateObject) MarkDelete() {
	bsObj.deleted = true
}

func (bsObj *BeaconReDelegateStateObject) Reset() bool {
	bsObj.reDelegateState = NewBeaconDelegateState()
	return true
}

func (bsObj BeaconReDelegateStateObject) IsDeleted() bool {
	return bsObj.deleted
}

// value is either default or nil
func (bsObj BeaconReDelegateStateObject) IsEmpty() bool {
	temp := NewBeaconDelegateState()
	return reflect.DeepEqual(temp, bsObj.reDelegateState) || bsObj.reDelegateState == nil
}
