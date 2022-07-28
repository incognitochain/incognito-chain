package statedb

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type Pdexv3ShareLastLmRewardPerShareState struct {
	tokenID common.Hash
	value   *big.Int
}

func (state *Pdexv3ShareLastLmRewardPerShareState) Value() big.Int {
	return *state.value
}

func (state *Pdexv3ShareLastLmRewardPerShareState) TokenID() common.Hash {
	return state.tokenID
}

func (state *Pdexv3ShareLastLmRewardPerShareState) MarshalJSON() ([]byte, error) {
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

func (state *Pdexv3ShareLastLmRewardPerShareState) UnmarshalJSON(data []byte) error {
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

func (state *Pdexv3ShareLastLmRewardPerShareState) Clone() *Pdexv3ShareLastLmRewardPerShareState {
	return &Pdexv3ShareLastLmRewardPerShareState{
		tokenID: state.tokenID,
		value:   state.value,
	}
}

func NewPdexv3ShareLastLmRewardPerShareState() *Pdexv3ShareLastLmRewardPerShareState {
	return &Pdexv3ShareLastLmRewardPerShareState{}
}

func NewPdexv3ShareLastLmRewardPerShareStateWithValue(
	tokenID common.Hash, value *big.Int,
) *Pdexv3ShareLastLmRewardPerShareState {
	return &Pdexv3ShareLastLmRewardPerShareState{
		tokenID: tokenID,
		value:   value,
	}
}

type Pdexv3ShareLastLmRewardPerShareObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *Pdexv3ShareLastLmRewardPerShareState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPdexv3ShareLastLmRewardPerShareObject(db *StateDB, hash common.Hash) *Pdexv3ShareLastLmRewardPerShareObject {
	return &Pdexv3ShareLastLmRewardPerShareObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewPdexv3ShareLastLmRewardPerShareState(),
		objectType: Pdexv3ShareLastLmRewardPerShareObjectType,
		deleted:    false,
	}
}

func newPdexv3ShareLastLmRewardPerShareObjectWithValue(
	db *StateDB, key common.Hash, data interface{},
) (*Pdexv3ShareLastLmRewardPerShareObject, error) {
	var newPdexv3ShareLastLmRewardPerShareState = NewPdexv3ShareLastLmRewardPerShareState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPdexv3ShareLastLmRewardPerShareState)
		if err != nil {
			return nil, err
		}
	} else {
		newPdexv3ShareLastLmRewardPerShareState, ok = data.(*Pdexv3ShareLastLmRewardPerShareState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3LastLmRewardPerShareStateType, reflect.TypeOf(data))
		}
	}
	return &Pdexv3ShareLastLmRewardPerShareObject{
		version:    defaultVersion,
		hash:       key,
		state:      newPdexv3ShareLastLmRewardPerShareState,
		db:         db,
		objectType: Pdexv3ShareLastLmRewardPerShareObjectType,
		deleted:    false,
	}, nil
}

func generatePdexv3ShareLastLmRewardPerShareObjectPrefix(poolPairID, nftID string) []byte {
	b := append(GetPdexv3ShareLastLmRewardPerSharesPrefix(), []byte(poolPairID)...)
	b = append(b, []byte(nftID)...)
	h := common.HashH(b)
	return h[:prefixHashKeyLength]
}

func GeneratePdexv3ShareLastLmRewardPerShareObjectKey(stakingPoolID, nftID, tokenID string) common.Hash {
	prefixHash := generatePdexv3ShareLastLmRewardPerShareObjectPrefix(stakingPoolID, nftID)
	valueHash := common.HashH([]byte(tokenID))
	return common.BytesToHash(append(prefixHash, valueHash[:prefixKeyLength]...))
}

func (object *Pdexv3ShareLastLmRewardPerShareObject) GetVersion() int {
	return object.version
}

// setError remembers the first non-nil error it is called with.
func (object *Pdexv3ShareLastLmRewardPerShareObject) SetError(err error) {
	if object.dbErr == nil {
		object.dbErr = err
	}
}

func (object *Pdexv3ShareLastLmRewardPerShareObject) GetTrie(db DatabaseAccessWarper) Trie {
	return object.trie
}

func (object *Pdexv3ShareLastLmRewardPerShareObject) SetValue(data interface{}) error {
	newPdexv3ShareLastLmRewardPerShareState, ok := data.(*Pdexv3ShareLastLmRewardPerShareState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3LastLmRewardPerShareStateType, reflect.TypeOf(data))
	}
	object.state = newPdexv3ShareLastLmRewardPerShareState
	return nil
}

func (object *Pdexv3ShareLastLmRewardPerShareObject) GetValue() interface{} {
	return object.state
}

func (object *Pdexv3ShareLastLmRewardPerShareObject) GetValueBytes() []byte {
	state, ok := object.GetValue().(*Pdexv3ShareLastLmRewardPerShareState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal pdexv3 share lastLmRewardPerShare state")
	}
	return value
}

func (object *Pdexv3ShareLastLmRewardPerShareObject) GetHash() common.Hash {
	return object.hash
}

func (object *Pdexv3ShareLastLmRewardPerShareObject) GetType() int {
	return object.objectType
}

// MarkDelete will delete an object in trie
func (object *Pdexv3ShareLastLmRewardPerShareObject) MarkDelete() {
	object.deleted = true
}

// reset all shard committee value into default value
func (object *Pdexv3ShareLastLmRewardPerShareObject) Reset() bool {
	object.state = NewPdexv3ShareLastLmRewardPerShareState()
	return true
}

func (object *Pdexv3ShareLastLmRewardPerShareObject) IsDeleted() bool {
	return object.deleted
}

// value is either default or nil
func (object *Pdexv3ShareLastLmRewardPerShareObject) IsEmpty() bool {
	temp := NewPdexv3ShareLastLmRewardPerShareState()
	return reflect.DeepEqual(temp, object.state) || object.state == nil
}
