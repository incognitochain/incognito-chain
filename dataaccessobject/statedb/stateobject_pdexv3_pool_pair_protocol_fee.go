package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type Pdexv3PoolPairProtocolFeeState struct {
	tokenID common.Hash
	value   uint64
}

func (state *Pdexv3PoolPairProtocolFeeState) Value() uint64 {
	return state.value
}

func (state *Pdexv3PoolPairProtocolFeeState) TokenID() common.Hash {
	return state.tokenID
}

func (state *Pdexv3PoolPairProtocolFeeState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		TokenID common.Hash `json:"TokenID"`
		Value   uint64      `json:"Value"`
	}{
		TokenID: state.tokenID,
		Value:   state.value,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (state *Pdexv3PoolPairProtocolFeeState) UnmarshalJSON(data []byte) error {
	temp := struct {
		TokenID common.Hash `json:"TokenID"`
		Value   uint64      `json:"Value"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	state.tokenID = temp.TokenID
	state.value = temp.Value
	return nil
}

func (state *Pdexv3PoolPairProtocolFeeState) Clone() *Pdexv3PoolPairProtocolFeeState {
	return &Pdexv3PoolPairProtocolFeeState{
		tokenID: state.tokenID,
		value:   state.value,
	}
}

func NewPdexv3PoolPairProtocolFeeState() *Pdexv3PoolPairProtocolFeeState {
	return &Pdexv3PoolPairProtocolFeeState{}
}

func NewPdexv3PoolPairProtocolFeeStateWithValue(
	tokenID common.Hash, value uint64,
) *Pdexv3PoolPairProtocolFeeState {
	return &Pdexv3PoolPairProtocolFeeState{
		tokenID: tokenID,
		value:   value,
	}
}

type Pdexv3PoolPairProtocolFeeObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *Pdexv3PoolPairProtocolFeeState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPdexv3PoolPairProtocolFeeObject(db *StateDB, hash common.Hash) *Pdexv3PoolPairProtocolFeeObject {
	return &Pdexv3PoolPairProtocolFeeObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewPdexv3PoolPairProtocolFeeState(),
		objectType: Pdexv3PoolPairProtocolFeeObjectType,
		deleted:    false,
	}
}

func newPdexv3PoolPairProtocolFeeObjectWithValue(
	db *StateDB, key common.Hash, data interface{},
) (*Pdexv3PoolPairProtocolFeeObject, error) {
	var newPdexv3PoolPairProtocolFeeState = NewPdexv3PoolPairProtocolFeeState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPdexv3PoolPairProtocolFeeState)
		if err != nil {
			return nil, err
		}
	} else {
		newPdexv3PoolPairProtocolFeeState, ok = data.(*Pdexv3PoolPairProtocolFeeState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3PoolPairProtocolFeeStateType, reflect.TypeOf(data))
		}
	}
	return &Pdexv3PoolPairProtocolFeeObject{
		version:    defaultVersion,
		hash:       key,
		state:      newPdexv3PoolPairProtocolFeeState,
		db:         db,
		objectType: Pdexv3PoolPairProtocolFeeObjectType,
		deleted:    false,
	}, nil
}

func generatePdexv3PoolPairProtocolFeeObjectPrefix(poolPairID string) []byte {
	b := append(GetPdexv3PoolPairProtocolFeesPrefix(), []byte(poolPairID)...)
	h := common.HashH(b)
	return h[:prefixHashKeyLength]
}

func GeneratePdexv3PoolPairProtocolFeeObjectKey(poolPairID, tokenID string) common.Hash {
	prefixHash := generatePdexv3PoolPairProtocolFeeObjectPrefix(poolPairID)
	valueHash := common.HashH([]byte(tokenID))
	return common.BytesToHash(append(prefixHash, valueHash[:prefixKeyLength]...))
}

func (object *Pdexv3PoolPairProtocolFeeObject) GetVersion() int {
	return object.version
}

// setError remembers the first non-nil error it is called with.
func (object *Pdexv3PoolPairProtocolFeeObject) SetError(err error) {
	if object.dbErr == nil {
		object.dbErr = err
	}
}

func (object *Pdexv3PoolPairProtocolFeeObject) GetTrie(db DatabaseAccessWarper) Trie {
	return object.trie
}

func (object *Pdexv3PoolPairProtocolFeeObject) SetValue(data interface{}) error {
	newPdexv3PoolPairProtocolFeeState, ok := data.(*Pdexv3PoolPairProtocolFeeState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3PoolPairProtocolFeeStateType, reflect.TypeOf(data))
	}
	object.state = newPdexv3PoolPairProtocolFeeState
	return nil
}

func (object *Pdexv3PoolPairProtocolFeeObject) GetValue() interface{} {
	return object.state
}

func (object *Pdexv3PoolPairProtocolFeeObject) GetValueBytes() []byte {
	state, ok := object.GetValue().(*Pdexv3PoolPairProtocolFeeState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal pdexv3 pool pair protocol fee state")
	}
	return value
}

func (object *Pdexv3PoolPairProtocolFeeObject) GetHash() common.Hash {
	return object.hash
}

func (object *Pdexv3PoolPairProtocolFeeObject) GetType() int {
	return object.objectType
}

// MarkDelete will delete an object in trie
func (object *Pdexv3PoolPairProtocolFeeObject) MarkDelete() {
	object.deleted = true
}

// reset all shard committee value into default value
func (object *Pdexv3PoolPairProtocolFeeObject) Reset() bool {
	object.state = NewPdexv3PoolPairProtocolFeeState()
	return true
}

func (object *Pdexv3PoolPairProtocolFeeObject) IsDeleted() bool {
	return object.deleted
}

// value is either default or nil
func (object *Pdexv3PoolPairProtocolFeeObject) IsEmpty() bool {
	temp := NewPdexv3PoolPairProtocolFeeState()
	return reflect.DeepEqual(temp, object.state) || object.state == nil
}
