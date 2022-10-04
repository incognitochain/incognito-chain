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

func (avaxTx BridgeAVAXTxState) UniqueAVAXTx() []byte {
	return avaxTx.uniqueAVAXTx
}

func (avaxTx *BridgeAVAXTxState) SetUniqueAVAXTx(uniqueAVAXTx []byte) {
	avaxTx.uniqueAVAXTx = uniqueAVAXTx
}

func NewBridgeAVAXTxState() *BridgeAVAXTxState {
	return &BridgeAVAXTxState{}
}

func NewBridgeAVAXTxStateWithValue(uniqueAVAXTx []byte) *BridgeAVAXTxState {
	return &BridgeAVAXTxState{uniqueAVAXTx: uniqueAVAXTx}
}

func (avaxTx BridgeAVAXTxState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		UniqueAVAXTx []byte
	}{
		UniqueAVAXTx: avaxTx.uniqueAVAXTx,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (avaxTx *BridgeAVAXTxState) UnmarshalJSON(data []byte) error {
	temp := struct {
		UniqueAVAXTx []byte
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	avaxTx.uniqueAVAXTx = temp.UniqueAVAXTx
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

func (avaxTx BridgeAVAXTxObject) GetVersion() int {
	return avaxTx.version
}

// setError remembers the first non-nil error it is called with.
func (avaxTx *BridgeAVAXTxObject) SetError(err error) {
	if avaxTx.dbErr == nil {
		avaxTx.dbErr = err
	}
}

func (avaxTx BridgeAVAXTxObject) GetTrie(db DatabaseAccessWarper) Trie {
	return avaxTx.trie
}

func (avaxTx *BridgeAVAXTxObject) SetValue(data interface{}) error {
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
	avaxTx.bridgeAVAXTxState = newBridgeAVAXTxState
	return nil
}

func (avaxTx BridgeAVAXTxObject) GetValue() interface{} {
	return avaxTx.bridgeAVAXTxState
}

func (avaxTx BridgeAVAXTxObject) GetValueBytes() []byte {
	data := avaxTx.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal bridge BSC tx state")
	}
	return value
}

func (avaxTx BridgeAVAXTxObject) GetHash() common.Hash {
	return avaxTx.bridgeAVAXTxHash
}

func (avaxTx BridgeAVAXTxObject) GetType() int {
	return avaxTx.objectType
}

// MarkDelete will delete an object in trie
func (avaxTx *BridgeAVAXTxObject) MarkDelete() {
	avaxTx.deleted = true
}

func (avaxTx *BridgeAVAXTxObject) Reset() bool {
	avaxTx.bridgeAVAXTxState = NewBridgeAVAXTxState()
	return true
}

func (avaxTx BridgeAVAXTxObject) IsDeleted() bool {
	return avaxTx.deleted
}

// value is either default or nil
func (avaxTx BridgeAVAXTxObject) IsEmpty() bool {
	temp := NewBridgeAVAXTxState()
	return reflect.DeepEqual(temp, avaxTx.bridgeAVAXTxState) || avaxTx.bridgeAVAXTxState == nil
}
