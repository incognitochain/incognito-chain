package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BridgeAggParamState struct {
	percentFeeWithDec uint64
}

func (b BridgeAggParamState) PercentFeeWithDec() uint64 {
	return b.percentFeeWithDec
}

func (b *BridgeAggParamState) SetPercentFeeWithDec(percentFeeWithDec uint64) {
	b.percentFeeWithDec = percentFeeWithDec
}

func (b BridgeAggParamState) Clone() *BridgeAggParamState {
	return &BridgeAggParamState{
		percentFeeWithDec: b.percentFeeWithDec,
	}
}

func (b *BridgeAggParamState) IsDiff(compareParam *BridgeAggParamState) bool {
	if compareParam == nil {
		return true
	}
	return b.percentFeeWithDec != compareParam.percentFeeWithDec
}

func (b BridgeAggParamState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		PercentFeeWithDec uint64
	}{
		PercentFeeWithDec: b.percentFeeWithDec,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (b *BridgeAggParamState) UnmarshalJSON(data []byte) error {
	temp := struct {
		PercentFeeWithDec uint64
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	b.percentFeeWithDec = temp.PercentFeeWithDec
	return nil
}

func NewBridgeAggParamState() *BridgeAggParamState {
	return &BridgeAggParamState{}
}

func NewBridgeAggParamStateWithValue(percentFeeWithDec uint64) *BridgeAggParamState {
	return &BridgeAggParamState{percentFeeWithDec: percentFeeWithDec}
}

type BridgeAggParamObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *BridgeAggParamState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBridgeAggParamObject(db *StateDB, hash common.Hash) *BridgeAggParamObject {
	return &BridgeAggParamObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewBridgeAggParamState(),
		objectType: BridgeAggParamObjectType,
		deleted:    false,
	}
}

func newBridgeAggParamObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BridgeAggParamObject, error) {
	var newBridgeAggParam = NewBridgeAggParamState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeAggParam)
		if err != nil {
			return nil, err
		}
	} else {
		newBridgeAggParam, ok = data.(*BridgeAggParamState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeAggParamStateType, reflect.TypeOf(data))
		}
	}
	return &BridgeAggParamObject{
		version:    defaultVersion,
		hash:       key,
		state:      newBridgeAggParam,
		db:         db,
		objectType: BridgeAggParamObjectType,
		deleted:    false,
	}, nil
}

func GenerateBridgeAggParamObjectKey() common.Hash {
	prefixHash := GetBridgeAggParamPrefix()
	// valueHash := common.HashH()
	return common.BytesToHash(prefixHash)
}

func (t BridgeAggParamObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *BridgeAggParamObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t BridgeAggParamObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *BridgeAggParamObject) SetValue(data interface{}) error {
	newBridgeAggStatus, ok := data.(*BridgeAggParamState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeAggParamStateType, reflect.TypeOf(data))
	}
	t.state = newBridgeAggStatus
	return nil
}

func (t BridgeAggParamObject) GetValue() interface{} {
	return t.state
}

func (t BridgeAggParamObject) GetValueBytes() []byte {
	bridgeAggState, ok := t.GetValue().(*BridgeAggParamState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(bridgeAggState)
	if err != nil {
		panic("failed to marshal BridgeAggParamState")
	}
	return value
}

func (t BridgeAggParamObject) GetHash() common.Hash {
	return t.hash
}

func (t BridgeAggParamObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *BridgeAggParamObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *BridgeAggParamObject) Reset() bool {
	t.state = NewBridgeAggParamState()
	return true
}

func (t BridgeAggParamObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t BridgeAggParamObject) IsEmpty() bool {
	temp := NewBridgeAggParamState()
	return reflect.DeepEqual(temp, t.state) || t.state == nil
}
