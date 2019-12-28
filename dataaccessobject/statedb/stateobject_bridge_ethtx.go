package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BridgeEthTxState struct {
	uniqueEthTx []byte
}

func (ethtx BridgeEthTxState) UniqueEthTx() []byte {
	return ethtx.uniqueEthTx
}

func (ethtx *BridgeEthTxState) SetUniqueEthTx(uniqueEthTx []byte) {
	ethtx.uniqueEthTx = uniqueEthTx
}

func NewBridgeEthTxState() *BridgeEthTxState {
	return &BridgeEthTxState{}
}

func NewBridgeEthTxStateWithValue(uniqueETHTx []byte) *BridgeEthTxState {
	return &BridgeEthTxState{uniqueEthTx: uniqueETHTx}
}

func (ethtx BridgeEthTxState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		UniqueEthTx []byte
	}{
		UniqueEthTx: ethtx.uniqueEthTx,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (ethtx *BridgeEthTxState) UnmarshalJSON(data []byte) error {
	temp := struct {
		UniqueEthTx []byte
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	ethtx.uniqueEthTx = temp.UniqueEthTx
	return nil
}

type BridgeEthTxObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version          int
	bridgeEthTxHash  common.Hash
	bridgeEthTxState *BridgeEthTxState
	objectType       int
	deleted          bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBridgeEthTxObject(db *StateDB, hash common.Hash) *BridgeEthTxObject {
	return &BridgeEthTxObject{
		version:          defaultVersion,
		db:               db,
		bridgeEthTxHash:  hash,
		bridgeEthTxState: NewBridgeEthTxState(),
		objectType:       BridgeEthTxObjectType,
		deleted:          false,
	}
}
func newBridgeEthTxObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BridgeEthTxObject, error) {
	var newBridgeEthTxState = NewBridgeEthTxState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeEthTxState)
		if err != nil {
			return nil, err
		}
	} else {
		newBridgeEthTxState, ok = data.(*BridgeEthTxState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeEthTxStateType, reflect.TypeOf(data))
		}
	}
	return &BridgeEthTxObject{
		version:          defaultVersion,
		bridgeEthTxHash:  key,
		bridgeEthTxState: newBridgeEthTxState,
		db:               db,
		objectType:       BridgeEthTxObjectType,
		deleted:          false,
	}, nil
}

func GenerateBridgeEthTxObjectKey(uniqueEthTx []byte) common.Hash {
	prefixHash := GetBridgeEthTxPrefix()
	valueHash := common.HashH(uniqueEthTx)
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (ethtx BridgeEthTxObject) GetVersion() int {
	return ethtx.version
}

// setError remembers the first non-nil error it is called with.
func (ethtx *BridgeEthTxObject) SetError(err error) {
	if ethtx.dbErr == nil {
		ethtx.dbErr = err
	}
}

func (ethtx BridgeEthTxObject) GetTrie(db DatabaseAccessWarper) Trie {
	return ethtx.trie
}

func (ethtx *BridgeEthTxObject) SetValue(data interface{}) error {
	var newBridgeEthTxState = NewBridgeEthTxState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeEthTxState)
		if err != nil {
			return err
		}
	} else {
		newBridgeEthTxState, ok = data.(*BridgeEthTxState)
		if !ok {
			return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeEthTxStateType, reflect.TypeOf(data))
		}
	}
	ethtx.bridgeEthTxState = newBridgeEthTxState
	return nil
}

func (ethtx BridgeEthTxObject) GetValue() interface{} {
	return ethtx.bridgeEthTxState
}

func (ethtx BridgeEthTxObject) GetValueBytes() []byte {
	data := ethtx.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal bridge eth tx state")
	}
	return []byte(value)
}

func (ethtx BridgeEthTxObject) GetHash() common.Hash {
	return ethtx.bridgeEthTxHash
}

func (ethtx BridgeEthTxObject) GetType() int {
	return ethtx.objectType
}

// MarkDelete will delete an object in trie
func (ethtx *BridgeEthTxObject) MarkDelete() {
	ethtx.deleted = true
}

func (ethtx *BridgeEthTxObject) Reset() bool {
	ethtx.bridgeEthTxState = NewBridgeEthTxState()
	return true
}

func (ethtx BridgeEthTxObject) IsDeleted() bool {
	return ethtx.deleted
}

// value is either default or nil
func (ethtx BridgeEthTxObject) IsEmpty() bool {
	temp := NewBridgeEthTxState()
	return reflect.DeepEqual(temp, ethtx.bridgeEthTxState) || ethtx.bridgeEthTxState == nil
}
