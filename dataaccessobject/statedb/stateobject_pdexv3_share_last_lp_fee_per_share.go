package statedb

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type Pdexv3ShareLastLpFeePerShareState struct {
	tokenID common.Hash
	value   *big.Int
}

func (state *Pdexv3ShareLastLpFeePerShareState) Value() big.Int {
	return *state.value
}

func (state *Pdexv3ShareLastLpFeePerShareState) TokenID() common.Hash {
	return state.tokenID
}

func (state *Pdexv3ShareLastLpFeePerShareState) MarshalJSON() ([]byte, error) {
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

func (state *Pdexv3ShareLastLpFeePerShareState) UnmarshalJSON(data []byte) error {
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

func (state *Pdexv3ShareLastLpFeePerShareState) Clone() *Pdexv3ShareLastLpFeePerShareState {
	return &Pdexv3ShareLastLpFeePerShareState{
		tokenID: state.tokenID,
		value:   state.value,
	}
}

func NewPdexv3ShareLastLpFeePerShareState() *Pdexv3ShareLastLpFeePerShareState {
	return &Pdexv3ShareLastLpFeePerShareState{}
}

func NewPdexv3ShareLastLpFeePerShareStateWithValue(
	tokenID common.Hash, value *big.Int,
) *Pdexv3ShareLastLpFeePerShareState {
	return &Pdexv3ShareLastLpFeePerShareState{
		tokenID: tokenID,
		value:   value,
	}
}

type Pdexv3ShareLastLpFeePerShareObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *Pdexv3ShareLastLpFeePerShareState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPdexv3ShareLastLpFeePerShareObject(db *StateDB, hash common.Hash) *Pdexv3ShareLastLpFeePerShareObject {
	return &Pdexv3ShareLastLpFeePerShareObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewPdexv3ShareLastLpFeePerShareState(),
		objectType: Pdexv3ShareLastLPFeesPerShareObjectType,
		deleted:    false,
	}
}

func newPdexv3ShareLastLpFeePerShareObjectWithValue(
	db *StateDB, key common.Hash, data interface{},
) (*Pdexv3ShareLastLpFeePerShareObject, error) {
	var newPdexv3ShareLastLpFeePerShareState = NewPdexv3ShareLastLpFeePerShareState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPdexv3ShareLastLpFeePerShareState)
		if err != nil {
			return nil, err
		}
	} else {
		newPdexv3ShareLastLpFeePerShareState, ok = data.(*Pdexv3ShareLastLpFeePerShareState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3LastLPFeesPerShareStateType, reflect.TypeOf(data))
		}
	}
	return &Pdexv3ShareLastLpFeePerShareObject{
		version:    defaultVersion,
		hash:       key,
		state:      newPdexv3ShareLastLpFeePerShareState,
		db:         db,
		objectType: Pdexv3ShareLastLPFeesPerShareObjectType,
		deleted:    false,
	}, nil
}

func generatePdexv3ShareLastLpFeePerShareObjectPrefix(poolPairID, nftID string) []byte {
	b := append(GetPdexv3ShareLastLpFeePerSharesPrefix(), []byte(poolPairID)...)
	b = append(b, []byte(nftID)...)
	h := common.HashH(b)
	return h[:prefixHashKeyLength]
}

func GeneratePdexv3ShareLastLpFeePerShareObjectKey(stakingPoolID, nftID, tokenID string) common.Hash {
	prefixHash := generatePdexv3ShareLastLpFeePerShareObjectPrefix(stakingPoolID, nftID)
	valueHash := common.HashH([]byte(tokenID))
	return common.BytesToHash(append(prefixHash, valueHash[:prefixKeyLength]...))
}

func (object *Pdexv3ShareLastLpFeePerShareObject) GetVersion() int {
	return object.version
}

// setError remembers the first non-nil error it is called with.
func (object *Pdexv3ShareLastLpFeePerShareObject) SetError(err error) {
	if object.dbErr == nil {
		object.dbErr = err
	}
}

func (object *Pdexv3ShareLastLpFeePerShareObject) GetTrie(db DatabaseAccessWarper) Trie {
	return object.trie
}

func (object *Pdexv3ShareLastLpFeePerShareObject) SetValue(data interface{}) error {
	newPdexv3ShareLastLpFeePerShareState, ok := data.(*Pdexv3ShareLastLpFeePerShareState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3LastLPFeesPerShareStateType, reflect.TypeOf(data))
	}
	object.state = newPdexv3ShareLastLpFeePerShareState
	return nil
}

func (object *Pdexv3ShareLastLpFeePerShareObject) GetValue() interface{} {
	return object.state
}

func (object *Pdexv3ShareLastLpFeePerShareObject) GetValueBytes() []byte {
	state, ok := object.GetValue().(*Pdexv3ShareLastLpFeePerShareState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal pdexv3 share lastLPFeesPerShare state")
	}
	return value
}

func (object *Pdexv3ShareLastLpFeePerShareObject) GetHash() common.Hash {
	return object.hash
}

func (object *Pdexv3ShareLastLpFeePerShareObject) GetType() int {
	return object.objectType
}

// MarkDelete will delete an object in trie
func (object *Pdexv3ShareLastLpFeePerShareObject) MarkDelete() {
	object.deleted = true
}

// reset all shard committee value into default value
func (object *Pdexv3ShareLastLpFeePerShareObject) Reset() bool {
	object.state = NewPdexv3ShareLastLpFeePerShareState()
	return true
}

func (object *Pdexv3ShareLastLpFeePerShareObject) IsDeleted() bool {
	return object.deleted
}

// value is either default or nil
func (object *Pdexv3ShareLastLpFeePerShareObject) IsEmpty() bool {
	temp := NewPdexv3ShareLastLpFeePerShareState()
	return reflect.DeepEqual(temp, object.state) || object.state == nil
}
