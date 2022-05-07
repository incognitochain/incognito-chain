package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BridgeLUNTxState struct {
	uniqueLUNTx []byte
}

func (ftmTx BridgeLUNTxState) UniqueLUNTx() []byte {
	return ftmTx.uniqueLUNTx
}

func (ftmTx *BridgeLUNTxState) SetUniqueLUNTx(uniqueLUNTx []byte) {
	ftmTx.uniqueLUNTx = uniqueLUNTx
}

//
func NewBridgeLUNTxState() *BridgeLUNTxState {
	return &BridgeLUNTxState{}
}

func NewBridgeLUNTxStateWithValue(uniqueLUNTx []byte) *BridgeLUNTxState {
	return &BridgeLUNTxState{uniqueLUNTx: uniqueLUNTx}
}

func (ftmTx BridgeLUNTxState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		UniqueLUNTx []byte
	}{
		UniqueLUNTx: ftmTx.uniqueLUNTx,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (ftmTx *BridgeLUNTxState) UnmarshalJSON(data []byte) error {
	temp := struct {
		UniqueLUNTx []byte
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	ftmTx.uniqueLUNTx = temp.UniqueLUNTx
	return nil
}

type BridgeLUNTxObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version          int
	bridgeLUNTxHash  common.Hash
	bridgeLUNTxState *BridgeLUNTxState
	objectType       int
	deleted          bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBridgeLUNTxObject(db *StateDB, hash common.Hash) *BridgeLUNTxObject {
	return &BridgeLUNTxObject{
		version:          defaultVersion,
		db:               db,
		bridgeLUNTxHash:  hash,
		bridgeLUNTxState: NewBridgeLUNTxState(),
		objectType:       BridgeLUNTxObjectType,
		deleted:          false,
	}
}

func newBridgeLUNTxObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BridgeLUNTxObject, error) {
	var newBridgeLUNTxState = NewBridgeLUNTxState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeLUNTxState)
		if err != nil {
			return nil, err
		}
	} else {
		newBridgeLUNTxState, ok = data.(*BridgeLUNTxState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeLUNTxStateType, reflect.TypeOf(data))
		}
	}
	return &BridgeLUNTxObject{
		version:          defaultVersion,
		bridgeLUNTxHash:  key,
		bridgeLUNTxState: newBridgeLUNTxState,
		db:               db,
		objectType:       BridgeLUNTxObjectType,
		deleted:          false,
	}, nil
}

func GenerateBridgeLUNTxObjectKey(uniqueLUNTx []byte) common.Hash {
	prefixHash := GetBridgeLUNTxPrefix()
	valueHash := common.HashH(uniqueLUNTx)
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (ftmTx BridgeLUNTxObject) GetVersion() int {
	return ftmTx.version
}

// setError remembers the first non-nil error it is called with.
func (ftmTx *BridgeLUNTxObject) SetError(err error) {
	if ftmTx.dbErr == nil {
		ftmTx.dbErr = err
	}
}

func (ftmTx BridgeLUNTxObject) GetTrie(db DatabaseAccessWarper) Trie {
	return ftmTx.trie
}

func (ftmTx *BridgeLUNTxObject) SetValue(data interface{}) error {
	var newBridgeLUNTxState = NewBridgeLUNTxState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeLUNTxState)
		if err != nil {
			return err
		}
	} else {
		newBridgeLUNTxState, ok = data.(*BridgeLUNTxState)
		if !ok {
			return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeLUNTxStateType, reflect.TypeOf(data))
		}
	}
	ftmTx.bridgeLUNTxState = newBridgeLUNTxState
	return nil
}

func (ftmTx BridgeLUNTxObject) GetValue() interface{} {
	return ftmTx.bridgeLUNTxState
}

func (ftmTx BridgeLUNTxObject) GetValueBytes() []byte {
	data := ftmTx.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal bridge BSC tx state")
	}
	return value
}

func (ftmTx BridgeLUNTxObject) GetHash() common.Hash {
	return ftmTx.bridgeLUNTxHash
}

func (ftmTx BridgeLUNTxObject) GetType() int {
	return ftmTx.objectType
}

// MarkDelete will delete an object in trie
func (ftmTx *BridgeLUNTxObject) MarkDelete() {
	ftmTx.deleted = true
}

func (ftmTx *BridgeLUNTxObject) Reset() bool {
	ftmTx.bridgeLUNTxState = NewBridgeLUNTxState()
	return true
}

func (ftmTx BridgeLUNTxObject) IsDeleted() bool {
	return ftmTx.deleted
}

// value is either default or nil
func (ftmTx BridgeLUNTxObject) IsEmpty() bool {
	temp := NewBridgeLUNTxState()
	return reflect.DeepEqual(temp, ftmTx.bridgeLUNTxState) || ftmTx.bridgeLUNTxState == nil
}
