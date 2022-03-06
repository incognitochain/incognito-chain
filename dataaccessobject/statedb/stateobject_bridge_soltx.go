package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BridgeSOLTxState struct {
	uniqueSOLTx []byte
}

func (solTx BridgeSOLTxState) UniqueSOLTx() []byte {
	return solTx.uniqueSOLTx
}

func (solTx *BridgeSOLTxState) SetUniqueSOLTx(uniqueSOLTx []byte) {
	solTx.uniqueSOLTx = uniqueSOLTx
}

func NewBridgeSOLTxState() *BridgeSOLTxState {
	return &BridgeSOLTxState{}
}

func NewBridgeSOLTxStateWithValue(uniqueSOLTx []byte) *BridgeSOLTxState {
	return &BridgeSOLTxState{uniqueSOLTx: uniqueSOLTx}
}

func (solTx BridgeSOLTxState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		UniqueSOLTx []byte
	}{
		UniqueSOLTx: solTx.uniqueSOLTx,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (solTx *BridgeSOLTxState) UnmarshalJSON(data []byte) error {
	temp := struct {
		UniqueSOLTx []byte
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	solTx.uniqueSOLTx = temp.UniqueSOLTx
	return nil
}

type BridgeSOLTxObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version          int
	bridgeSOLTxHash  common.Hash
	BridgeSOLTxState *BridgeSOLTxState
	objectType       int
	deleted          bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBridgeSOLTxObject(db *StateDB, hash common.Hash) *BridgeSOLTxObject {
	return &BridgeSOLTxObject{
		version:          defaultVersion,
		db:               db,
		bridgeSOLTxHash:  hash,
		BridgeSOLTxState: NewBridgeSOLTxState(),
		objectType:       BridgeSOLTxObjectType,
		deleted:          false,
	}
}

func newBridgeSOLTxObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BridgeSOLTxObject, error) {
	var newBridgeSOLTxState = NewBridgeSOLTxState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeSOLTxState)
		if err != nil {
			return nil, err
		}
	} else {
		newBridgeSOLTxState, ok = data.(*BridgeSOLTxState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeSOLTxStateType, reflect.TypeOf(data))
		}
	}
	return &BridgeSOLTxObject{
		version:          defaultVersion,
		bridgeSOLTxHash:  key,
		BridgeSOLTxState: newBridgeSOLTxState,
		db:               db,
		objectType:       BridgeSOLTxObjectType,
		deleted:          false,
	}, nil
}

func GenerateBridgeSOLTxObjectKey(uniqueSOLTx []byte) common.Hash {
	prefixHash := GetBridgeSOLTxPrefix()
	valueHash := common.HashH(uniqueSOLTx)
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (solTx BridgeSOLTxObject) GetVersion() int {
	return solTx.version
}

// setError remembers the first non-nil error it is called with.
func (solTx *BridgeSOLTxObject) SetError(err error) {
	if solTx.dbErr == nil {
		solTx.dbErr = err
	}
}

func (solTx BridgeSOLTxObject) GetTrie(db DatabaseAccessWarper) Trie {
	return solTx.trie
}

func (solTx *BridgeSOLTxObject) SetValue(data interface{}) error {
	var newBridgeSOLTxState = NewBridgeSOLTxState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeSOLTxState)
		if err != nil {
			return err
		}
	} else {
		newBridgeSOLTxState, ok = data.(*BridgeSOLTxState)
		if !ok {
			return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeSOLTxStateType, reflect.TypeOf(data))
		}
	}
	solTx.BridgeSOLTxState = newBridgeSOLTxState
	return nil
}

func (solTx BridgeSOLTxObject) GetValue() interface{} {
	return solTx.BridgeSOLTxState
}

func (solTx BridgeSOLTxObject) GetValueBytes() []byte {
	data := solTx.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal bridge SOL tx state")
	}
	return value
}

func (solTx BridgeSOLTxObject) GetHash() common.Hash {
	return solTx.bridgeSOLTxHash
}

func (solTx BridgeSOLTxObject) GetType() int {
	return solTx.objectType
}

// MarkDelete will delete an object in trie
func (solTx *BridgeSOLTxObject) MarkDelete() {
	solTx.deleted = true
}

func (solTx *BridgeSOLTxObject) Reset() bool {
	solTx.BridgeSOLTxState = NewBridgeSOLTxState()
	return true
}

func (solTx BridgeSOLTxObject) IsDeleted() bool {
	return solTx.deleted
}

// value is either default or nil
func (solTx BridgeSOLTxObject) IsEmpty() bool {
	temp := NewBridgeSOLTxState()
	return reflect.DeepEqual(temp, solTx.BridgeSOLTxState) || solTx.BridgeSOLTxState == nil
}
