package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BridgeAURORATxState struct {
	uniqueAURORATx []byte
}

func (auroraTx BridgeAURORATxState) UniqueAURORATx() []byte {
	return auroraTx.uniqueAURORATx
}

func (auroraTx *BridgeAURORATxState) SetUniqueAURORATx(uniqueAURORATx []byte) {
	auroraTx.uniqueAURORATx = uniqueAURORATx
}

func NewBridgeAURORATxState() *BridgeAURORATxState {
	return &BridgeAURORATxState{}
}

func NewBridgeAURORATxStateWithValue(uniqueAURORATx []byte) *BridgeAURORATxState {
	return &BridgeAURORATxState{uniqueAURORATx: uniqueAURORATx}
}

func (auroraTx BridgeAURORATxState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		UniqueAURORATx []byte
	}{
		UniqueAURORATx: auroraTx.uniqueAURORATx,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (auroraTx *BridgeAURORATxState) UnmarshalJSON(data []byte) error {
	temp := struct {
		UniqueAURORATx []byte
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	auroraTx.uniqueAURORATx = temp.UniqueAURORATx
	return nil
}

type BridgeAURORATxObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version             int
	bridgeAURORATxHash  common.Hash
	bridgeAURORATxState *BridgeAURORATxState
	objectType          int
	deleted             bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBridgeAURORATxObject(db *StateDB, hash common.Hash) *BridgeAURORATxObject {
	return &BridgeAURORATxObject{
		version:             defaultVersion,
		db:                  db,
		bridgeAURORATxHash:  hash,
		bridgeAURORATxState: NewBridgeAURORATxState(),
		objectType:          BridgeAURORATxObjectType,
		deleted:             false,
	}
}

func newBridgeAURORATxObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BridgeAURORATxObject, error) {
	var newBridgeAURORATxState = NewBridgeAURORATxState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeAURORATxState)
		if err != nil {
			return nil, err
		}
	} else {
		newBridgeAURORATxState, ok = data.(*BridgeAURORATxState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeAURORATxStateType, reflect.TypeOf(data))
		}
	}
	return &BridgeAURORATxObject{
		version:             defaultVersion,
		bridgeAURORATxHash:  key,
		bridgeAURORATxState: newBridgeAURORATxState,
		db:                  db,
		objectType:          BridgeAURORATxObjectType,
		deleted:             false,
	}, nil
}

func GenerateBridgeAURORATxObjectKey(uniqueAURORATx []byte) common.Hash {
	prefixHash := GetBridgeAURORATxPrefix()
	valueHash := common.HashH(uniqueAURORATx)
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (auroraTx BridgeAURORATxObject) GetVersion() int {
	return auroraTx.version
}

// setError remembers the first non-nil error it is called with.
func (auroraTx *BridgeAURORATxObject) SetError(err error) {
	if auroraTx.dbErr == nil {
		auroraTx.dbErr = err
	}
}

func (auroraTx BridgeAURORATxObject) GetTrie(db DatabaseAccessWarper) Trie {
	return auroraTx.trie
}

func (auroraTx *BridgeAURORATxObject) SetValue(data interface{}) error {
	var newBridgeAURORATxState = NewBridgeAURORATxState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeAURORATxState)
		if err != nil {
			return err
		}
	} else {
		newBridgeAURORATxState, ok = data.(*BridgeAURORATxState)
		if !ok {
			return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeAURORATxStateType, reflect.TypeOf(data))
		}
	}
	auroraTx.bridgeAURORATxState = newBridgeAURORATxState
	return nil
}

func (auroraTx BridgeAURORATxObject) GetValue() interface{} {
	return auroraTx.bridgeAURORATxState
}

func (auroraTx BridgeAURORATxObject) GetValueBytes() []byte {
	data := auroraTx.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal bridge BSC tx state")
	}
	return value
}

func (auroraTx BridgeAURORATxObject) GetHash() common.Hash {
	return auroraTx.bridgeAURORATxHash
}

func (auroraTx BridgeAURORATxObject) GetType() int {
	return auroraTx.objectType
}

// MarkDelete will delete an object in trie
func (auroraTx *BridgeAURORATxObject) MarkDelete() {
	auroraTx.deleted = true
}

func (auroraTx *BridgeAURORATxObject) Reset() bool {
	auroraTx.bridgeAURORATxState = NewBridgeAURORATxState()
	return true
}

func (auroraTx BridgeAURORATxObject) IsDeleted() bool {
	return auroraTx.deleted
}

// value is either default or nil
func (auroraTx BridgeAURORATxObject) IsEmpty() bool {
	temp := NewBridgeAURORATxState()
	return reflect.DeepEqual(temp, auroraTx.bridgeAURORATxState) || auroraTx.bridgeAURORATxState == nil
}
