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

func (nearTx BridgeNEARTxState) UniqueNEARTx() []byte {
	return nearTx.uniqueNEARTx
}

func (nearTx *BridgeNEARTxState) SetUniqueNEARTx(uniqueNEARTx []byte) {
	nearTx.uniqueNEARTx = uniqueNEARTx
}

func NewBridgeNEARTxState() *BridgeNEARTxState {
	return &BridgeNEARTxState{}
}

func NewBridgeNEARTxStateWithValue(uniqueNEARTx []byte) *BridgeNEARTxState {
	return &BridgeNEARTxState{uniqueNEARTx: uniqueNEARTx}
}

func (nearTx BridgeNEARTxState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		UniqueNEARTx []byte
	}{
		UniqueNEARTx: nearTx.uniqueNEARTx,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (nearTx *BridgeNEARTxState) UnmarshalJSON(data []byte) error {
	temp := struct {
		UniqueNEARTx []byte
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	nearTx.uniqueNEARTx = temp.UniqueNEARTx
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

func (nearTx BridgeNEARTxObject) GetVersion() int {
	return nearTx.version
}

// setError remembers the first non-nil error it is called with.
func (nearTx *BridgeNEARTxObject) SetError(err error) {
	if nearTx.dbErr == nil {
		nearTx.dbErr = err
	}
}

func (nearTx BridgeNEARTxObject) GetTrie(db DatabaseAccessWarper) Trie {
	return nearTx.trie
}

func (nearTx *BridgeNEARTxObject) SetValue(data interface{}) error {
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
	nearTx.bridgeNEARTxState = newBridgeNEARTxState
	return nil
}

func (nearTx BridgeNEARTxObject) GetValue() interface{} {
	return nearTx.bridgeNEARTxState
}

func (nearTx BridgeNEARTxObject) GetValueBytes() []byte {
	data := nearTx.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal bridge BSC tx state")
	}
	return value
}

func (nearTx BridgeNEARTxObject) GetHash() common.Hash {
	return nearTx.bridgeNEARTxHash
}

func (nearTx BridgeNEARTxObject) GetType() int {
	return nearTx.objectType
}

// MarkDelete will delete an object in trie
func (nearTx *BridgeNEARTxObject) MarkDelete() {
	nearTx.deleted = true
}

func (nearTx *BridgeNEARTxObject) Reset() bool {
	nearTx.bridgeNEARTxState = NewBridgeNEARTxState()
	return true
}

func (nearTx BridgeNEARTxObject) IsDeleted() bool {
	return nearTx.deleted
}

// value is either default or nil
func (nearTx BridgeNEARTxObject) IsEmpty() bool {
	temp := NewBridgeNEARTxState()
	return reflect.DeepEqual(temp, nearTx.bridgeNEARTxState) || nearTx.bridgeNEARTxState == nil
}
