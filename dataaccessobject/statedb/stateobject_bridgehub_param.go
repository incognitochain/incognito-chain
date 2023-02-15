package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BridgeHubParamState struct {
	minNumberValidators      uint
	minStakedAmountValidator uint64
}

func (b BridgeHubParamState) MinNumberValidators() uint {
	return b.minNumberValidators
}

func (b *BridgeHubParamState) SetMinNumberValidators(minNumberValidators uint) {
	b.minNumberValidators = minNumberValidators
}

func (b BridgeHubParamState) MinStakedAmountValidator() uint64 {
	return b.minStakedAmountValidator
}

func (b *BridgeHubParamState) SetMinStakedAmountValidator(minStakedAmountValidator uint64) {
	b.minStakedAmountValidator = minStakedAmountValidator
}

func (b BridgeHubParamState) Clone() *BridgeHubParamState {
	return &BridgeHubParamState{
		minNumberValidators:      b.minNumberValidators,
		minStakedAmountValidator: b.minStakedAmountValidator,
	}
}

func (b *BridgeHubParamState) IsDiff(compareParam *BridgeHubParamState) bool {
	if compareParam == nil {
		return true
	}
	return b.minNumberValidators != compareParam.minNumberValidators || b.minStakedAmountValidator != compareParam.minStakedAmountValidator
}

func (b BridgeHubParamState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		MinNumberValidators      uint
		MinStakedAmountValidator uint64
	}{
		MinNumberValidators:      b.minNumberValidators,
		MinStakedAmountValidator: b.minStakedAmountValidator,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (b *BridgeHubParamState) UnmarshalJSON(data []byte) error {
	temp := struct {
		MinNumberValidators      uint
		MinStakedAmountValidator uint64
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	b.minNumberValidators = temp.MinNumberValidators
	b.minStakedAmountValidator = temp.MinStakedAmountValidator
	return nil
}

func NewBridgeHubParamState() *BridgeHubParamState {
	return &BridgeHubParamState{}
}

func NewBridgeHubParamStateWithValue(minNumberValidators uint, minStakedAmountValidator uint64) *BridgeHubParamState {
	return &BridgeHubParamState{
		minNumberValidators:      minNumberValidators,
		minStakedAmountValidator: minStakedAmountValidator,
	}
}

type BridgeHubParamObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *BridgeHubParamState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBridgeHubParamObject(db *StateDB, hash common.Hash) *BridgeHubParamObject {
	return &BridgeHubParamObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewBridgeHubParamState(),
		objectType: BridgeHubParamObjectType,
		deleted:    false,
	}
}

func newBridgeHubParamObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BridgeHubParamObject, error) {
	var newBridgeHubParam = NewBridgeHubParamState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeHubParam)
		if err != nil {
			return nil, err
		}
	} else {
		newBridgeHubParam, ok = data.(*BridgeHubParamState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeHubParamStateType, reflect.TypeOf(data))
		}
	}
	return &BridgeHubParamObject{
		version:    defaultVersion,
		hash:       key,
		state:      newBridgeHubParam,
		db:         db,
		objectType: BridgeHubParamObjectType,
		deleted:    false,
	}, nil
}

func GenerateBridgeHubParamObjectKey() common.Hash {
	prefixHash := GetBridgeHubParamPrefix()
	valueHash := common.HashH([]byte{})
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (t BridgeHubParamObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *BridgeHubParamObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t BridgeHubParamObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *BridgeHubParamObject) SetValue(data interface{}) error {
	newBridgeHubParam, ok := data.(*BridgeHubParamState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeHubParamStateType, reflect.TypeOf(data))
	}
	t.state = newBridgeHubParam
	return nil
}

func (t BridgeHubParamObject) GetValue() interface{} {
	return t.state
}

func (t BridgeHubParamObject) GetValueBytes() []byte {
	bridgeHubParamState, ok := t.GetValue().(*BridgeHubParamState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(bridgeHubParamState)
	if err != nil {
		panic("failed to marshal BridgeHubParamState")
	}
	return value
}

func (t BridgeHubParamObject) GetHash() common.Hash {
	return t.hash
}

func (t BridgeHubParamObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *BridgeHubParamObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *BridgeHubParamObject) Reset() bool {
	t.state = NewBridgeHubParamState()
	return true
}

func (t BridgeHubParamObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t BridgeHubParamObject) IsEmpty() bool {
	return t.state == nil
}
