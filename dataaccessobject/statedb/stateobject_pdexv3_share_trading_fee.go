package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type Pdexv3ShareTradingFeeState struct {
	tokenID common.Hash
	value   uint64
}

func (state *Pdexv3ShareTradingFeeState) Value() uint64 {
	return state.value
}

func (state *Pdexv3ShareTradingFeeState) TokenID() common.Hash {
	return state.tokenID
}

func (state *Pdexv3ShareTradingFeeState) MarshalJSON() ([]byte, error) {
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

func (state *Pdexv3ShareTradingFeeState) UnmarshalJSON(data []byte) error {
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

func (state *Pdexv3ShareTradingFeeState) Clone() *Pdexv3ShareTradingFeeState {
	return &Pdexv3ShareTradingFeeState{
		tokenID: state.tokenID,
		value:   state.value,
	}
}

func NewPdexv3ShareTradingFeeState() *Pdexv3ShareTradingFeeState { return &Pdexv3ShareTradingFeeState{} }

func NewPdexv3ShareTradingFeeStateWithValue(tokenID common.Hash, value uint64) *Pdexv3ShareTradingFeeState {
	return &Pdexv3ShareTradingFeeState{
		tokenID: tokenID,
		value:   value,
	}
}

type Pdexv3ShareTradingFeeObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *Pdexv3ShareTradingFeeState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPdexv3ShareTradingFeeObject(db *StateDB, hash common.Hash) *Pdexv3ShareTradingFeeObject {
	return &Pdexv3ShareTradingFeeObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewPdexv3ShareTradingFeeState(),
		objectType: Pdexv3ShareTradingFeeObjectType,
		deleted:    false,
	}
}

func newPdexv3ShareTradingFeeObjectWithValue(
	db *StateDB, key common.Hash, data interface{},
) (*Pdexv3ShareTradingFeeObject, error) {
	var newPdexv3ShareTradingFeeState = NewPdexv3ShareTradingFeeState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPdexv3ShareTradingFeeState)
		if err != nil {
			return nil, err
		}
	} else {
		newPdexv3ShareTradingFeeState, ok = data.(*Pdexv3ShareTradingFeeState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3ShareTradingFeeStateType, reflect.TypeOf(data))
		}
	}
	return &Pdexv3ShareTradingFeeObject{
		version:    defaultVersion,
		hash:       key,
		state:      newPdexv3ShareTradingFeeState,
		db:         db,
		objectType: Pdexv3ShareTradingFeeObjectType,
		deleted:    false,
	}, nil
}

func generatePdexv3ShareTradingFeeObjectPrefix(poolPairID, nftID string) []byte {
	b := append(GetPdexv3ShareTradingFeesPrefix(), []byte(poolPairID)...)
	b = append(b, []byte(nftID)...)
	h := common.HashH(b)
	return h[:prefixHashKeyLength]
}

func GeneratePdexv3ShareTradingFeeObjectKey(poolPairID, nftID, tokenID string) common.Hash {
	prefixHash := generatePdexv3ShareTradingFeeObjectPrefix(poolPairID, nftID)
	valueHash := common.HashH([]byte(tokenID))
	return common.BytesToHash(append(prefixHash, valueHash[:prefixKeyLength]...))
}

func (object *Pdexv3ShareTradingFeeObject) GetVersion() int {
	return object.version
}

// setError remembers the first non-nil error it is called with.
func (object *Pdexv3ShareTradingFeeObject) SetError(err error) {
	if object.dbErr == nil {
		object.dbErr = err
	}
}

func (object *Pdexv3ShareTradingFeeObject) GetTrie(db DatabaseAccessWarper) Trie {
	return object.trie
}

func (object *Pdexv3ShareTradingFeeObject) SetValue(data interface{}) error {
	newPdexv3ShareTradingFeeState, ok := data.(*Pdexv3ShareTradingFeeState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3ShareTradingFeeStateType, reflect.TypeOf(data))
	}
	object.state = newPdexv3ShareTradingFeeState
	return nil
}

func (object *Pdexv3ShareTradingFeeObject) GetValue() interface{} {
	return object.state
}

func (object *Pdexv3ShareTradingFeeObject) GetValueBytes() []byte {
	state, ok := object.GetValue().(*Pdexv3ShareTradingFeeState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal pdexv3 share trading fee state")
	}
	return value
}

func (object *Pdexv3ShareTradingFeeObject) GetHash() common.Hash {
	return object.hash
}

func (object *Pdexv3ShareTradingFeeObject) GetType() int {
	return object.objectType
}

// MarkDelete will delete an object in trie
func (object *Pdexv3ShareTradingFeeObject) MarkDelete() {
	object.deleted = true
}

// reset all shard committee value into default value
func (object *Pdexv3ShareTradingFeeObject) Reset() bool {
	object.state = NewPdexv3ShareTradingFeeState()
	return true
}

func (object *Pdexv3ShareTradingFeeObject) IsDeleted() bool {
	return object.deleted
}

// value is either default or nil
func (object *Pdexv3ShareTradingFeeObject) IsEmpty() bool {
	temp := NewPdexv3ShareState()
	return reflect.DeepEqual(temp, object.state) || object.state == nil
}
