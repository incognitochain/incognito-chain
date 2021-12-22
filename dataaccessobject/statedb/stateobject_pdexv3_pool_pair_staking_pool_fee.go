package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type Pdexv3PoolPairStakingPoolFeeState struct {
	tokenID common.Hash
	value   uint64
}

func (state *Pdexv3PoolPairStakingPoolFeeState) Value() uint64 {
	return state.value
}

func (state *Pdexv3PoolPairStakingPoolFeeState) TokenID() common.Hash {
	return state.tokenID
}

func (state *Pdexv3PoolPairStakingPoolFeeState) MarshalJSON() ([]byte, error) {
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

func (state *Pdexv3PoolPairStakingPoolFeeState) UnmarshalJSON(data []byte) error {
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

func (state *Pdexv3PoolPairStakingPoolFeeState) Clone() *Pdexv3PoolPairStakingPoolFeeState {
	return &Pdexv3PoolPairStakingPoolFeeState{
		tokenID: state.tokenID,
		value:   state.value,
	}
}

func NewPdexv3PoolPairStakingPoolFeeState() *Pdexv3PoolPairStakingPoolFeeState {
	return &Pdexv3PoolPairStakingPoolFeeState{}
}

func NewPdexv3PoolPairStakingPoolFeeStateWithValue(
	tokenID common.Hash, value uint64,
) *Pdexv3PoolPairStakingPoolFeeState {
	return &Pdexv3PoolPairStakingPoolFeeState{
		tokenID: tokenID,
		value:   value,
	}
}

type Pdexv3PoolPairStakingPoolFeeObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *Pdexv3PoolPairStakingPoolFeeState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPdexv3PoolPairStakingPoolFeeObject(
	db *StateDB, hash common.Hash,
) *Pdexv3PoolPairStakingPoolFeeObject {
	return &Pdexv3PoolPairStakingPoolFeeObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewPdexv3PoolPairStakingPoolFeeState(),
		objectType: Pdexv3PoolPairStakingPoolFeeObjectType,
		deleted:    false,
	}
}

func newPdexv3PoolPairStakingPoolFeeObjectWithValue(
	db *StateDB, key common.Hash, data interface{},
) (*Pdexv3PoolPairStakingPoolFeeObject, error) {
	var newPdexv3PoolPairStakingPoolFeeState = NewPdexv3PoolPairStakingPoolFeeState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPdexv3PoolPairStakingPoolFeeState)
		if err != nil {
			return nil, err
		}
	} else {
		newPdexv3PoolPairStakingPoolFeeState, ok = data.(*Pdexv3PoolPairStakingPoolFeeState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3PoolPairStakingPoolFeeStateType, reflect.TypeOf(data))
		}
	}
	return &Pdexv3PoolPairStakingPoolFeeObject{
		version:    defaultVersion,
		hash:       key,
		state:      newPdexv3PoolPairStakingPoolFeeState,
		db:         db,
		objectType: Pdexv3PoolPairStakingPoolFeeObjectType,
		deleted:    false,
	}, nil
}

func generatePdexv3PoolPairStakingPoolFeeObjectPrefix(poolPairID string) []byte {
	b := append(GetPdexv3PoolPairStakingPoolFeesPrefix(), []byte(poolPairID)...)
	h := common.HashH(b)
	return h[:prefixHashKeyLength]
}

func GeneratePdexv3PoolPairStakingPoolFeeObjectKey(poolPairID, tokenID string) common.Hash {
	prefixHash := generatePdexv3PoolPairStakingPoolFeeObjectPrefix(poolPairID)
	valueHash := common.HashH([]byte(tokenID))
	return common.BytesToHash(append(prefixHash, valueHash[:prefixKeyLength]...))
}

func (object *Pdexv3PoolPairStakingPoolFeeObject) GetVersion() int {
	return object.version
}

// setError remembers the first non-nil error it is called with.
func (object *Pdexv3PoolPairStakingPoolFeeObject) SetError(err error) {
	if object.dbErr == nil {
		object.dbErr = err
	}
}

func (object *Pdexv3PoolPairStakingPoolFeeObject) GetTrie(db DatabaseAccessWarper) Trie {
	return object.trie
}

func (object *Pdexv3PoolPairStakingPoolFeeObject) SetValue(data interface{}) error {
	newPdexv3PoolPairStakingPoolFeeState, ok := data.(*Pdexv3PoolPairStakingPoolFeeState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3PoolPairStakingPoolFeeStateType, reflect.TypeOf(data))
	}
	object.state = newPdexv3PoolPairStakingPoolFeeState
	return nil
}

func (object *Pdexv3PoolPairStakingPoolFeeObject) GetValue() interface{} {
	return object.state
}

func (object *Pdexv3PoolPairStakingPoolFeeObject) GetValueBytes() []byte {
	state, ok := object.GetValue().(*Pdexv3PoolPairStakingPoolFeeState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal pdexv3 pool pair staking pool fee state")
	}
	return value
}

func (object *Pdexv3PoolPairStakingPoolFeeObject) GetHash() common.Hash {
	return object.hash
}

func (object *Pdexv3PoolPairStakingPoolFeeObject) GetType() int {
	return object.objectType
}

// MarkDelete will delete an object in trie
func (object *Pdexv3PoolPairStakingPoolFeeObject) MarkDelete() {
	object.deleted = true
}

// reset all shard committee value into default value
func (object *Pdexv3PoolPairStakingPoolFeeObject) Reset() bool {
	object.state = NewPdexv3PoolPairStakingPoolFeeState()
	return true
}

func (object *Pdexv3PoolPairStakingPoolFeeObject) IsDeleted() bool {
	return object.deleted
}

// value is either default or nil
func (object *Pdexv3PoolPairStakingPoolFeeObject) IsEmpty() bool {
	temp := NewPdexv3PoolPairStakingPoolFeeState()
	return reflect.DeepEqual(temp, object.state) || object.state == nil
}
