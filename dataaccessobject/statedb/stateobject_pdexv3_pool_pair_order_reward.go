package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type Pdexv3PoolPairOrderRewardState struct {
	nftID     string
	accessOTA []byte
}

func (state *Pdexv3PoolPairOrderRewardState) AccessOTA() []byte {
	return state.accessOTA
}

func (state *Pdexv3PoolPairOrderRewardState) NftID() string {
	return state.nftID
}

func (state *Pdexv3PoolPairOrderRewardState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		NftID     string `json:"NftID"`
		AccessOTA []byte `json:"AccessOTA,omitempty"`
	}{
		NftID:     state.nftID,
		AccessOTA: state.accessOTA,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (state *Pdexv3PoolPairOrderRewardState) UnmarshalJSON(data []byte) error {
	temp := struct {
		NftID     string `json:"NftID"`
		AccessOTA []byte `json:"AccessOTA,omitempty"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	state.nftID = temp.NftID
	state.accessOTA = temp.AccessOTA
	return nil
}

func (state *Pdexv3PoolPairOrderRewardState) Clone() *Pdexv3PoolPairOrderRewardState {
	return &Pdexv3PoolPairOrderRewardState{
		nftID:     state.nftID,
		accessOTA: state.accessOTA,
	}
}

func NewPdexv3PoolPairOrderRewardState() *Pdexv3PoolPairOrderRewardState {
	return &Pdexv3PoolPairOrderRewardState{}
}

func NewPdexv3PoolPairOrderRewardStateWithValue(
	nftID string, accessOTA []byte,
) *Pdexv3PoolPairOrderRewardState {
	return &Pdexv3PoolPairOrderRewardState{
		nftID:     nftID,
		accessOTA: accessOTA,
	}
}

type Pdexv3PoolPairOrderRewardObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *Pdexv3PoolPairOrderRewardState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPdexv3PoolPairOrderRewardObject(db *StateDB, hash common.Hash) *Pdexv3PoolPairOrderRewardObject {
	return &Pdexv3PoolPairOrderRewardObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewPdexv3PoolPairOrderRewardState(),
		objectType: Pdexv3PoolPairOrderRewardObjectType,
		deleted:    false,
	}
}

func newPdexv3PoolPairOrderRewardObjectWithValue(
	db *StateDB, key common.Hash, data interface{},
) (*Pdexv3PoolPairOrderRewardObject, error) {
	var newPdexv3PoolPairOrderRewardState = NewPdexv3PoolPairOrderRewardState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPdexv3PoolPairOrderRewardState)
		if err != nil {
			return nil, err
		}
	} else {
		newPdexv3PoolPairOrderRewardState, ok = data.(*Pdexv3PoolPairOrderRewardState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3PoolPairOrderRewardStateType, reflect.TypeOf(data))
		}
	}
	return &Pdexv3PoolPairOrderRewardObject{
		version:    defaultVersion,
		hash:       key,
		state:      newPdexv3PoolPairOrderRewardState,
		db:         db,
		objectType: Pdexv3PoolPairOrderRewardObjectType,
		deleted:    false,
	}, nil
}

func generatePdexv3PoolPairOrderRewardObjectPrefix(poolPairID string) []byte {
	b := append(GetPdexv3PoolPairOrderRewardPrefix(), []byte(poolPairID)...)
	h := common.HashH(b)
	return h[:prefixHashKeyLength]
}

func GeneratePdexv3PoolPairOrderRewardObjectPrefix(poolPairID, nftID string) common.Hash {
	prefixHash := generatePdexv3PoolPairOrderRewardObjectPrefix(poolPairID)
	valueHash := common.HashH([]byte(nftID))
	return common.BytesToHash(append(prefixHash, valueHash[:prefixKeyLength]...))
}

func (object *Pdexv3PoolPairOrderRewardObject) GetVersion() int {
	return object.version
}

// setError remembers the first non-nil error it is called with.
func (object *Pdexv3PoolPairOrderRewardObject) SetError(err error) {
	if object.dbErr == nil {
		object.dbErr = err
	}
}

func (object *Pdexv3PoolPairOrderRewardObject) GetTrie(db DatabaseAccessWarper) Trie {
	return object.trie
}

func (object *Pdexv3PoolPairOrderRewardObject) SetValue(data interface{}) error {
	newPdexv3PoolPairOrderRewardState, ok := data.(*Pdexv3PoolPairOrderRewardState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3PoolPairOrderRewardStateType, reflect.TypeOf(data))
	}
	object.state = newPdexv3PoolPairOrderRewardState
	return nil
}

func (object *Pdexv3PoolPairOrderRewardObject) GetValue() interface{} {
	return object.state
}

func (object *Pdexv3PoolPairOrderRewardObject) GetValueBytes() []byte {
	state, ok := object.GetValue().(*Pdexv3PoolPairOrderRewardState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal pdexv3 pool pair order reward state")
	}
	return value
}

func (object *Pdexv3PoolPairOrderRewardObject) GetHash() common.Hash {
	return object.hash
}

func (object *Pdexv3PoolPairOrderRewardObject) GetType() int {
	return object.objectType
}

// MarkDelete will delete an object in trie
func (object *Pdexv3PoolPairOrderRewardObject) MarkDelete() {
	object.deleted = true
}

// reset all shard committee value into default value
func (object *Pdexv3PoolPairOrderRewardObject) Reset() bool {
	object.state = NewPdexv3PoolPairOrderRewardState()
	return true
}

func (object *Pdexv3PoolPairOrderRewardObject) IsDeleted() bool {
	return object.deleted
}

// value is either default or nil
func (object *Pdexv3PoolPairOrderRewardObject) IsEmpty() bool {
	return reflect.DeepEqual(NewPdexv3PoolPairOrderRewardState, object.state) || object.state == nil
}
