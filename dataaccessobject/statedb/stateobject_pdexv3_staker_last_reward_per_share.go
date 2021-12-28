package statedb

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type Pdexv3StakerLastRewardPerShareState struct {
	tokenID common.Hash
	value   *big.Int
}

func (state *Pdexv3StakerLastRewardPerShareState) Value() big.Int {
	return *state.value
}

func (state *Pdexv3StakerLastRewardPerShareState) TokenID() common.Hash {
	return state.tokenID
}

func (state *Pdexv3StakerLastRewardPerShareState) MarshalJSON() ([]byte, error) {
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

func (state *Pdexv3StakerLastRewardPerShareState) UnmarshalJSON(data []byte) error {
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

func (state *Pdexv3StakerLastRewardPerShareState) Clone() *Pdexv3StakerLastRewardPerShareState {
	return &Pdexv3StakerLastRewardPerShareState{
		tokenID: state.tokenID,
		value:   big.NewInt(0).SetBytes(state.value.Bytes()),
	}
}

func NewPdexv3StakerLastRewardPerShareState() *Pdexv3StakerLastRewardPerShareState {
	return &Pdexv3StakerLastRewardPerShareState{}
}

func NewPdexv3StakerLastRewardPerShareStateWithValue(
	tokenID common.Hash, value *big.Int,
) *Pdexv3StakerLastRewardPerShareState {
	return &Pdexv3StakerLastRewardPerShareState{
		tokenID: tokenID,
		value:   value,
	}
}

type Pdexv3StakerLastRewardPerShareObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *Pdexv3StakerLastRewardPerShareState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPdexv3StakerLastRewardPerShareObject(db *StateDB, hash common.Hash) *Pdexv3StakerLastRewardPerShareObject {
	return &Pdexv3StakerLastRewardPerShareObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewPdexv3StakerLastRewardPerShareState(),
		objectType: Pdexv3StakerLastRewardPerShareObjectType,
		deleted:    false,
	}
}

func newPdexv3StakerLastRewardPerShareObjectWithValue(
	db *StateDB, key common.Hash, data interface{},
) (*Pdexv3StakerLastRewardPerShareObject, error) {
	var newPdexv3StakerLastRewardPerShareState = NewPdexv3StakerLastRewardPerShareState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPdexv3StakerLastRewardPerShareState)
		if err != nil {
			return nil, err
		}
	} else {
		newPdexv3StakerLastRewardPerShareState, ok = data.(*Pdexv3StakerLastRewardPerShareState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3StakerLastRewardPerShareStateType, reflect.TypeOf(data))
		}
	}
	return &Pdexv3StakerLastRewardPerShareObject{
		version:    defaultVersion,
		hash:       key,
		state:      newPdexv3StakerLastRewardPerShareState,
		db:         db,
		objectType: Pdexv3StakerLastRewardPerShareObjectType,
		deleted:    false,
	}, nil
}

func generatePdexv3StakerLastRewardPerShareObjectPrefix(stakingPoolID, nftID string) []byte {
	b := append(GetPdexv3StakerLastRewardPerShare(), []byte(stakingPoolID)...)
	b = append(b, []byte(nftID)...)
	h := common.HashH(b)
	return h[:prefixHashKeyLength]
}

func GeneratePdexv3StakerLastRewardPerShareObjectKey(stakingPoolID, nftID, tokenID string) common.Hash {
	prefixHash := generatePdexv3StakerLastRewardPerShareObjectPrefix(stakingPoolID, nftID)
	valueHash := common.HashH([]byte(tokenID))
	return common.BytesToHash(append(prefixHash, valueHash[:prefixKeyLength]...))
}

func (object *Pdexv3StakerLastRewardPerShareObject) GetVersion() int {
	return object.version
}

// setError remembers the first non-nil error it is called with.
func (object *Pdexv3StakerLastRewardPerShareObject) SetError(err error) {
	if object.dbErr == nil {
		object.dbErr = err
	}
}

func (object *Pdexv3StakerLastRewardPerShareObject) GetTrie(db DatabaseAccessWarper) Trie {
	return object.trie
}

func (object *Pdexv3StakerLastRewardPerShareObject) SetValue(data interface{}) error {
	newPdexv3StakerLastRewardPerShareState, ok := data.(*Pdexv3StakerLastRewardPerShareState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3StakerLastRewardPerShareStateType, reflect.TypeOf(data))
	}
	object.state = newPdexv3StakerLastRewardPerShareState
	return nil
}

func (object *Pdexv3StakerLastRewardPerShareObject) GetValue() interface{} {
	return object.state
}

func (object *Pdexv3StakerLastRewardPerShareObject) GetValueBytes() []byte {
	state, ok := object.GetValue().(*Pdexv3StakerLastRewardPerShareState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal pdexv3 staker last reward per share state")
	}
	return value
}

func (object *Pdexv3StakerLastRewardPerShareObject) GetHash() common.Hash {
	return object.hash
}

func (object *Pdexv3StakerLastRewardPerShareObject) GetType() int {
	return object.objectType
}

// MarkDelete will delete an object in trie
func (object *Pdexv3StakerLastRewardPerShareObject) MarkDelete() {
	object.deleted = true
}

// reset all shard committee value into default value
func (object *Pdexv3StakerLastRewardPerShareObject) Reset() bool {
	object.state = NewPdexv3StakerLastRewardPerShareState()
	return true
}

func (object *Pdexv3StakerLastRewardPerShareObject) IsDeleted() bool {
	return object.deleted
}

// value is either default or nil
func (object *Pdexv3StakerLastRewardPerShareObject) IsEmpty() bool {
	temp := NewPdexv3StakerLastRewardPerShareState()
	return reflect.DeepEqual(temp, object.state) || object.state == nil
}
