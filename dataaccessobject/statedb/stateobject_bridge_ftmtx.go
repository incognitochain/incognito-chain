package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BridgeFTMTxState struct {
	uniqueFTMTx []byte
}

func (ftmTx BridgeFTMTxState) UniqueFTMTx() []byte {
	return ftmTx.uniqueFTMTx
}

func (ftmTx *BridgeFTMTxState) SetUniqueFTMTx(uniqueFTMTx []byte) {
	ftmTx.uniqueFTMTx = uniqueFTMTx
}

func NewBridgeFTMTxState() *BridgeFTMTxState {
	return &BridgeFTMTxState{}
}

func NewBridgeFTMTxStateWithValue(uniqueFTMTx []byte) *BridgeFTMTxState {
	return &BridgeFTMTxState{uniqueFTMTx: uniqueFTMTx}
}

func (ftmTx BridgeFTMTxState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		UniqueFTMTx []byte
	}{
		UniqueFTMTx: ftmTx.uniqueFTMTx,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (ftmTx *BridgeFTMTxState) UnmarshalJSON(data []byte) error {
	temp := struct {
		UniqueFTMTx []byte
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	ftmTx.uniqueFTMTx = temp.UniqueFTMTx
	return nil
}

type BridgeFTMTxObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version          int
	bridgeFTMTxHash  common.Hash
	bridgeFTMTxState *BridgeFTMTxState
	objectType       int
	deleted          bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBridgeFTMTxObject(db *StateDB, hash common.Hash) *BridgeFTMTxObject {
	return &BridgeFTMTxObject{
		version:          defaultVersion,
		db:               db,
		bridgeFTMTxHash:  hash,
		bridgeFTMTxState: NewBridgeFTMTxState(),
		objectType:       BridgeFTMTxObjectType,
		deleted:          false,
	}
}

func newBridgeFTMTxObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BridgeFTMTxObject, error) {
	var newBridgeFTMTxState = NewBridgeFTMTxState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeFTMTxState)
		if err != nil {
			return nil, err
		}
	} else {
		newBridgeFTMTxState, ok = data.(*BridgeFTMTxState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeFTMTxStateType, reflect.TypeOf(data))
		}
	}
	return &BridgeFTMTxObject{
		version:          defaultVersion,
		bridgeFTMTxHash:  key,
		bridgeFTMTxState: newBridgeFTMTxState,
		db:               db,
		objectType:       BridgeFTMTxObjectType,
		deleted:          false,
	}, nil
}

func GenerateBridgeFTMTxObjectKey(uniqueFTMTx []byte) common.Hash {
	prefixHash := GetBridgeFTMTxPrefix()
	valueHash := common.HashH(uniqueFTMTx)
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (ftmTx BridgeFTMTxObject) GetVersion() int {
	return ftmTx.version
}

// setError remembers the first non-nil error it is called with.
func (ftmTx *BridgeFTMTxObject) SetError(err error) {
	if ftmTx.dbErr == nil {
		ftmTx.dbErr = err
	}
}

func (ftmTx BridgeFTMTxObject) GetTrie(db DatabaseAccessWarper) Trie {
	return ftmTx.trie
}

func (ftmTx *BridgeFTMTxObject) SetValue(data interface{}) error {
	var newBridgeFTMTxState = NewBridgeFTMTxState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeFTMTxState)
		if err != nil {
			return err
		}
	} else {
		newBridgeFTMTxState, ok = data.(*BridgeFTMTxState)
		if !ok {
			return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeFTMTxStateType, reflect.TypeOf(data))
		}
	}
	ftmTx.bridgeFTMTxState = newBridgeFTMTxState
	return nil
}

func (ftmTx BridgeFTMTxObject) GetValue() interface{} {
	return ftmTx.bridgeFTMTxState
}

func (ftmTx BridgeFTMTxObject) GetValueBytes() []byte {
	data := ftmTx.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal bridge BSC tx state")
	}
	return value
}

func (ftmTx BridgeFTMTxObject) GetHash() common.Hash {
	return ftmTx.bridgeFTMTxHash
}

func (ftmTx BridgeFTMTxObject) GetType() int {
	return ftmTx.objectType
}

// MarkDelete will delete an object in trie
func (ftmTx *BridgeFTMTxObject) MarkDelete() {
	ftmTx.deleted = true
}

func (ftmTx *BridgeFTMTxObject) Reset() bool {
	ftmTx.bridgeFTMTxState = NewBridgeFTMTxState()
	return true
}

func (ftmTx BridgeFTMTxObject) IsDeleted() bool {
	return ftmTx.deleted
}

// value is either default or nil
func (ftmTx BridgeFTMTxObject) IsEmpty() bool {
	temp := NewBridgeFTMTxState()
	return reflect.DeepEqual(temp, ftmTx.bridgeFTMTxState) || ftmTx.bridgeFTMTxState == nil
}
