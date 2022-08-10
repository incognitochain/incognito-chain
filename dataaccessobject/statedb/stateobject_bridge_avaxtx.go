package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BridgeAVAXTxState struct {
	uniqueAVAXTx []byte
}

func (ftmTx BridgeAVAXTxState) UniqueAVAXTx() []byte {
	return ftmTx.uniqueAVAXTx
}

func (ftmTx *BridgeAVAXTxState) SetUniqueAVAXTx(uniqueAVAXTx []byte) {
	ftmTx.uniqueAVAXTx = uniqueAVAXTx
}

func NewBridgeAVAXTxState() *BridgeAVAXTxState {
	return &BridgeAVAXTxState{}
}

func NewBridgeAVAXTxStateWithValue(uniqueAVAXTx []byte) *BridgeAVAXTxState {
	return &BridgeAVAXTxState{uniqueAVAXTx: uniqueAVAXTx}
}

func (ftmTx BridgeAVAXTxState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		UniqueAVAXTx []byte
	}{
		UniqueAVAXTx: ftmTx.uniqueAVAXTx,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (ftmTx *BridgeAVAXTxState) UnmarshalJSON(data []byte) error {
	temp := struct {
		UniqueAVAXTx []byte
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	ftmTx.uniqueAVAXTx = temp.UniqueAVAXTx
	return nil
}

type BridgeAVAXTxObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version           int
	bridgeAVAXTxHash  common.Hash
	bridgeAVAXTxState *BridgeAVAXTxState
	objectType        int
	deleted           bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBridgeAVAXTxObject(db *StateDB, hash common.Hash) *BridgeAVAXTxObject {
	return &BridgeAVAXTxObject{
		version:           defaultVersion,
		db:                db,
		bridgeAVAXTxHash:  hash,
		bridgeAVAXTxState: NewBridgeAVAXTxState(),
		objectType:        BridgeAVAXTxObjectType,
		deleted:           false,
	}
}

func newBridgeAVAXTxObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BridgeAVAXTxObject, error) {
	var newBridgeAVAXTxState = NewBridgeAVAXTxState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeAVAXTxState)
		if err != nil {
			return nil, err
		}
	} else {
		newBridgeAVAXTxState, ok = data.(*BridgeAVAXTxState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeAVAXTxStateType, reflect.TypeOf(data))
		}
	}
	return &BridgeAVAXTxObject{
		version:           defaultVersion,
		bridgeAVAXTxHash:  key,
		bridgeAVAXTxState: newBridgeAVAXTxState,
		db:                db,
		objectType:        BridgeAVAXTxObjectType,
		deleted:           false,
	}, nil
}

func GenerateBridgeAVAXTxObjectKey(uniqueAVAXTx []byte) common.Hash {
	prefixHash := GetBridgeAVAXTxPrefix()
	valueHash := common.HashH(uniqueAVAXTx)
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (ftmTx BridgeAVAXTxObject) GetVersion() int {
	return ftmTx.version
}

// setError remembers the first non-nil error it is called with.
func (ftmTx *BridgeAVAXTxObject) SetError(err error) {
	if ftmTx.dbErr == nil {
		ftmTx.dbErr = err
	}
}

func (ftmTx BridgeAVAXTxObject) GetTrie(db DatabaseAccessWarper) Trie {
	return ftmTx.trie
}

func (ftmTx *BridgeAVAXTxObject) SetValue(data interface{}) error {
	var newBridgeAVAXTxState = NewBridgeAVAXTxState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeAVAXTxState)
		if err != nil {
			return err
		}
	} else {
		newBridgeAVAXTxState, ok = data.(*BridgeAVAXTxState)
		if !ok {
			return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeAVAXTxStateType, reflect.TypeOf(data))
		}
	}
	ftmTx.bridgeAVAXTxState = newBridgeAVAXTxState
	return nil
}

func (ftmTx BridgeAVAXTxObject) GetValue() interface{} {
	return ftmTx.bridgeAVAXTxState
}

func (ftmTx BridgeAVAXTxObject) GetValueBytes() []byte {
	data := ftmTx.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal bridge BSC tx state")
	}
	return value
}

func (ftmTx BridgeAVAXTxObject) GetHash() common.Hash {
	return ftmTx.bridgeAVAXTxHash
}

func (ftmTx BridgeAVAXTxObject) GetType() int {
	return ftmTx.objectType
}

// MarkDelete will delete an object in trie
func (ftmTx *BridgeAVAXTxObject) MarkDelete() {
	ftmTx.deleted = true
}

func (ftmTx *BridgeAVAXTxObject) Reset() bool {
	ftmTx.bridgeAVAXTxState = NewBridgeAVAXTxState()
	return true
}

func (ftmTx BridgeAVAXTxObject) IsDeleted() bool {
	return ftmTx.deleted
}

// value is either default or nil
func (ftmTx BridgeAVAXTxObject) IsEmpty() bool {
	temp := NewBridgeAVAXTxState()
	return reflect.DeepEqual(temp, ftmTx.bridgeAVAXTxState) || ftmTx.bridgeAVAXTxState == nil
}
