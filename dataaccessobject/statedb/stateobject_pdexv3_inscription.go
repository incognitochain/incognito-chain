package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type InscriptionTokenIDState struct {
	id common.Hash
}

func (state *InscriptionTokenIDState) ID() common.Hash {
	return state.id
}

func (state *InscriptionTokenIDState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		ID common.Hash `json:"ID"`
	}{
		ID: state.id,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (state *InscriptionTokenIDState) UnmarshalJSON(data []byte) error {
	temp := struct {
		ID common.Hash `json:"ID"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	state.id = temp.ID
	return nil
}

func (state *InscriptionTokenIDState) Clone() *InscriptionTokenIDState {
	return &InscriptionTokenIDState{
		id: state.id,
	}
}

func NewInscriptionTokenIDState() *InscriptionTokenIDState {
	return &InscriptionTokenIDState{}
}

func NewInscriptionTokenIDStateWithValue(id common.Hash) *InscriptionTokenIDState {
	return &InscriptionTokenIDState{
		id: id,
	}
}

type InscriptionTokenIDObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *InscriptionTokenIDState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newInscriptionTokenIDObject(db *StateDB, hash common.Hash) *InscriptionTokenIDObject {
	return &InscriptionTokenIDObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewInscriptionTokenIDState(),
		objectType: InscriptionTokenIDObjectType,
		deleted:    false,
	}
}

func newInscriptionTokenIDObjectWithValue(db *StateDB, key common.Hash, data interface{}) (
	*InscriptionTokenIDObject, error,
) {
	var newInscriptionTokenIDState = NewInscriptionTokenIDState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newInscriptionTokenIDState)
		if err != nil {
			return nil, err
		}
	} else {
		newInscriptionTokenIDState, ok = data.(*InscriptionTokenIDState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3NftStateType, reflect.TypeOf(data))
		}
	}
	return &InscriptionTokenIDObject{
		version:    defaultVersion,
		hash:       key,
		state:      newInscriptionTokenIDState,
		db:         db,
		objectType: InscriptionTokenIDObjectType,
		deleted:    false,
	}, nil
}

func GeneratePdexv3InscriptionObjectKey(tokenID common.Hash) common.Hash {
	prefixHash := GetPdexv3InscriptionPrefix()
	valueHash := common.HashH([]byte(tokenID.String()))
	return common.BytesToHash(append(prefixHash, valueHash[:prefixKeyLength]...))
}

func (object *InscriptionTokenIDObject) GetVersion() int {
	return object.version
}

// setError remembers the first non-nil error it is called with.
func (object *InscriptionTokenIDObject) SetError(err error) {
	if object.dbErr == nil {
		object.dbErr = err
	}
}

func (object *InscriptionTokenIDObject) GetTrie(db DatabaseAccessWarper) Trie {
	return object.trie
}

func (object *InscriptionTokenIDObject) SetValue(data interface{}) error {
	newInscriptionTokenIDState, ok := data.(*InscriptionTokenIDState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3NftStateType, reflect.TypeOf(data))
	}
	object.state = newInscriptionTokenIDState
	return nil
}

func (object *InscriptionTokenIDObject) GetValue() interface{} {
	return object.state
}

func (object *InscriptionTokenIDObject) GetValueBytes() []byte {
	state, ok := object.GetValue().(*InscriptionTokenIDState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal pdexv3 nft state")
	}
	return value
}

func (object *InscriptionTokenIDObject) GetHash() common.Hash {
	return object.hash
}

func (object *InscriptionTokenIDObject) GetType() int {
	return object.objectType
}

// MarkDelete will delete an object in trie
func (object *InscriptionTokenIDObject) MarkDelete() {
	object.deleted = true
}

// reset all shard committee value into default value
func (object *InscriptionTokenIDObject) Reset() bool {
	object.state = NewInscriptionTokenIDState()
	return true
}

func (object *InscriptionTokenIDObject) IsDeleted() bool {
	return object.deleted
}

// value is either default or nil
func (object *InscriptionTokenIDObject) IsEmpty() bool {
	temp := NewInscriptionTokenIDState()
	return reflect.DeepEqual(temp, object.state) || object.state == nil
}

type InscriptionNumberState struct {
	number uint64
}

func (state *InscriptionNumberState) Number() uint64 {
	return state.number
}

func (state *InscriptionNumberState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Number uint64 `json:"number"`
	}{
		Number: state.number,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (state *InscriptionNumberState) UnmarshalJSON(data []byte) error {
	temp := struct {
		Number uint64 `json:"number"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	state.number = temp.Number
	return nil
}

func (state *InscriptionNumberState) Clone() *InscriptionNumberState {
	return &InscriptionNumberState{
		number: state.number,
	}
}

func NewInscriptionNumberState() *InscriptionNumberState {
	return &InscriptionNumberState{}
}

func NewInscriptionNumberStateWithValue(number uint64) *InscriptionNumberState {
	return &InscriptionNumberState{
		number: number,
	}
}

type InscriptionNumberObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *InscriptionNumberState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newInscriptionNumberObject(db *StateDB, hash common.Hash) *InscriptionNumberObject {
	return &InscriptionNumberObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewInscriptionNumberState(),
		objectType: InscriptionNumberObjectType,
		deleted:    false,
	}
}

func newInscriptionNumberObjectWithValue(db *StateDB, key common.Hash, data interface{}) (
	*InscriptionNumberObject, error,
) {
	var newInscriptionNumberState = NewInscriptionNumberState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newInscriptionNumberState)
		if err != nil {
			return nil, err
		}
	} else {
		newInscriptionNumberState, ok = data.(*InscriptionNumberState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3NftStateType, reflect.TypeOf(data))
		}
	}
	return &InscriptionNumberObject{
		version:    defaultVersion,
		hash:       key,
		state:      newInscriptionNumberState,
		db:         db,
		objectType: InscriptionNumberObjectType,
		deleted:    false,
	}, nil
}

func (object *InscriptionNumberObject) GetVersion() int {
	return object.version
}

// setError remembers the first non-nil error it is called with.
func (object *InscriptionNumberObject) SetError(err error) {
	if object.dbErr == nil {
		object.dbErr = err
	}
}

func (object *InscriptionNumberObject) GetTrie(db DatabaseAccessWarper) Trie {
	return object.trie
}

func (object *InscriptionNumberObject) SetValue(data interface{}) error {
	s, ok := data.(*InscriptionNumberState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3NftStateType, reflect.TypeOf(data))
	}
	object.state = s
	return nil
}

func (object *InscriptionNumberObject) GetValue() interface{} {
	return object.state
}

func (object *InscriptionNumberObject) GetValueBytes() []byte {
	state, ok := object.GetValue().(*InscriptionTokenIDState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal pdexv3 nft state")
	}
	return value
}

func (object *InscriptionNumberObject) GetHash() common.Hash {
	return object.hash
}

func (object *InscriptionNumberObject) GetType() int {
	return object.objectType
}

// MarkDelete will delete an object in trie
func (object *InscriptionNumberObject) MarkDelete() {
	object.deleted = true
}

// reset all shard committee value into default value
func (object *InscriptionNumberObject) Reset() bool {
	object.state = NewInscriptionNumberState()
	return true
}

func (object *InscriptionNumberObject) IsDeleted() bool {
	return object.deleted
}

// value is either default or nil
func (object *InscriptionNumberObject) IsEmpty() bool {
	temp := NewInscriptionTokenIDState()
	return reflect.DeepEqual(temp, object.state) || object.state == nil
}

func GeneratePdexv3InscriptionNumberObjectKey() common.Hash {
	prefixHash := GetPdexv3InscriptionNumberPrefix()
	return common.BytesToHash(prefixHash)
}
