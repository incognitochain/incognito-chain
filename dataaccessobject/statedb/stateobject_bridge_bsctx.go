package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BridgeBSCTxState struct {
	uniqueBSCTx []byte
}

func (BSCtx BridgeBSCTxState) UniqueBSCTx() []byte {
	return BSCtx.uniqueBSCTx
}

func (BSCtx *BridgeBSCTxState) SetUniqueBSCTx(uniqueBSCTx []byte) {
	BSCtx.uniqueBSCTx = uniqueBSCTx
}

func NewBridgeBSCTxState() *BridgeBSCTxState {
	return &BridgeBSCTxState{}
}

func NewBridgeBSCTxStateWithValue(uniqueBSCTx []byte) *BridgeBSCTxState {
	return &BridgeBSCTxState{uniqueBSCTx: uniqueBSCTx}
}

func (BSCtx BridgeBSCTxState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		UniqueBSCTx []byte
	}{
		UniqueBSCTx: BSCtx.uniqueBSCTx,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (BSCtx *BridgeBSCTxState) UnmarshalJSON(data []byte) error {
	temp := struct {
		UniqueBSCTx []byte
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	BSCtx.uniqueBSCTx = temp.UniqueBSCTx
	return nil
}

type BridgeBSCTxObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version          int
	bridgeBSCTxHash  common.Hash
	bridgeBSCTxState *BridgeBSCTxState
	objectType       int
	deleted          bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBridgeBSCTxObject(db *StateDB, hash common.Hash) *BridgeBSCTxObject {
	return &BridgeBSCTxObject{
		version:          defaultVersion,
		db:               db,
		bridgeBSCTxHash:  hash,
		bridgeBSCTxState: NewBridgeBSCTxState(),
		objectType:       BridgeBSCTxObjectType,
		deleted:          false,
	}
}

func newBridgeBSCTxObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BridgeBSCTxObject, error) {
	var newBridgeBSCTxState = NewBridgeBSCTxState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeBSCTxState)
		if err != nil {
			return nil, err
		}
	} else {
		newBridgeBSCTxState, ok = data.(*BridgeBSCTxState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeBSCTxStateType, reflect.TypeOf(data))
		}
	}
	return &BridgeBSCTxObject{
		version:          defaultVersion,
		bridgeBSCTxHash:  key,
		bridgeBSCTxState: newBridgeBSCTxState,
		db:               db,
		objectType:       BridgeBSCTxObjectType,
		deleted:          false,
	}, nil
}

func GenerateBridgeBSCTxObjectKey(uniqueBSCTx []byte) common.Hash {
	prefixHash := GetBridgeBSCTxPrefix()
	valueHash := common.HashH(uniqueBSCTx)
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (BSCtx BridgeBSCTxObject) GetVersion() int {
	return BSCtx.version
}

// setError remembers the first non-nil error it is called with.
func (BSCtx *BridgeBSCTxObject) SetError(err error) {
	if BSCtx.dbErr == nil {
		BSCtx.dbErr = err
	}
}

func (BSCtx BridgeBSCTxObject) GetTrie(db DatabaseAccessWarper) Trie {
	return BSCtx.trie
}

func (BSCtx *BridgeBSCTxObject) SetValue(data interface{}) error {
	var newBridgeBSCTxState = NewBridgeBSCTxState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeBSCTxState)
		if err != nil {
			return err
		}
	} else {
		newBridgeBSCTxState, ok = data.(*BridgeBSCTxState)
		if !ok {
			return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeBSCTxStateType, reflect.TypeOf(data))
		}
	}
	BSCtx.bridgeBSCTxState = newBridgeBSCTxState
	return nil
}

func (BSCtx BridgeBSCTxObject) GetValue() interface{} {
	return BSCtx.bridgeBSCTxState
}

func (BSCtx BridgeBSCTxObject) GetValueBytes() []byte {
	data := BSCtx.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal bridge BSC tx state")
	}
	return value
}

func (BSCtx BridgeBSCTxObject) GetHash() common.Hash {
	return BSCtx.bridgeBSCTxHash
}

func (BSCtx BridgeBSCTxObject) GetType() int {
	return BSCtx.objectType
}

// MarkDelete will delete an object in trie
func (BSCtx *BridgeBSCTxObject) MarkDelete() {
	BSCtx.deleted = true
}

func (BSCtx *BridgeBSCTxObject) Reset() bool {
	BSCtx.bridgeBSCTxState = NewBridgeBSCTxState()
	return true
}

func (BSCtx BridgeBSCTxObject) IsDeleted() bool {
	return BSCtx.deleted
}

// value is either default or nil
func (BSCtx BridgeBSCTxObject) IsEmpty() bool {
	temp := NewBridgeBSCTxState()
	return reflect.DeepEqual(temp, BSCtx.bridgeBSCTxState) || BSCtx.bridgeBSCTxState == nil
}
