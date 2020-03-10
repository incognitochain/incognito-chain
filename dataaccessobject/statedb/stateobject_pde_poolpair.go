package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

type PDEPoolPairState struct {
	token1ID        string
	token1PoolValue uint64
	token2ID        string
	token2PoolValue uint64
}

func (pp PDEPoolPairState) Token1ID() string {
	return pp.token1ID
}

func (pp *PDEPoolPairState) SetToken1ID(token1ID string) {
	pp.token1ID = token1ID
}

func (pp PDEPoolPairState) Token1PoolValue() uint64 {
	return pp.token1PoolValue
}

func (pp *PDEPoolPairState) SetToken1PoolValue(token1PoolValue uint64) {
	pp.token1PoolValue = token1PoolValue
}

func (pp PDEPoolPairState) Token2ID() string {
	return pp.token2ID
}

func (pp *PDEPoolPairState) SetToken2ID(token2ID string) {
	pp.token2ID = token2ID
}

func (pp PDEPoolPairState) Token2PoolValue() uint64 {
	return pp.token2PoolValue
}

func (pp *PDEPoolPairState) SetToken2PoolValue(token2PoolValue uint64) {
	pp.token2PoolValue = token2PoolValue
}

func (pp PDEPoolPairState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Token1ID        string
		Token1PoolValue uint64
		Token2ID        string
		Token2PoolValue uint64
	}{
		Token1ID:        pp.token1ID,
		Token1PoolValue: pp.token1PoolValue,
		Token2ID:        pp.token2ID,
		Token2PoolValue: pp.token2PoolValue,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (pp *PDEPoolPairState) UnmarshalJSON(data []byte) error {
	temp := struct {
		Token1ID        string
		Token1PoolValue uint64
		Token2ID        string
		Token2PoolValue uint64
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	pp.token1ID = temp.Token1ID
	pp.token1PoolValue = temp.Token1PoolValue
	pp.token2ID = temp.Token2ID
	pp.token2PoolValue = temp.Token2PoolValue
	return nil
}

func NewPDEPoolPairState() *PDEPoolPairState {
	return &PDEPoolPairState{}
}

func NewPDEPoolPairStateWithValue(token1ID string, token1PoolValue uint64, token2ID string, token2PoolValue uint64) *PDEPoolPairState {
	return &PDEPoolPairState{token1ID: token1ID, token1PoolValue: token1PoolValue, token2ID: token2ID, token2PoolValue: token2PoolValue}
}

type PDEPoolPairObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version          int
	pdePoolPairHash  common.Hash
	pdePoolPairState *PDEPoolPairState
	objectType       int
	deleted          bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPDEPoolPairObject(db *StateDB, hash common.Hash) *PDEPoolPairObject {
	return &PDEPoolPairObject{
		version:          defaultVersion,
		db:               db,
		pdePoolPairHash:  hash,
		pdePoolPairState: NewPDEPoolPairState(),
		objectType:       PDEPoolPairObjectType,
		deleted:          false,
	}
}

func newPDEPoolPairObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*PDEPoolPairObject, error) {
	var newPDEPoolPairState = NewPDEPoolPairState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPDEPoolPairState)
		if err != nil {
			return nil, err
		}
	} else {
		newPDEPoolPairState, ok = data.(*PDEPoolPairState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPDEPoolPairStateType, reflect.TypeOf(data))
		}
	}
	return &PDEPoolPairObject{
		version:          defaultVersion,
		pdePoolPairHash:  key,
		pdePoolPairState: newPDEPoolPairState,
		db:               db,
		objectType:       PDEPoolPairObjectType,
		deleted:          false,
	}, nil
}

func GeneratePDEPoolPairObjectKey(token1ID, token2ID string) common.Hash {
	prefixHash := GetPDEPoolPairPrefix()
	valueHash := common.HashH([]byte(token1ID + token2ID))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (t PDEPoolPairObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *PDEPoolPairObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t PDEPoolPairObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *PDEPoolPairObject) SetValue(data interface{}) error {
	newPDEPoolPairState, ok := data.(*PDEPoolPairState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPDEPoolPairStateType, reflect.TypeOf(data))
	}
	t.pdePoolPairState = newPDEPoolPairState
	return nil
}

func (t PDEPoolPairObject) GetValue() interface{} {
	return t.pdePoolPairState
}

func (t PDEPoolPairObject) GetValueBytes() []byte {
	pdePoolPairState, ok := t.GetValue().(*PDEPoolPairState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(pdePoolPairState)
	if err != nil {
		panic("failed to marshal pde pool pair state")
	}
	return value
}

func (t PDEPoolPairObject) GetHash() common.Hash {
	return t.pdePoolPairHash
}

func (t PDEPoolPairObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *PDEPoolPairObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *PDEPoolPairObject) Reset() bool {
	t.pdePoolPairState = NewPDEPoolPairState()
	return true
}

func (t PDEPoolPairObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t PDEPoolPairObject) IsEmpty() bool {
	temp := NewPDEPoolPairState()
	return reflect.DeepEqual(temp, t.pdePoolPairState) || t.pdePoolPairState == nil
}
