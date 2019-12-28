package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BurningConfirmState struct {
	txID   common.Hash
	height uint64
}

func (b BurningConfirmState) Height() uint64 {
	return b.height
}

func (b *BurningConfirmState) SetHeight(height uint64) {
	b.height = height
}

func (b BurningConfirmState) TxID() common.Hash {
	return b.txID
}

func (b *BurningConfirmState) SetTxID(txID common.Hash) {
	b.txID = txID
}

func NewBurningConfirmState() *BurningConfirmState {
	return &BurningConfirmState{}
}

func NewBurningConfirmStateWithValue(txID common.Hash, height uint64) *BurningConfirmState {
	return &BurningConfirmState{txID: txID, height: height}
}

func (b BurningConfirmState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		TxID   common.Hash
		Height uint64
	}{
		TxID:   b.txID,
		Height: b.height,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (b *BurningConfirmState) UnmarshalJSON(data []byte) error {
	temp := struct {
		TxID   common.Hash
		Height uint64
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	b.txID = temp.TxID
	b.height = temp.Height
	return nil
}

type BurningConfirmObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version       int
	burnStateHash common.Hash
	burnState     *BurningConfirmState
	objectType    int
	deleted       bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBurningConfirmObject(db *StateDB, hash common.Hash) *BurningConfirmObject {
	return &BurningConfirmObject{
		version:       defaultVersion,
		db:            db,
		burnStateHash: hash,
		burnState:     NewBurningConfirmState(),
		objectType:    BurningConfirmObjectType,
		deleted:       false,
	}
}
func newBurningConfirmObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BurningConfirmObject, error) {
	var newBurningConfirmState = NewBurningConfirmState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBurningConfirmState)
		if err != nil {
			return nil, err
		}
	} else {
		newBurningConfirmState, ok = data.(*BurningConfirmState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBurningConfirmStateType, reflect.TypeOf(data))
		}
	}
	return &BurningConfirmObject{
		version:       defaultVersion,
		burnStateHash: key,
		burnState:     newBurningConfirmState,
		db:            db,
		objectType:    BurningConfirmObjectType,
		deleted:       false,
	}, nil
}

func GenerateBurningConfirmObjectKey(txID common.Hash) common.Hash {
	prefixHash := GetBurningConfirmPrefix()
	valueHash := common.HashH(txID[:])
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (ethtx BurningConfirmObject) GetVersion() int {
	return ethtx.version
}

// setError remembers the first non-nil error it is called with.
func (ethtx *BurningConfirmObject) SetError(err error) {
	if ethtx.dbErr == nil {
		ethtx.dbErr = err
	}
}

func (ethtx BurningConfirmObject) GetTrie(db DatabaseAccessWarper) Trie {
	return ethtx.trie
}

func (ethtx *BurningConfirmObject) SetValue(data interface{}) error {
	var newBurningConfirmState = NewBurningConfirmState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBurningConfirmState)
		if err != nil {
			return err
		}
	} else {
		newBurningConfirmState, ok = data.(*BurningConfirmState)
		if !ok {
			return fmt.Errorf("%+v, got type %+v", ErrInvalidBurningConfirmStateType, reflect.TypeOf(data))
		}
	}
	ethtx.burnState = newBurningConfirmState
	return nil
}

func (ethtx BurningConfirmObject) GetValue() interface{} {
	return ethtx.burnState
}

func (ethtx BurningConfirmObject) GetValueBytes() []byte {
	data := ethtx.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal burning confirm state")
	}
	return []byte(value)
}

func (ethtx BurningConfirmObject) GetHash() common.Hash {
	return ethtx.burnStateHash
}

func (ethtx BurningConfirmObject) GetType() int {
	return ethtx.objectType
}

// MarkDelete will delete an object in trie
func (ethtx *BurningConfirmObject) MarkDelete() {
	ethtx.deleted = true
}

func (ethtx *BurningConfirmObject) Reset() bool {
	ethtx.burnState = NewBurningConfirmState()
	return true
}

func (ethtx BurningConfirmObject) IsDeleted() bool {
	return ethtx.deleted
}

// value is either default or nil
func (ethtx BurningConfirmObject) IsEmpty() bool {
	temp := NewBurningConfirmState()
	return reflect.DeepEqual(temp, ethtx.burnState) || ethtx.burnState == nil
}
