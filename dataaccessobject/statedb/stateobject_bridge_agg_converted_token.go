package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BridgeAggConvertedTokenState struct {
	tokenID   common.Hash
	networkID uint
}

func (state *BridgeAggConvertedTokenState) NetworkID() uint {
	return state.networkID
}

func (state *BridgeAggConvertedTokenState) TokenID() common.Hash {
	return state.tokenID
}

func (state *BridgeAggConvertedTokenState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		TokenID   common.Hash `json:"TokenID"`
		NetworkID uint        `json:"NetworkID,omitempty"`
	}{
		TokenID:   state.tokenID,
		NetworkID: state.networkID,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (state *BridgeAggConvertedTokenState) UnmarshalJSON(data []byte) error {
	temp := struct {
		TokenID   common.Hash `json:"TokenID"`
		NetworkID uint        `json:"NetworkID"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	state.tokenID = temp.TokenID
	state.networkID = temp.NetworkID
	return nil
}

func (state *BridgeAggConvertedTokenState) Clone() *BridgeAggConvertedTokenState {
	return &BridgeAggConvertedTokenState{
		tokenID:   state.tokenID,
		networkID: state.networkID,
	}
}

func NewBridgeAggConvertedTokenState() *BridgeAggConvertedTokenState {
	return &BridgeAggConvertedTokenState{}
}

func NewBridgeAggConvertedTokenStateWithValue(tokenID common.Hash, networkID uint) *BridgeAggConvertedTokenState {
	return &BridgeAggConvertedTokenState{
		tokenID:   tokenID,
		networkID: networkID,
	}
}

type BridgeAggConvertedTokenObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *BridgeAggConvertedTokenState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBridgeAggConvertedObject(db *StateDB, hash common.Hash) *BridgeAggConvertedTokenObject {
	return &BridgeAggConvertedTokenObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewBridgeAggConvertedTokenState(),
		objectType: BridgeAggConvertedTokenObjectType,
		deleted:    false,
	}
}

func newBridgeAggConvertedObjectWithValue(db *StateDB, key common.Hash, data interface{}) (
	*BridgeAggConvertedTokenObject, error,
) {
	var newBridgeAggConvertedTokenState = NewBridgeAggConvertedTokenState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeAggConvertedTokenState)
		if err != nil {
			return nil, err
		}
	} else {
		newBridgeAggConvertedTokenState, ok = data.(*BridgeAggConvertedTokenState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeAggConvertedTokenStateType, reflect.TypeOf(data))
		}
	}
	return &BridgeAggConvertedTokenObject{
		version:    defaultVersion,
		hash:       key,
		state:      newBridgeAggConvertedTokenState,
		db:         db,
		objectType: BridgeAggConvertedTokenObjectType,
		deleted:    false,
	}, nil
}

func generateBridgeAggConvertedTokenObjectPrefix(unifiedTokenID common.Hash) []byte {
	b := append(GetBridgeAggConvertedTokenPrefix(), unifiedTokenID.Bytes()...)
	h := common.HashH(b)
	return h[:prefixHashKeyLength]
}

func GenerateBridgeAggConvertedTokenObjectKey(unifiedTokenID, tokenID common.Hash) common.Hash {
	prefixHash := generateBridgeAggConvertedTokenObjectPrefix(unifiedTokenID)
	valueHash := common.HashH(tokenID.Bytes())
	return common.BytesToHash(append(prefixHash, valueHash[:prefixKeyLength]...))
}

func (object *BridgeAggConvertedTokenObject) GetVersion() int {
	return object.version
}

// setError remembers the first non-nil error it is called with.
func (object *BridgeAggConvertedTokenObject) SetError(err error) {
	if object.dbErr == nil {
		object.dbErr = err
	}
}

func (object *BridgeAggConvertedTokenObject) GetTrie(db DatabaseAccessWarper) Trie {
	return object.trie
}

func (object *BridgeAggConvertedTokenObject) SetValue(data interface{}) error {
	newBridgeAggConvertedTokenState, ok := data.(*BridgeAggConvertedTokenState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeAggConvertedTokenStateType, reflect.TypeOf(data))
	}
	object.state = newBridgeAggConvertedTokenState
	return nil
}

func (object *BridgeAggConvertedTokenObject) GetValue() interface{} {
	return object.state
}

func (object *BridgeAggConvertedTokenObject) GetValueBytes() []byte {
	state, ok := object.GetValue().(*BridgeAggConvertedTokenState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal bridge agg converted token state")
	}
	return value
}

func (object *BridgeAggConvertedTokenObject) GetHash() common.Hash {
	return object.hash
}

func (object *BridgeAggConvertedTokenObject) GetType() int {
	return object.objectType
}

// MarkDelete will delete an object in trie
func (object *BridgeAggConvertedTokenObject) MarkDelete() {
	object.deleted = true
}

// reset all shard committee value into default value
func (object *BridgeAggConvertedTokenObject) Reset() bool {
	object.state = NewBridgeAggConvertedTokenState()
	return true
}

func (object *BridgeAggConvertedTokenObject) IsDeleted() bool {
	return object.deleted
}

// value is either default or nil
func (object *BridgeAggConvertedTokenObject) IsEmpty() bool {
	temp := NewBridgeAggConvertedTokenState()
	return reflect.DeepEqual(temp, object.state) || object.state == nil
}
