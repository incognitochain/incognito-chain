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

func (bscTx BridgeBSCTxState) UniqueBSCTx() []byte {
	return bscTx.uniqueBSCTx
}

func (bscTx *BridgeBSCTxState) SetUniqueBSCTx(uniqueBSCTx []byte) {
	bscTx.uniqueBSCTx = uniqueBSCTx
}

func NewBridgeBSCTxState() *BridgeBSCTxState {
	return &BridgeBSCTxState{}
}

func NewBridgeBSCTxStateWithValue(uniqueBSCTx []byte) *BridgeBSCTxState {
	return &BridgeBSCTxState{uniqueBSCTx: uniqueBSCTx}
}

func (bscTx BridgeBSCTxState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		UniqueBSCTx []byte
	}{
		UniqueBSCTx: bscTx.uniqueBSCTx,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (bscTx *BridgeBSCTxState) UnmarshalJSON(data []byte) error {
	temp := struct {
		UniqueBSCTx []byte
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	bscTx.uniqueBSCTx = temp.UniqueBSCTx
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

func (bscTx BridgeBSCTxObject) GetVersion() int {
	return bscTx.version
}

// setError remembers the first non-nil error it is called with.
func (bscTx *BridgeBSCTxObject) SetError(err error) {
	if bscTx.dbErr == nil {
		bscTx.dbErr = err
	}
}

func (bscTx BridgeBSCTxObject) GetTrie(db DatabaseAccessWarper) Trie {
	return bscTx.trie
}

func (bscTx *BridgeBSCTxObject) SetValue(data interface{}) error {
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
	bscTx.bridgeBSCTxState = newBridgeBSCTxState
	return nil
}

func (bscTx BridgeBSCTxObject) GetValue() interface{} {
	return bscTx.bridgeBSCTxState
}

func (bscTx BridgeBSCTxObject) GetValueBytes() []byte {
	data := bscTx.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal bridge BSC tx state")
	}
	return value
}

func (bscTx BridgeBSCTxObject) GetHash() common.Hash {
	return bscTx.bridgeBSCTxHash
}

func (bscTx BridgeBSCTxObject) GetType() int {
	return bscTx.objectType
}

// MarkDelete will delete an object in trie
func (bscTx *BridgeBSCTxObject) MarkDelete() {
	bscTx.deleted = true
}

func (bscTx *BridgeBSCTxObject) Reset() bool {
	bscTx.bridgeBSCTxState = NewBridgeBSCTxState()
	return true
}

func (bscTx BridgeBSCTxObject) IsDeleted() bool {
	return bscTx.deleted
}

// value is either default or nil
func (bscTx BridgeBSCTxObject) IsEmpty() bool {
	temp := NewBridgeBSCTxState()
	return reflect.DeepEqual(temp, bscTx.bridgeBSCTxState) || bscTx.bridgeBSCTxState == nil
}
