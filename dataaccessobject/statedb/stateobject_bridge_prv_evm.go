package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BrigePRVEVMState struct {
	uniquePRVEVMTx []byte
}

func (prvEvmTx BrigePRVEVMState) UniquePRVEVMTx() []byte {
	return prvEvmTx.uniquePRVEVMTx
}

func (prvEvmTx *BrigePRVEVMState) SetuniquePRVEVMTx(uniquePRVEVMTx []byte) {
	prvEvmTx.uniquePRVEVMTx = uniquePRVEVMTx
}

func NewBrigePRVEVMState() *BrigePRVEVMState {
	return &BrigePRVEVMState{}
}

func NewBrigePRVEVMStateWithValue(uniquePRVEVMTx []byte) *BrigePRVEVMState {
	return &BrigePRVEVMState{uniquePRVEVMTx: uniquePRVEVMTx}
}

func (prvEvmTx BrigePRVEVMState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		UniquePRVEVMTx []byte
	}{
		UniquePRVEVMTx: prvEvmTx.uniquePRVEVMTx,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (prvEvmTx *BrigePRVEVMState) UnmarshalJSON(data []byte) error {
	temp := struct {
		UniquePRVEVMTx []byte
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	prvEvmTx.uniquePRVEVMTx = temp.UniquePRVEVMTx
	return nil
}

type BrigePRVEVMObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version          int
	BrigePRVEVMHash  common.Hash
	BrigePRVEVMState *BrigePRVEVMState
	objectType       int
	deleted          bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBrigePRVEVMObject(db *StateDB, hash common.Hash) *BrigePRVEVMObject {
	return &BrigePRVEVMObject{
		version:          defaultVersion,
		db:               db,
		BrigePRVEVMHash:  hash,
		BrigePRVEVMState: NewBrigePRVEVMState(),
		objectType:       BridgePRVEVMObjectType,
		deleted:          false,
	}
}

func newBrigePRVEVMObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BrigePRVEVMObject, error) {
	var newBrigePRVEVMState = NewBrigePRVEVMState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBrigePRVEVMState)
		if err != nil {
			return nil, err
		}
	} else {
		newBrigePRVEVMState, ok = data.(*BrigePRVEVMState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgePRVEVMStateType, reflect.TypeOf(data))
		}
	}
	return &BrigePRVEVMObject{
		version:          defaultVersion,
		BrigePRVEVMHash:  key,
		BrigePRVEVMState: newBrigePRVEVMState,
		db:               db,
		objectType:       BridgePRVEVMObjectType,
		deleted:          false,
	}, nil
}

func GenerateBrigePRVEVMObjectKey(uniquePRVEVMTx []byte) common.Hash {
	prefixHash := GetBridgePRVEVMPrefix()
	valueHash := common.HashH(uniquePRVEVMTx)
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (prvEvmTx BrigePRVEVMObject) GetVersion() int {
	return prvEvmTx.version
}

// setError remembers the first non-nil error it is called with.
func (prvEvmTx *BrigePRVEVMObject) SetError(err error) {
	if prvEvmTx.dbErr == nil {
		prvEvmTx.dbErr = err
	}
}

func (prvEvmTx BrigePRVEVMObject) GetTrie(db DatabaseAccessWarper) Trie {
	return prvEvmTx.trie
}

func (prvEvmTx *BrigePRVEVMObject) SetValue(data interface{}) error {
	var newBrigePRVEVMState = NewBrigePRVEVMState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBrigePRVEVMState)
		if err != nil {
			return err
		}
	} else {
		newBrigePRVEVMState, ok = data.(*BrigePRVEVMState)
		if !ok {
			return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgePRVEVMStateType, reflect.TypeOf(data))
		}
	}
	prvEvmTx.BrigePRVEVMState = newBrigePRVEVMState
	return nil
}

func (prvEvmTx BrigePRVEVMObject) GetValue() interface{} {
	return prvEvmTx.BrigePRVEVMState
}

func (prvEvmTx BrigePRVEVMObject) GetValueBytes() []byte {
	data := prvEvmTx.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal bridge PRV EVM tx state")
	}
	return []byte(value)
}

func (prvEvmTx BrigePRVEVMObject) GetHash() common.Hash {
	return prvEvmTx.BrigePRVEVMHash
}

func (prvEvmTx BrigePRVEVMObject) GetType() int {
	return prvEvmTx.objectType
}

// MarkDelete will delete an object in trie
func (prvEvmTx *BrigePRVEVMObject) MarkDelete() {
	prvEvmTx.deleted = true
}

func (prvEvmTx *BrigePRVEVMObject) Reset() bool {
	prvEvmTx.BrigePRVEVMState = NewBrigePRVEVMState()
	return true
}

func (prvEvmTx BrigePRVEVMObject) IsDeleted() bool {
	return prvEvmTx.deleted
}

// value is either default or nil
func (prvEvmTx BrigePRVEVMObject) IsEmpty() bool {
	temp := NewBrigePRVEVMState()
	return reflect.DeepEqual(temp, prvEvmTx.BrigePRVEVMState) || prvEvmTx.BrigePRVEVMState == nil
}
