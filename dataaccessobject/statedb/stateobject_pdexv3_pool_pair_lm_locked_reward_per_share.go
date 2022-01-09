package statedb

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type Pdexv3PoolPairLmLockedRewardPerShareState struct {
	tokenID common.Hash
	value   *big.Int
}

func (state *Pdexv3PoolPairLmLockedRewardPerShareState) Value() big.Int {
	return *state.value
}

func (state *Pdexv3PoolPairLmLockedRewardPerShareState) TokenID() common.Hash {
	return state.tokenID
}

func (state *Pdexv3PoolPairLmLockedRewardPerShareState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		TokenID common.Hash `json:"TokenID"`
		Value   *big.Int    `json:"Value"`
	}{
		TokenID: state.tokenID,
		Value:   state.value,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (state *Pdexv3PoolPairLmLockedRewardPerShareState) UnmarshalJSON(data []byte) error {
	temp := struct {
		TokenID common.Hash `json:"TokenID"`
		Value   *big.Int    `json:"Value"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	state.tokenID = temp.TokenID
	state.value = temp.Value
	return nil
}

func (state *Pdexv3PoolPairLmLockedRewardPerShareState) Clone() *Pdexv3PoolPairLmLockedRewardPerShareState {
	return &Pdexv3PoolPairLmLockedRewardPerShareState{
		tokenID: state.tokenID,
		value:   big.NewInt(0).Set(state.value),
	}
}

func NewPdexv3PoolPairLmLockedRewardPerShareState() *Pdexv3PoolPairLmLockedRewardPerShareState {
	return &Pdexv3PoolPairLmLockedRewardPerShareState{}
}

func NewPdexv3PoolPairLmLockedRewardPerShareStateWithValue(
	tokenID common.Hash, value *big.Int,
) *Pdexv3PoolPairLmLockedRewardPerShareState {
	return &Pdexv3PoolPairLmLockedRewardPerShareState{
		tokenID: tokenID,
		value:   value,
	}
}

type Pdexv3PoolPairLmLockedRewardPerShareObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *Pdexv3PoolPairLmLockedRewardPerShareState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPdexv3PoolPairLmLockedRewardPerShareObject(db *StateDB, hash common.Hash) *Pdexv3PoolPairLmLockedRewardPerShareObject {
	return &Pdexv3PoolPairLmLockedRewardPerShareObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewPdexv3PoolPairLmLockedRewardPerShareState(),
		objectType: Pdexv3PoolPairLmLockedRewardPerShareObjectType,
		deleted:    false,
	}
}

func newPdexv3PoolPairLmLockedRewardPerShareObjectWithValue(
	db *StateDB, key common.Hash, data interface{},
) (*Pdexv3PoolPairLmLockedRewardPerShareObject, error) {
	var newPdexv3PoolPairLmLockedRewardPerShareState = NewPdexv3PoolPairLmLockedRewardPerShareState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPdexv3PoolPairLmLockedRewardPerShareState)
		if err != nil {
			return nil, err
		}
	} else {
		newPdexv3PoolPairLmLockedRewardPerShareState, ok = data.(*Pdexv3PoolPairLmLockedRewardPerShareState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3PoolPairLmLockedRewardPerShareStateType, reflect.TypeOf(data))
		}
	}
	return &Pdexv3PoolPairLmLockedRewardPerShareObject{
		version:    defaultVersion,
		hash:       key,
		state:      newPdexv3PoolPairLmLockedRewardPerShareState,
		db:         db,
		objectType: Pdexv3PoolPairLmLockedRewardPerShareObjectType,
		deleted:    false,
	}, nil
}

func generatePdexv3PoolPairLmLockedRewardPerShareObjectPrefix(poolPairID string) []byte {
	b := append(GetPdexv3PoolPairLmLockedRewardPerSharesPrefix(), []byte(poolPairID)...)
	h := common.HashH(b)
	return h[:prefixHashKeyLength]
}

func GeneratePdexv3PoolPairLmLockedRewardPerShareObjectKey(poolPairID, tokenID string) common.Hash {
	prefixHash := generatePdexv3PoolPairLmLockedRewardPerShareObjectPrefix(poolPairID)
	valueHash := common.HashH([]byte(tokenID))
	return common.BytesToHash(append(prefixHash, valueHash[:prefixKeyLength]...))
}

func (object *Pdexv3PoolPairLmLockedRewardPerShareObject) GetVersion() int {
	return object.version
}

// setError remembers the first non-nil error it is called with.
func (object *Pdexv3PoolPairLmLockedRewardPerShareObject) SetError(err error) {
	if object.dbErr == nil {
		object.dbErr = err
	}
}

func (object *Pdexv3PoolPairLmLockedRewardPerShareObject) GetTrie(db DatabaseAccessWarper) Trie {
	return object.trie
}

func (object *Pdexv3PoolPairLmLockedRewardPerShareObject) SetValue(data interface{}) error {
	newPdexv3PoolPairLmLockedRewardPerShareState, ok := data.(*Pdexv3PoolPairLmLockedRewardPerShareState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3PoolPairLmLockedRewardPerShareStateType, reflect.TypeOf(data))
	}
	object.state = newPdexv3PoolPairLmLockedRewardPerShareState
	return nil
}

func (object *Pdexv3PoolPairLmLockedRewardPerShareObject) GetValue() interface{} {
	return object.state
}

func (object *Pdexv3PoolPairLmLockedRewardPerShareObject) GetValueBytes() []byte {
	state, ok := object.GetValue().(*Pdexv3PoolPairLmLockedRewardPerShareState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal pdexv3 pool pair lp fee per share state")
	}
	return value
}

func (object *Pdexv3PoolPairLmLockedRewardPerShareObject) GetHash() common.Hash {
	return object.hash
}

func (object *Pdexv3PoolPairLmLockedRewardPerShareObject) GetType() int {
	return object.objectType
}

// MarkDelete will delete an object in trie
func (object *Pdexv3PoolPairLmLockedRewardPerShareObject) MarkDelete() {
	object.deleted = true
}

// reset all shard committee value into default value
func (object *Pdexv3PoolPairLmLockedRewardPerShareObject) Reset() bool {
	object.state = NewPdexv3PoolPairLmLockedRewardPerShareState()
	return true
}

func (object *Pdexv3PoolPairLmLockedRewardPerShareObject) IsDeleted() bool {
	return object.deleted
}

// value is either default or nil
func (object *Pdexv3PoolPairLmLockedRewardPerShareObject) IsEmpty() bool {
	temp := NewPdexv3PoolPairLmLockedRewardPerShareState()
	return reflect.DeepEqual(temp, object.state) || object.state == nil
}
