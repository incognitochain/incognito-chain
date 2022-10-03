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

func (ftmTx BridgeAURORATxState) UniqueAURORATx() []byte {
	return ftmTx.uniqueAURORATx
}

func (ftmTx *BridgeAURORATxState) SetUniqueAURORATx(uniqueAURORATx []byte) {
	ftmTx.uniqueAURORATx = uniqueAURORATx
}

func NewBridgeAURORATxState() *BridgeAURORATxState {
	return &BridgeAURORATxState{}
}

func NewBridgeAURORATxStateWithValue(uniqueAURORATx []byte) *BridgeAURORATxState {
	return &BridgeAURORATxState{uniqueAURORATx: uniqueAURORATx}
}

func (ftmTx BridgeAURORATxState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		UniqueAURORATx []byte
	}{
		UniqueAURORATx: ftmTx.uniqueAURORATx,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (ftmTx *BridgeAURORATxState) UnmarshalJSON(data []byte) error {
	temp := struct {
		UniqueAURORATx []byte
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	ftmTx.uniqueAURORATx = temp.UniqueAURORATx
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

func (ftmTx BridgeAURORATxObject) GetVersion() int {
	return ftmTx.version
}

// setError remembers the first non-nil error it is called with.
func (ftmTx *BridgeAURORATxObject) SetError(err error) {
	if ftmTx.dbErr == nil {
		ftmTx.dbErr = err
	}
}

func (ftmTx BridgeAURORATxObject) GetTrie(db DatabaseAccessWarper) Trie {
	return ftmTx.trie
}

func (ftmTx *BridgeAURORATxObject) SetValue(data interface{}) error {
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
	ftmTx.bridgeAURORATxState = newBridgeAURORATxState
	return nil
}

func (ftmTx BridgeAURORATxObject) GetValue() interface{} {
	return ftmTx.bridgeAURORATxState
}

func (ftmTx BridgeAURORATxObject) GetValueBytes() []byte {
	data := ftmTx.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal bridge BSC tx state")
	}
	return value
}

func (ftmTx BridgeAURORATxObject) GetHash() common.Hash {
	return ftmTx.bridgeAURORATxHash
}

func (ftmTx BridgeAURORATxObject) GetType() int {
	return ftmTx.objectType
}

// MarkDelete will delete an object in trie
func (ftmTx *BridgeAURORATxObject) MarkDelete() {
	ftmTx.deleted = true
}

func (ftmTx *BridgeAURORATxObject) Reset() bool {
	ftmTx.bridgeAURORATxState = NewBridgeAURORATxState()
	return true
}

func (ftmTx BridgeAURORATxObject) IsDeleted() bool {
	return ftmTx.deleted
}

// value is either default or nil
func (ftmTx BridgeAURORATxObject) IsEmpty() bool {
	temp := NewBridgeAURORATxState()
	return reflect.DeepEqual(temp, ftmTx.bridgeAURORATxState) || ftmTx.bridgeAURORATxState == nil
}
