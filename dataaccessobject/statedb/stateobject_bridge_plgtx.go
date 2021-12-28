package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BridgePLGTxState struct {
	uniquePLGTx []byte
}

func (plgTx BridgePLGTxState) UniquePLGTx() []byte {
	return plgTx.uniquePLGTx
}

func (plgTx *BridgePLGTxState) SetUniquePLGTx(uniquePLGTx []byte) {
	plgTx.uniquePLGTx = uniquePLGTx
}

func NewBridgePLGTxState() *BridgePLGTxState {
	return &BridgePLGTxState{}
}

func NewBridgePLGTxStateWithValue(uniquePLGTx []byte) *BridgePLGTxState {
	return &BridgePLGTxState{uniquePLGTx: uniquePLGTx}
}

func (plgTx BridgePLGTxState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		UniquePLGTx []byte
	}{
		UniquePLGTx: plgTx.uniquePLGTx,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (plgTx *BridgePLGTxState) UnmarshalJSON(data []byte) error {
	temp := struct {
		UniquePLGTx []byte
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	plgTx.uniquePLGTx = temp.UniquePLGTx
	return nil
}

type BridgePLGTxObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version          int
	bridgePLGTxHash  common.Hash
	BridgePLGTxState *BridgePLGTxState
	objectType       int
	deleted          bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBridgePLGTxObject(db *StateDB, hash common.Hash) *BridgePLGTxObject {
	return &BridgePLGTxObject{
		version:          defaultVersion,
		db:               db,
		bridgePLGTxHash:  hash,
		BridgePLGTxState: NewBridgePLGTxState(),
		objectType:       BridgePLGTxObjectType,
		deleted:          false,
	}
}

func newBridgePLGTxObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BridgePLGTxObject, error) {
	var newBridgePLGTxState = NewBridgePLGTxState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgePLGTxState)
		if err != nil {
			return nil, err
		}
	} else {
		newBridgePLGTxState, ok = data.(*BridgePLGTxState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgePLGTxStateType, reflect.TypeOf(data))
		}
	}
	return &BridgePLGTxObject{
		version:          defaultVersion,
		bridgePLGTxHash:  key,
		BridgePLGTxState: newBridgePLGTxState,
		db:               db,
		objectType:       BridgePLGTxObjectType,
		deleted:          false,
	}, nil
}

func GenerateBridgePLGTxObjectKey(uniquePLGTx []byte) common.Hash {
	prefixHash := GetBridgePLGTxPrefix()
	valueHash := common.HashH(uniquePLGTx)
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (plgTx BridgePLGTxObject) GetVersion() int {
	return plgTx.version
}

// setError remembers the first non-nil error it is called with.
func (plgTx *BridgePLGTxObject) SetError(err error) {
	if plgTx.dbErr == nil {
		plgTx.dbErr = err
	}
}

func (plgTx BridgePLGTxObject) GetTrie(db DatabaseAccessWarper) Trie {
	return plgTx.trie
}

func (plgTx *BridgePLGTxObject) SetValue(data interface{}) error {
	var newBridgePLGTxState = NewBridgePLGTxState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgePLGTxState)
		if err != nil {
			return err
		}
	} else {
		newBridgePLGTxState, ok = data.(*BridgePLGTxState)
		if !ok {
			return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgePLGTxStateType, reflect.TypeOf(data))
		}
	}
	plgTx.BridgePLGTxState = newBridgePLGTxState
	return nil
}

func (plgTx BridgePLGTxObject) GetValue() interface{} {
	return plgTx.BridgePLGTxState
}

func (plgTx BridgePLGTxObject) GetValueBytes() []byte {
	data := plgTx.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal bridge BSC tx state")
	}
	return value
}

func (plgTx BridgePLGTxObject) GetHash() common.Hash {
	return plgTx.bridgePLGTxHash
}

func (plgTx BridgePLGTxObject) GetType() int {
	return plgTx.objectType
}

// MarkDelete will delete an object in trie
func (plgTx *BridgePLGTxObject) MarkDelete() {
	plgTx.deleted = true
}

func (plgTx *BridgePLGTxObject) Reset() bool {
	plgTx.BridgePLGTxState = NewBridgePLGTxState()
	return true
}

func (plgTx BridgePLGTxObject) IsDeleted() bool {
	return plgTx.deleted
}

// value is either default or nil
func (plgTx BridgePLGTxObject) IsEmpty() bool {
	temp := NewBridgePLGTxState()
	return reflect.DeepEqual(temp, plgTx.BridgePLGTxState) || plgTx.BridgePLGTxState == nil
}
