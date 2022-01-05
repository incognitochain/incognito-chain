package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type Pdexv3PoolPairOrderRewardDetailState struct {
	tokenID common.Hash
	value   uint64
}

func (state *Pdexv3PoolPairOrderRewardDetailState) Value() uint64 {
	return state.value
}

func (state *Pdexv3PoolPairOrderRewardDetailState) TokenID() common.Hash {
	return state.tokenID
}

func (state *Pdexv3PoolPairOrderRewardDetailState) MarshalJSON() ([]byte, error) {
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

func (state *Pdexv3PoolPairOrderRewardDetailState) UnmarshalJSON(data []byte) error {
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

func (state *Pdexv3PoolPairOrderRewardDetailState) Clone() *Pdexv3PoolPairOrderRewardDetailState {
	return &Pdexv3PoolPairOrderRewardDetailState{
		tokenID: state.tokenID,
		value:   state.value,
	}
}

func NewPdexv3PoolPairOrderRewardDetailState() *Pdexv3PoolPairOrderRewardDetailState {
	return &Pdexv3PoolPairOrderRewardDetailState{}
}

func NewPdexv3PoolPairOrderRewardDetailStateWithValue(
	tokenID common.Hash, value uint64,
) *Pdexv3PoolPairOrderRewardDetailState {
	return &Pdexv3PoolPairOrderRewardDetailState{
		tokenID: tokenID,
		value:   value,
	}
}

type Pdexv3PoolPairOrderRewardDetailObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *Pdexv3PoolPairOrderRewardDetailState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPdexv3PoolPairOrderRewardDetailObject(db *StateDB, hash common.Hash) *Pdexv3PoolPairOrderRewardDetailObject {
	return &Pdexv3PoolPairOrderRewardDetailObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewPdexv3PoolPairOrderRewardDetailState(),
		objectType: Pdexv3PoolPairOrderRewardDetailObjectType,
		deleted:    false,
	}
}

func newPdexv3PoolPairOrderRewardDetailObjectWithValue(
	db *StateDB, key common.Hash, data interface{},
) (*Pdexv3PoolPairOrderRewardDetailObject, error) {
	var newPdexv3PoolPairOrderRewardDetailState = NewPdexv3PoolPairOrderRewardDetailState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPdexv3PoolPairOrderRewardDetailState)
		if err != nil {
			return nil, err
		}
	} else {
		newPdexv3PoolPairOrderRewardDetailState, ok = data.(*Pdexv3PoolPairOrderRewardDetailState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3PoolPairOrderRewardDetailStateType, reflect.TypeOf(data))
		}
	}
	return &Pdexv3PoolPairOrderRewardDetailObject{
		version:    defaultVersion,
		hash:       key,
		state:      newPdexv3PoolPairOrderRewardDetailState,
		db:         db,
		objectType: Pdexv3PoolPairOrderRewardDetailObjectType,
		deleted:    false,
	}, nil
}

func generatePdexv3PoolPairOrderRewardDetailObjectPrefix(poolPairID, nftID string) []byte {
	b := append(GetPdexv3PoolPairOrderRewardPrefix(), []byte(poolPairID)...)
	b = append(b, []byte(nftID)...)
	h := common.HashH(b)
	return h[:prefixHashKeyLength]
}

func GeneratePdexv3PoolPairOrderRewardDetailObjectPrefix(poolPairID, nftID string, tokenID common.Hash) common.Hash {
	prefixHash := generatePdexv3PoolPairOrderRewardDetailObjectPrefix(poolPairID, nftID)
	valueHash := common.HashH(tokenID.Bytes())
	return common.BytesToHash(append(prefixHash, valueHash[:prefixKeyLength]...))
}

func (object *Pdexv3PoolPairOrderRewardDetailObject) GetVersion() int {
	return object.version
}

// setError remembers the first non-nil error it is called with.
func (object *Pdexv3PoolPairOrderRewardDetailObject) SetError(err error) {
	if object.dbErr == nil {
		object.dbErr = err
	}
}

func (object *Pdexv3PoolPairOrderRewardDetailObject) GetTrie(db DatabaseAccessWarper) Trie {
	return object.trie
}

func (object *Pdexv3PoolPairOrderRewardDetailObject) SetValue(data interface{}) error {
	newPdexv3PoolPairOrderRewardDetailState, ok := data.(*Pdexv3PoolPairOrderRewardDetailState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3PoolPairOrderRewardDetailStateType, reflect.TypeOf(data))
	}
	object.state = newPdexv3PoolPairOrderRewardDetailState
	return nil
}

func (object *Pdexv3PoolPairOrderRewardDetailObject) GetValue() interface{} {
	return object.state
}

func (object *Pdexv3PoolPairOrderRewardDetailObject) GetValueBytes() []byte {
	state, ok := object.GetValue().(*Pdexv3PoolPairOrderRewardDetailState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal pdexv3 pool pair order reward state")
	}
	return value
}

func (object *Pdexv3PoolPairOrderRewardDetailObject) GetHash() common.Hash {
	return object.hash
}

func (object *Pdexv3PoolPairOrderRewardDetailObject) GetType() int {
	return object.objectType
}

// MarkDelete will delete an object in trie
func (object *Pdexv3PoolPairOrderRewardDetailObject) MarkDelete() {
	object.deleted = true
}

// reset all shard committee value into default value
func (object *Pdexv3PoolPairOrderRewardDetailObject) Reset() bool {
	object.state = NewPdexv3PoolPairOrderRewardDetailState()
	return true
}

func (object *Pdexv3PoolPairOrderRewardDetailObject) IsDeleted() bool {
	return object.deleted
}

// value is either default or nil
func (object *Pdexv3PoolPairOrderRewardDetailObject) IsEmpty() bool {
	return reflect.DeepEqual(NewPdexv3PoolPairOrderRewardDetailState, object.state) || object.state == nil
}
