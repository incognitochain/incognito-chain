package statedb

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type Pdexv3StakingPoolRewardPerShareState struct {
	tokenID common.Hash
	value   *big.Int
}

func (state *Pdexv3StakingPoolRewardPerShareState) Value() big.Int {
	return *state.value
}

func (state *Pdexv3StakingPoolRewardPerShareState) TokenID() common.Hash {
	return state.tokenID
}

func (state *Pdexv3StakingPoolRewardPerShareState) MarshalJSON() ([]byte, error) {
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

func (state *Pdexv3StakingPoolRewardPerShareState) UnmarshalJSON(data []byte) error {
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

func (state *Pdexv3StakingPoolRewardPerShareState) Clone() *Pdexv3StakingPoolRewardPerShareState {
	return &Pdexv3StakingPoolRewardPerShareState{
		tokenID: state.tokenID,
		value:   big.NewInt(0).SetBytes(state.value.Bytes()),
	}
}

func NewPdexv3StakingPoolRewardPerShareState() *Pdexv3StakingPoolRewardPerShareState {
	return &Pdexv3StakingPoolRewardPerShareState{}
}

func NewPdexv3StakingPoolRewardPerShareStateWithValue(
	tokenID common.Hash, value *big.Int,
) *Pdexv3StakingPoolRewardPerShareState {
	return &Pdexv3StakingPoolRewardPerShareState{
		tokenID: tokenID,
		value:   value,
	}
}

type Pdexv3StakingPoolRewardPerShareObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *Pdexv3StakingPoolRewardPerShareState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPdexv3StakingPoolRewardPerShareObject(
	db *StateDB, hash common.Hash,
) *Pdexv3StakingPoolRewardPerShareObject {
	return &Pdexv3StakingPoolRewardPerShareObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewPdexv3StakingPoolRewardPerShareState(),
		objectType: Pdexv3StakingPoolRewardPerShareObjectType,
		deleted:    false,
	}
}

func newPdexv3StakingPoolRewardPerShareObjectWithValue(
	db *StateDB, key common.Hash, data interface{},
) (*Pdexv3StakingPoolRewardPerShareObject, error) {
	var newPdexv3StakingPoolRewardPerShareState = NewPdexv3StakingPoolRewardPerShareState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPdexv3StakingPoolRewardPerShareState)
		if err != nil {
			return nil, err
		}
	} else {
		newPdexv3StakingPoolRewardPerShareState, ok = data.(*Pdexv3StakingPoolRewardPerShareState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3StakingPoolRewardPerShareStateType, reflect.TypeOf(data))
		}
	}
	return &Pdexv3StakingPoolRewardPerShareObject{
		version:    defaultVersion,
		hash:       key,
		state:      newPdexv3StakingPoolRewardPerShareState,
		db:         db,
		objectType: Pdexv3StakingPoolRewardPerShareObjectType,
		deleted:    false,
	}, nil
}

func generatePdexv3StakingPoolRewardPerShareObjectPrefix(stakingPoolID string) []byte {
	b := append(GetPdexv3StakingPoolRewardPerSharePrefix(), []byte(stakingPoolID)...)
	h := common.HashH(b)
	return h[:prefixHashKeyLength]
}

func GeneratePdexv3StakingPoolRewardPerShareObjectKey(stakingPoolID, tokenID string) common.Hash {
	prefixHash := generatePdexv3StakingPoolRewardPerShareObjectPrefix(stakingPoolID)
	valueHash := common.HashH([]byte(tokenID))
	return common.BytesToHash(append(prefixHash, valueHash[:prefixKeyLength]...))
}

func (object *Pdexv3StakingPoolRewardPerShareObject) GetVersion() int {
	return object.version
}

// setError remembers the first non-nil error it is called with.
func (object *Pdexv3StakingPoolRewardPerShareObject) SetError(err error) {
	if object.dbErr == nil {
		object.dbErr = err
	}
}

func (object *Pdexv3StakingPoolRewardPerShareObject) GetTrie(db DatabaseAccessWarper) Trie {
	return object.trie
}

func (object *Pdexv3StakingPoolRewardPerShareObject) SetValue(data interface{}) error {
	newPdexv3StakingPoolRewardPerShareState, ok := data.(*Pdexv3StakingPoolRewardPerShareState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3StakingPoolRewardPerShareStateType, reflect.TypeOf(data))
	}
	object.state = newPdexv3StakingPoolRewardPerShareState
	return nil
}

func (object *Pdexv3StakingPoolRewardPerShareObject) GetValue() interface{} {
	return object.state
}

func (object *Pdexv3StakingPoolRewardPerShareObject) GetValueBytes() []byte {
	state, ok := object.GetValue().(*Pdexv3StakingPoolRewardPerShareState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal pdexv3 staking pool reward per share state")
	}
	return value
}

func (object *Pdexv3StakingPoolRewardPerShareObject) GetHash() common.Hash {
	return object.hash
}

func (object *Pdexv3StakingPoolRewardPerShareObject) GetType() int {
	return object.objectType
}

// MarkDelete will delete an object in trie
func (object *Pdexv3StakingPoolRewardPerShareObject) MarkDelete() {
	object.deleted = true
}

// reset all shard committee value into default value
func (object *Pdexv3StakingPoolRewardPerShareObject) Reset() bool {
	object.state = NewPdexv3StakingPoolRewardPerShareState()
	return true
}

func (object *Pdexv3StakingPoolRewardPerShareObject) IsDeleted() bool {
	return object.deleted
}

// value is either default or nil
func (object *Pdexv3StakingPoolRewardPerShareObject) IsEmpty() bool {
	temp := NewPdexv3StakingPoolRewardPerShareState()
	return reflect.DeepEqual(temp, object.state) || object.state == nil
}
