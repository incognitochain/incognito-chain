package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BridgePDEXEVMState struct {
	uniquePDEXEVMTx []byte
}

func (pDexEvmTx BridgePDEXEVMState) UniquePDEXEVMTx() []byte {
	return pDexEvmTx.uniquePDEXEVMTx
}

func (pDexEvmTx *BridgePDEXEVMState) SetUniquePDEXEVMTx(uniquePDEXEVMTx []byte) {
	pDexEvmTx.uniquePDEXEVMTx = uniquePDEXEVMTx
}

func NewBridgePDEXEVMState() *BridgePDEXEVMState {
	return &BridgePDEXEVMState{}
}

func NewBridgePDEXEVMStateWithValue(uniquePDEXEVMTx []byte) *BridgePDEXEVMState {
	return &BridgePDEXEVMState{uniquePDEXEVMTx: uniquePDEXEVMTx}
}

func (pDexEvmTx BridgePDEXEVMState) MarshalJSON() ([]byte, error) {
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

func (pDexEvmTx *BridgePDEXEVMState) UnmarshalJSON(data []byte) error {
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

type BridgePDEXEVMObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version           int
	BrigePDEXEVMHash  common.Hash
	BrigePDEXEVMState *BridgePDEXEVMState
	objectType        int
	deleted           bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBridgePDEXEVMObject(db *StateDB, hash common.Hash) *BridgePDEXEVMObject {
	return &BridgePDEXEVMObject{
		version:           defaultVersion,
		db:                db,
		BrigePDEXEVMHash:  hash,
		BrigePDEXEVMState: NewBridgePDEXEVMState(),
		objectType:        BridgePDEXEVMObjectType,
		deleted:           false,
	}
}

func newBridgePDEXEVMObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BridgePDEXEVMObject, error) {
	var newBrigePDEXEVMState = NewBridgePDEXEVMState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBrigePDEXEVMState)
		if err != nil {
			return nil, err
		}
	} else {
		newBrigePDEXEVMState, ok = data.(*BridgePDEXEVMState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgePDEXEVMStateType, reflect.TypeOf(data))
		}
	}
	return &BridgePDEXEVMObject{
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

func (pDexEvmTx BridgePDEXEVMObject) GetVersion() int {
	return pDexEvmTx.version
}

// setError remembers the first non-nil error it is called with.
func (pDexEvmTx *BridgePDEXEVMObject) SetError(err error) {
	if pDexEvmTx.dbErr == nil {
		pDexEvmTx.dbErr = err
	}
}

func (pDexEvmTx BridgePDEXEVMObject) GetTrie(db DatabaseAccessWarper) Trie {
	return pDexEvmTx.trie
}

func (pDexEvmTx *BridgePDEXEVMObject) SetValue(data interface{}) error {
	var newBrigePDEXEVMState = NewBridgePDEXEVMState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBrigePDEXEVMState)
		if err != nil {
			return err
		}
	} else {
		newBrigePDEXEVMState, ok = data.(*BridgePDEXEVMState)
		if !ok {
			return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgePDEXEVMStateType, reflect.TypeOf(data))
		}
	}
	pDexEvmTx.BrigePDEXEVMState = newBrigePDEXEVMState
	return nil
}

func (pDexEvmTx BridgePDEXEVMObject) GetValue() interface{} {
	return pDexEvmTx.BrigePDEXEVMState
}

func (pDexEvmTx BridgePDEXEVMObject) GetValueBytes() []byte {
	data := pDexEvmTx.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal bridge PRV EVM tx state")
	}
	return []byte(value)
}

func (pDexEvmTx BridgePDEXEVMObject) GetHash() common.Hash {
	return pDexEvmTx.BrigePDEXEVMHash
}

func (pDexEvmTx BridgePDEXEVMObject) GetType() int {
	return pDexEvmTx.objectType
}

// MarkDelete will delete an object in trie
func (pDexEvmTx *BridgePDEXEVMObject) MarkDelete() {
	pDexEvmTx.deleted = true
}

func (pDexEvmTx *BridgePDEXEVMObject) Reset() bool {
	pDexEvmTx.BrigePDEXEVMState = NewBridgePDEXEVMState()
	return true
}

func (pDexEvmTx BridgePDEXEVMObject) IsDeleted() bool {
	return pDexEvmTx.deleted
}

// value is either default or nil
func (pDexEvmTx BridgePDEXEVMObject) IsEmpty() bool {
	temp := NewBridgePDEXEVMState()
	return reflect.DeepEqual(temp, pDexEvmTx.BrigePDEXEVMState) || pDexEvmTx.BrigePDEXEVMState == nil
}
