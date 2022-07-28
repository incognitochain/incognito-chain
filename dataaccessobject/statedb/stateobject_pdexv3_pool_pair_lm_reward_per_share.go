package statedb

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type Pdexv3PoolPairLmRewardPerShareState struct {
	tokenID common.Hash
	value   *big.Int
}

func (state *Pdexv3PoolPairLmRewardPerShareState) Value() big.Int {
	return *state.value
}

func (state *Pdexv3PoolPairLmRewardPerShareState) TokenID() common.Hash {
	return state.tokenID
}

func (state *Pdexv3PoolPairLmRewardPerShareState) MarshalJSON() ([]byte, error) {
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

func (state *Pdexv3PoolPairLmRewardPerShareState) UnmarshalJSON(data []byte) error {
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

func (state *Pdexv3PoolPairLmRewardPerShareState) Clone() *Pdexv3PoolPairLmRewardPerShareState {
	return &Pdexv3PoolPairLmRewardPerShareState{
		tokenID: state.tokenID,
		value:   big.NewInt(0).Set(state.value),
	}
}

func NewPdexv3PoolPairLmRewardPerShareState() *Pdexv3PoolPairLmRewardPerShareState {
	return &Pdexv3PoolPairLmRewardPerShareState{}
}

func NewPdexv3PoolPairLmRewardPerShareStateWithValue(
	tokenID common.Hash, value *big.Int,
) *Pdexv3PoolPairLmRewardPerShareState {
	return &Pdexv3PoolPairLmRewardPerShareState{
		tokenID: tokenID,
		value:   value,
	}
}

type Pdexv3PoolPairLmRewardPerShareObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *Pdexv3PoolPairLmRewardPerShareState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPdexv3PoolPairLmRewardPerShareObject(db *StateDB, hash common.Hash) *Pdexv3PoolPairLmRewardPerShareObject {
	return &Pdexv3PoolPairLmRewardPerShareObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewPdexv3PoolPairLmRewardPerShareState(),
		objectType: Pdexv3PoolPairLmRewardPerShareObjectType,
		deleted:    false,
	}
}

func newPdexv3PoolPairLmRewardPerShareObjectWithValue(
	db *StateDB, key common.Hash, data interface{},
) (*Pdexv3PoolPairLmRewardPerShareObject, error) {
	var newPdexv3PoolPairLmRewardPerShareState = NewPdexv3PoolPairLmRewardPerShareState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPdexv3PoolPairLmRewardPerShareState)
		if err != nil {
			return nil, err
		}
	} else {
		newPdexv3PoolPairLmRewardPerShareState, ok = data.(*Pdexv3PoolPairLmRewardPerShareState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3PoolPairLmRewardPerShareStateType, reflect.TypeOf(data))
		}
	}
	return &Pdexv3PoolPairLmRewardPerShareObject{
		version:    defaultVersion,
		hash:       key,
		state:      newPdexv3PoolPairLmRewardPerShareState,
		db:         db,
		objectType: Pdexv3PoolPairLmRewardPerShareObjectType,
		deleted:    false,
	}, nil
}

func generatePdexv3PoolPairLmRewardPerShareObjectPrefix(poolPairID string) []byte {
	b := append(GetPdexv3PoolPairLmRewardPerSharesPrefix(), []byte(poolPairID)...)
	h := common.HashH(b)
	return h[:prefixHashKeyLength]
}

func GeneratePdexv3PoolPairLmRewardPerShareObjectKey(poolPairID, tokenID string) common.Hash {
	prefixHash := generatePdexv3PoolPairLmRewardPerShareObjectPrefix(poolPairID)
	valueHash := common.HashH([]byte(tokenID))
	return common.BytesToHash(append(prefixHash, valueHash[:prefixKeyLength]...))
}

func (object *Pdexv3PoolPairLmRewardPerShareObject) GetVersion() int {
	return object.version
}

// setError remembers the first non-nil error it is called with.
func (object *Pdexv3PoolPairLmRewardPerShareObject) SetError(err error) {
	if object.dbErr == nil {
		object.dbErr = err
	}
}

func (object *Pdexv3PoolPairLmRewardPerShareObject) GetTrie(db DatabaseAccessWarper) Trie {
	return object.trie
}

func (object *Pdexv3PoolPairLmRewardPerShareObject) SetValue(data interface{}) error {
	newPdexv3PoolPairLmRewardPerShareState, ok := data.(*Pdexv3PoolPairLmRewardPerShareState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3PoolPairLmRewardPerShareStateType, reflect.TypeOf(data))
	}
	object.state = newPdexv3PoolPairLmRewardPerShareState
	return nil
}

func (object *Pdexv3PoolPairLmRewardPerShareObject) GetValue() interface{} {
	return object.state
}

func (object *Pdexv3PoolPairLmRewardPerShareObject) GetValueBytes() []byte {
	state, ok := object.GetValue().(*Pdexv3PoolPairLmRewardPerShareState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal pdexv3 pool pair lp fee per share state")
	}
	return value
}

func (object *Pdexv3PoolPairLmRewardPerShareObject) GetHash() common.Hash {
	return object.hash
}

func (object *Pdexv3PoolPairLmRewardPerShareObject) GetType() int {
	return object.objectType
}

// MarkDelete will delete an object in trie
func (object *Pdexv3PoolPairLmRewardPerShareObject) MarkDelete() {
	object.deleted = true
}

// reset all shard committee value into default value
func (object *Pdexv3PoolPairLmRewardPerShareObject) Reset() bool {
	object.state = NewPdexv3PoolPairLmRewardPerShareState()
	return true
}

func (object *Pdexv3PoolPairLmRewardPerShareObject) IsDeleted() bool {
	return object.deleted
}

// value is either default or nil
func (object *Pdexv3PoolPairLmRewardPerShareObject) IsEmpty() bool {
	temp := NewPdexv3PoolPairLmRewardPerShareState()
	return reflect.DeepEqual(temp, object.state) || object.state == nil
}
