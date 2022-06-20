package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BridgeAggUnifiedTokenState struct {
	tokenID common.Hash
}

func (state *BridgeAggUnifiedTokenState) TokenID() common.Hash {
	return state.tokenID
}

func (state *BridgeAggUnifiedTokenState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		TokenID common.Hash `json:"TokenID"`
	}{
		TokenID: state.tokenID,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (state *BridgeAggUnifiedTokenState) UnmarshalJSON(data []byte) error {
	temp := struct {
		TokenID common.Hash `json:"TokenID"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	state.tokenID = temp.TokenID
	return nil
}

func (state *BridgeAggUnifiedTokenState) Clone() *BridgeAggUnifiedTokenState {
	return &BridgeAggUnifiedTokenState{
		tokenID: state.tokenID,
	}
}

func NewBridgeAggUnifiedTokenState() *BridgeAggUnifiedTokenState {
	return &BridgeAggUnifiedTokenState{}
}

func NewBridgeAggUnifiedTokenStateWithValue(tokenID common.Hash) *BridgeAggUnifiedTokenState {
	return &BridgeAggUnifiedTokenState{
		tokenID: tokenID,
	}
}

type BridgeAggUnifiedTokenObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *BridgeAggUnifiedTokenState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBridgeAggUnifiedTokenObject(db *StateDB, hash common.Hash) *BridgeAggUnifiedTokenObject {
	return &BridgeAggUnifiedTokenObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewBridgeAggUnifiedTokenState(),
		objectType: BridgeAggUnifiedTokenObjectType,
		deleted:    false,
	}
}

func newBridgeAggUnifiedTokenObjectWithValue(db *StateDB, key common.Hash, data interface{}) (
	*BridgeAggUnifiedTokenObject, error,
) {
	var newBridgeAggUnifiedTokenState = NewBridgeAggUnifiedTokenState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeAggUnifiedTokenState)
		if err != nil {
			return nil, err
		}
	} else {
		newBridgeAggUnifiedTokenState, ok = data.(*BridgeAggUnifiedTokenState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeAggUnifiedTokenStateType, reflect.TypeOf(data))
		}
	}
	return &BridgeAggUnifiedTokenObject{
		version:    defaultVersion,
		hash:       key,
		state:      newBridgeAggUnifiedTokenState,
		db:         db,
		objectType: BridgeAggUnifiedTokenObjectType,
		deleted:    false,
	}, nil
}

func generateBridgeAggUnifiedTokenObjectPrefix() []byte {
	return GetBridgeAggUnifiedTokenPrefix()
}

func GenerateBridgeAggUnifiedTokenObjectKey(tokenID common.Hash) common.Hash {
	prefixHash := GetBridgeAggUnifiedTokenPrefix()
	valueHash := common.HashH(tokenID.Bytes())
	return common.BytesToHash(append(prefixHash, valueHash[:prefixKeyLength]...))
}

func (object *BridgeAggUnifiedTokenObject) GetVersion() int {
	return object.version
}

// setError remembers the first non-nil error it is called with.
func (object *BridgeAggUnifiedTokenObject) SetError(err error) {
	if object.dbErr == nil {
		object.dbErr = err
	}
}

func (object *BridgeAggUnifiedTokenObject) GetTrie(db DatabaseAccessWarper) Trie {
	return object.trie
}

func (object *BridgeAggUnifiedTokenObject) SetValue(data interface{}) error {
	newBridgeAggUnifiedTokenState, ok := data.(*BridgeAggUnifiedTokenState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeAggUnifiedTokenStateType, reflect.TypeOf(data))
	}
	object.state = newBridgeAggUnifiedTokenState
	return nil
}

func (object *BridgeAggUnifiedTokenObject) GetValue() interface{} {
	return object.state
}

func (object *BridgeAggUnifiedTokenObject) GetValueBytes() []byte {
	state, ok := object.GetValue().(*BridgeAggUnifiedTokenState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal bridge agg unified token state")
	}
	return value
}

func (object *BridgeAggUnifiedTokenObject) GetHash() common.Hash {
	return object.hash
}

func (object *BridgeAggUnifiedTokenObject) GetType() int {
	return object.objectType
}

// MarkDelete will delete an object in trie
func (object *BridgeAggUnifiedTokenObject) MarkDelete() {
	object.deleted = true
}

// reset all shard committee value into default value
func (object *BridgeAggUnifiedTokenObject) Reset() bool {
	object.state = NewBridgeAggUnifiedTokenState()
	return true
}

func (object *BridgeAggUnifiedTokenObject) IsDeleted() bool {
	return object.deleted
}

// value is either default or nil
func (object *BridgeAggUnifiedTokenObject) IsEmpty() bool {
	temp := NewBridgeAggUnifiedTokenState()
	return reflect.DeepEqual(temp, object.state) || object.state == nil
}
