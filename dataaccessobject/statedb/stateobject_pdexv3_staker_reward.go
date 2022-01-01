package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type Pdexv3StakerRewardState struct {
	tokenID common.Hash
	value   uint64
}

func (state *Pdexv3StakerRewardState) Value() uint64 {
	return state.value
}

func (state *Pdexv3StakerRewardState) TokenID() common.Hash {
	return state.tokenID
}

func (state *Pdexv3StakerRewardState) MarshalJSON() ([]byte, error) {
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

func (state *Pdexv3StakerRewardState) UnmarshalJSON(data []byte) error {
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

func (state *Pdexv3StakerRewardState) Clone() *Pdexv3StakerRewardState {
	return &Pdexv3StakerRewardState{
		tokenID: state.tokenID,
		value:   state.value,
	}
}

func NewPdexv3StakerRewardState() *Pdexv3StakerRewardState {
	return &Pdexv3StakerRewardState{}
}

func NewPdexv3StakerRewardStateWithValue(
	tokenID common.Hash, value uint64,
) *Pdexv3StakerRewardState {
	return &Pdexv3StakerRewardState{
		tokenID: tokenID,
		value:   value,
	}
}

type Pdexv3StakerRewardObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *Pdexv3StakerRewardState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPdexv3StakerRewardObject(db *StateDB, hash common.Hash) *Pdexv3StakerRewardObject {
	return &Pdexv3StakerRewardObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewPdexv3StakerRewardState(),
		objectType: Pdexv3StakerRewardObjectType,
		deleted:    false,
	}
}

func newPdexv3StakerRewardObjectWithValue(
	db *StateDB, key common.Hash, data interface{},
) (*Pdexv3StakerRewardObject, error) {
	var newPdexv3StakerRewardState = NewPdexv3StakerRewardState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPdexv3StakerRewardState)
		if err != nil {
			return nil, err
		}
	} else {
		newPdexv3StakerRewardState, ok = data.(*Pdexv3StakerRewardState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3StakerRewardStateType, reflect.TypeOf(data))
		}
	}
	return &Pdexv3StakerRewardObject{
		version:    defaultVersion,
		hash:       key,
		state:      newPdexv3StakerRewardState,
		db:         db,
		objectType: Pdexv3StakerRewardObjectType,
		deleted:    false,
	}, nil
}

func generatePdexv3StakerRewardObjectPrefix(stakingPoolID, nftID string) []byte {
	b := append(GetPdexv3StakerReward(), []byte(stakingPoolID)...)
	b = append(b, []byte(nftID)...)
	h := common.HashH(b)
	return h[:prefixHashKeyLength]
}

func GeneratePdexv3StakerRewardObjectKey(stakingPoolID, nftID, tokenID string) common.Hash {
	prefixHash := generatePdexv3StakerRewardObjectPrefix(stakingPoolID, nftID)
	valueHash := common.HashH([]byte(tokenID))
	return common.BytesToHash(append(prefixHash, valueHash[:prefixKeyLength]...))
}

func (object *Pdexv3StakerRewardObject) GetVersion() int {
	return object.version
}

// setError remembers the first non-nil error it is called with.
func (object *Pdexv3StakerRewardObject) SetError(err error) {
	if object.dbErr == nil {
		object.dbErr = err
	}
}

func (object *Pdexv3StakerRewardObject) GetTrie(db DatabaseAccessWarper) Trie {
	return object.trie
}

func (object *Pdexv3StakerRewardObject) SetValue(data interface{}) error {
	newPdexv3StakerRewardState, ok := data.(*Pdexv3StakerRewardState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3StakerRewardStateType, reflect.TypeOf(data))
	}
	object.state = newPdexv3StakerRewardState
	return nil
}

func (object *Pdexv3StakerRewardObject) GetValue() interface{} {
	return object.state
}

func (object *Pdexv3StakerRewardObject) GetValueBytes() []byte {
	state, ok := object.GetValue().(*Pdexv3StakerRewardState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal pdexv3 staker reward state")
	}
	return value
}

func (object *Pdexv3StakerRewardObject) GetHash() common.Hash {
	return object.hash
}

func (object *Pdexv3StakerRewardObject) GetType() int {
	return object.objectType
}

// MarkDelete will delete an object in trie
func (object *Pdexv3StakerRewardObject) MarkDelete() {
	object.deleted = true
}

// reset all shard committee value into default value
func (object *Pdexv3StakerRewardObject) Reset() bool {
	object.state = NewPdexv3StakerRewardState()
	return true
}

func (object *Pdexv3StakerRewardObject) IsDeleted() bool {
	return object.deleted
}

// value is either default or nil
func (object *Pdexv3StakerRewardObject) IsEmpty() bool {
	temp := NewPdexv3StakerState()
	return reflect.DeepEqual(temp, object.state) || object.state == nil
}
