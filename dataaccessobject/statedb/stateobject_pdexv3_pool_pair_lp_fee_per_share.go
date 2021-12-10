package statedb

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type Pdexv3PoolPairLpFeePerShareState struct {
	tokenID common.Hash
	value   *big.Int
}

func (state *Pdexv3PoolPairLpFeePerShareState) Value() big.Int {
	return *state.value
}

func (state *Pdexv3PoolPairLpFeePerShareState) TokenID() common.Hash {
	return state.tokenID
}

func (state *Pdexv3PoolPairLpFeePerShareState) MarshalJSON() ([]byte, error) {
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

func (state *Pdexv3PoolPairLpFeePerShareState) UnmarshalJSON(data []byte) error {
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

func (state *Pdexv3PoolPairLpFeePerShareState) Clone() *Pdexv3PoolPairLpFeePerShareState {
	return &Pdexv3PoolPairLpFeePerShareState{
		tokenID: state.tokenID,
		value:   big.NewInt(0).Set(state.value),
	}
}

func NewPdexv3PoolPairLpFeePerShareState() *Pdexv3PoolPairLpFeePerShareState {
	return &Pdexv3PoolPairLpFeePerShareState{}
}

func NewPdexv3PoolPairLpFeePerShareStateWithValue(
	tokenID common.Hash, value *big.Int,
) *Pdexv3PoolPairLpFeePerShareState {
	return &Pdexv3PoolPairLpFeePerShareState{
		tokenID: tokenID,
		value:   value,
	}
}

type Pdexv3PoolPairLpFeePerShareObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *Pdexv3PoolPairLpFeePerShareState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPdexv3PoolPairLpFeePerShareObject(db *StateDB, hash common.Hash) *Pdexv3PoolPairLpFeePerShareObject {
	return &Pdexv3PoolPairLpFeePerShareObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewPdexv3PoolPairLpFeePerShareState(),
		objectType: Pdexv3PoolPairLpFeePerShareObjectType,
		deleted:    false,
	}
}

func newPdexv3PoolPairLpFeePerShareObjectWithValue(
	db *StateDB, key common.Hash, data interface{},
) (*Pdexv3PoolPairLpFeePerShareObject, error) {
	var newPdexv3PoolPairLpFeePerShareState = NewPdexv3PoolPairLpFeePerShareState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPdexv3PoolPairLpFeePerShareState)
		if err != nil {
			return nil, err
		}
	} else {
		newPdexv3PoolPairLpFeePerShareState, ok = data.(*Pdexv3PoolPairLpFeePerShareState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3PoolPairLpFeePerShareStateType, reflect.TypeOf(data))
		}
	}
	return &Pdexv3PoolPairLpFeePerShareObject{
		version:    defaultVersion,
		hash:       key,
		state:      newPdexv3PoolPairLpFeePerShareState,
		db:         db,
		objectType: Pdexv3PoolPairLpFeePerShareObjectType,
		deleted:    false,
	}, nil
}

func generatePdexv3PoolPairLpFeePerShareObjectPrefix(poolPairID string) []byte {
	b := append(GetPdexv3PoolPairLpFeePerSharesPrefix(), []byte(poolPairID)...)
	h := common.HashH(b)
	return h[:prefixHashKeyLength]
}

func GeneratePdexv3PoolPairLpFeePerShareObjectKey(poolPairID, tokenID string) common.Hash {
	prefixHash := generatePdexv3PoolPairLpFeePerShareObjectPrefix(poolPairID)
	valueHash := common.HashH([]byte(tokenID))
	return common.BytesToHash(append(prefixHash, valueHash[:prefixKeyLength]...))
}

func (object *Pdexv3PoolPairLpFeePerShareObject) GetVersion() int {
	return object.version
}

// setError remembers the first non-nil error it is called with.
func (object *Pdexv3PoolPairLpFeePerShareObject) SetError(err error) {
	if object.dbErr == nil {
		object.dbErr = err
	}
}

func (object *Pdexv3PoolPairLpFeePerShareObject) GetTrie(db DatabaseAccessWarper) Trie {
	return object.trie
}

func (object *Pdexv3PoolPairLpFeePerShareObject) SetValue(data interface{}) error {
	newPdexv3PoolPairLpFeePerShareState, ok := data.(*Pdexv3PoolPairLpFeePerShareState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3PoolPairLpFeePerShareStateType, reflect.TypeOf(data))
	}
	object.state = newPdexv3PoolPairLpFeePerShareState
	return nil
}

func (object *Pdexv3PoolPairLpFeePerShareObject) GetValue() interface{} {
	return object.state
}

func (object *Pdexv3PoolPairLpFeePerShareObject) GetValueBytes() []byte {
	state, ok := object.GetValue().(*Pdexv3PoolPairLpFeePerShareState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal pdexv3 pool pair lp fee per share state")
	}
	return value
}

func (object *Pdexv3PoolPairLpFeePerShareObject) GetHash() common.Hash {
	return object.hash
}

func (object *Pdexv3PoolPairLpFeePerShareObject) GetType() int {
	return object.objectType
}

// MarkDelete will delete an object in trie
func (object *Pdexv3PoolPairLpFeePerShareObject) MarkDelete() {
	object.deleted = true
}

// reset all shard committee value into default value
func (object *Pdexv3PoolPairLpFeePerShareObject) Reset() bool {
	object.state = NewPdexv3PoolPairLpFeePerShareState()
	return true
}

func (object *Pdexv3PoolPairLpFeePerShareObject) IsDeleted() bool {
	return object.deleted
}

// value is either default or nil
func (object *Pdexv3PoolPairLpFeePerShareObject) IsEmpty() bool {
	temp := NewPdexv3PoolPairLpFeePerShareState()
	return reflect.DeepEqual(temp, object.state) || object.state == nil
}
