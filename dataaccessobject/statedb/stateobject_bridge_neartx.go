package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BridgeNEARTxState struct {
	uniqueNEARTx []byte
}

func (ftmTx BridgeNEARTxState) UniqueNEARTx() []byte {
	return ftmTx.uniqueNEARTx
}

func (ftmTx *BridgeNEARTxState) SetUniqueNEARTx(uniqueNEARTx []byte) {
	ftmTx.uniqueNEARTx = uniqueNEARTx
}

func NewBridgeNEARTxState() *BridgeNEARTxState {
	return &BridgeNEARTxState{}
}

func NewBridgeNEARTxStateWithValue(uniqueNEARTx []byte) *BridgeNEARTxState {
	return &BridgeNEARTxState{uniqueNEARTx: uniqueNEARTx}
}

func (ftmTx BridgeNEARTxState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		UniqueNEARTx []byte
	}{
		UniqueNEARTx: ftmTx.uniqueNEARTx,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (ftmTx *BridgeNEARTxState) UnmarshalJSON(data []byte) error {
	temp := struct {
		UniqueNEARTx []byte
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	ftmTx.uniqueNEARTx = temp.UniqueNEARTx
	return nil
}

type BridgeNEARTxObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version           int
	bridgeNEARTxHash  common.Hash
	bridgeNEARTxState *BridgeNEARTxState
	objectType        int
	deleted           bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBridgeNEARTxObject(db *StateDB, hash common.Hash) *BridgeNEARTxObject {
	return &BridgeNEARTxObject{
		version:           defaultVersion,
		db:                db,
		bridgeNEARTxHash:  hash,
		bridgeNEARTxState: NewBridgeNEARTxState(),
		objectType:        BridgeNEARTxObjectType,
		deleted:           false,
	}
}

func newBridgeNEARTxObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BridgeNEARTxObject, error) {
	var newBridgeNEARTxState = NewBridgeNEARTxState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeNEARTxState)
		if err != nil {
			return nil, err
		}
	} else {
		newBridgeNEARTxState, ok = data.(*BridgeNEARTxState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeNEARTxStateType, reflect.TypeOf(data))
		}
	}
	return &BridgeNEARTxObject{
		version:           defaultVersion,
		bridgeNEARTxHash:  key,
		bridgeNEARTxState: newBridgeNEARTxState,
		db:                db,
		objectType:        BridgeNEARTxObjectType,
		deleted:           false,
	}, nil
}

func GenerateBridgeNEARTxObjectKey(uniqueNEARTx []byte) common.Hash {
	prefixHash := GetBridgeNEARTxPrefix()
	valueHash := common.HashH(uniqueNEARTx)
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (ftmTx BridgeNEARTxObject) GetVersion() int {
	return ftmTx.version
}

// setError remembers the first non-nil error it is called with.
func (ftmTx *BridgeNEARTxObject) SetError(err error) {
	if ftmTx.dbErr == nil {
		ftmTx.dbErr = err
	}
}

func (ftmTx BridgeNEARTxObject) GetTrie(db DatabaseAccessWarper) Trie {
	return ftmTx.trie
}

func (ftmTx *BridgeNEARTxObject) SetValue(data interface{}) error {
	var newBridgeNEARTxState = NewBridgeNEARTxState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeNEARTxState)
		if err != nil {
			return err
		}
	} else {
		newBridgeNEARTxState, ok = data.(*BridgeNEARTxState)
		if !ok {
			return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeNEARTxStateType, reflect.TypeOf(data))
		}
	}
	ftmTx.bridgeNEARTxState = newBridgeNEARTxState
	return nil
}

func (ftmTx BridgeNEARTxObject) GetValue() interface{} {
	return ftmTx.bridgeNEARTxState
}

func (ftmTx BridgeNEARTxObject) GetValueBytes() []byte {
	data := ftmTx.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal bridge BSC tx state")
	}
	return value
}

func (ftmTx BridgeNEARTxObject) GetHash() common.Hash {
	return ftmTx.bridgeNEARTxHash
}

func (ftmTx BridgeNEARTxObject) GetType() int {
	return ftmTx.objectType
}

// MarkDelete will delete an object in trie
func (ftmTx *BridgeNEARTxObject) MarkDelete() {
	ftmTx.deleted = true
}

func (ftmTx *BridgeNEARTxObject) Reset() bool {
	ftmTx.bridgeNEARTxState = NewBridgeNEARTxState()
	return true
}

func (ftmTx BridgeNEARTxObject) IsDeleted() bool {
	return ftmTx.deleted
}

// value is either default or nil
func (ftmTx BridgeNEARTxObject) IsEmpty() bool {
	temp := NewBridgeNEARTxState()
	return reflect.DeepEqual(temp, ftmTx.bridgeNEARTxState) || ftmTx.bridgeNEARTxState == nil
}
