package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BrigePDEXEVMState struct {
	uniquePDEXEVMTx []byte
}

func (pDexEvmTx BrigePDEXEVMState) UniquePDEXEVMTx() []byte {
	return pDexEvmTx.uniquePDEXEVMTx
}

func (pDexEvmTx *BrigePDEXEVMState) SetUniquePDEXEVMTx(uniquePDEXEVMTx []byte) {
	pDexEvmTx.uniquePDEXEVMTx = uniquePDEXEVMTx
}

func NewBrigePDEXEVMState() *BrigePDEXEVMState {
	return &BrigePDEXEVMState{}
}

func NewBrigePDEXEVMStateWithValue(uniquePDEXEVMTx []byte) *BrigePDEXEVMState {
	return &BrigePDEXEVMState{uniquePDEXEVMTx: uniquePDEXEVMTx}
}

func (pDexEvmTx BrigePDEXEVMState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		UniquePDEXEVMTx []byte
	}{
		UniquePDEXEVMTx: pDexEvmTx.uniquePDEXEVMTx,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (pDexEvmTx *BrigePDEXEVMState) UnmarshalJSON(data []byte) error {
	temp := struct {
		UniquePDEXEVMTx []byte
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	pDexEvmTx.uniquePDEXEVMTx = temp.UniquePDEXEVMTx
	return nil
}

type BrigePDEXEVMObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version           int
	BrigePDEXEVMHash  common.Hash
	BrigePDEXEVMState *BrigePDEXEVMState
	objectType        int
	deleted           bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBrigePDEXEVMObject(db *StateDB, hash common.Hash) *BrigePDEXEVMObject {
	return &BrigePDEXEVMObject{
		version:           defaultVersion,
		db:                db,
		BrigePDEXEVMHash:  hash,
		BrigePDEXEVMState: NewBrigePDEXEVMState(),
		objectType:        BridgePDEXEVMObjectType,
		deleted:           false,
	}
}

func newBrigePDEXEVMObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BrigePDEXEVMObject, error) {
	var newBrigePDEXEVMState = NewBrigePDEXEVMState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBrigePDEXEVMState)
		if err != nil {
			return nil, err
		}
	} else {
		newBrigePDEXEVMState, ok = data.(*BrigePDEXEVMState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgePDEXEVMStateType, reflect.TypeOf(data))
		}
	}
	return &BrigePDEXEVMObject{
		version:           defaultVersion,
		BrigePDEXEVMHash:  key,
		BrigePDEXEVMState: newBrigePDEXEVMState,
		db:                db,
		objectType:        BridgePDEXEVMObjectType,
		deleted:           false,
	}, nil
}

func GenerateBridgePDEXEVMObjectKey(uniquePDEXEVMTx []byte) common.Hash {
	prefixHash := GetBridgePDEXEVMPrefix()
	valueHash := common.HashH(uniquePDEXEVMTx)
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (pDexEvmTx BrigePDEXEVMObject) GetVersion() int {
	return pDexEvmTx.version
}

// setError remembers the first non-nil error it is called with.
func (pDexEvmTx *BrigePDEXEVMObject) SetError(err error) {
	if pDexEvmTx.dbErr == nil {
		pDexEvmTx.dbErr = err
	}
}

func (pDexEvmTx BrigePDEXEVMObject) GetTrie(db DatabaseAccessWarper) Trie {
	return pDexEvmTx.trie
}

func (pDexEvmTx *BrigePDEXEVMObject) SetValue(data interface{}) error {
	var newBrigePDEXEVMState = NewBrigePDEXEVMState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBrigePDEXEVMState)
		if err != nil {
			return err
		}
	} else {
		newBrigePDEXEVMState, ok = data.(*BrigePDEXEVMState)
		if !ok {
			return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgePDEXEVMStateType, reflect.TypeOf(data))
		}
	}
	pDexEvmTx.BrigePDEXEVMState = newBrigePDEXEVMState
	return nil
}

func (pDexEvmTx BrigePDEXEVMObject) GetValue() interface{} {
	return pDexEvmTx.BrigePDEXEVMState
}

func (pDexEvmTx BrigePDEXEVMObject) GetValueBytes() []byte {
	data := pDexEvmTx.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal bridge PRV EVM tx state")
	}
	return []byte(value)
}

func (pDexEvmTx BrigePDEXEVMObject) GetHash() common.Hash {
	return pDexEvmTx.BrigePDEXEVMHash
}

func (pDexEvmTx BrigePDEXEVMObject) GetType() int {
	return pDexEvmTx.objectType
}

// MarkDelete will delete an object in trie
func (pDexEvmTx *BrigePDEXEVMObject) MarkDelete() {
	pDexEvmTx.deleted = true
}

func (pDexEvmTx *BrigePDEXEVMObject) Reset() bool {
	pDexEvmTx.BrigePDEXEVMState = NewBrigePDEXEVMState()
	return true
}

func (pDexEvmTx BrigePDEXEVMObject) IsDeleted() bool {
	return pDexEvmTx.deleted
}

// value is either default or nil
func (pDexEvmTx BrigePDEXEVMObject) IsEmpty() bool {
	temp := NewBrigePDEXEVMState()
	return reflect.DeepEqual(temp, pDexEvmTx.BrigePDEXEVMState) || pDexEvmTx.BrigePDEXEVMState == nil
}
