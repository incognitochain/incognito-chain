package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type PortalExternalTxState struct {
	uniqueExternalTx []byte
}

func (ethtx PortalExternalTxState) UniqueExternalTx() []byte {
	return ethtx.uniqueExternalTx
}

func (ethtx *PortalExternalTxState) SetUniqueExternalTx(uniqueExtTx []byte) {
	ethtx.uniqueExternalTx = uniqueExtTx
}

func NewPortalExternalTxState() *PortalExternalTxState {
	return &PortalExternalTxState{}
}

func NewPortalExternalTxStateWithValue(uniqueExtTx []byte) *PortalExternalTxState {
	return &PortalExternalTxState{uniqueExternalTx: uniqueExtTx}
}

func (ethtx PortalExternalTxState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		UniqueExternalTx []byte
	}{
		UniqueExternalTx: ethtx.uniqueExternalTx,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (ethtx *PortalExternalTxState) UnmarshalJSON(data []byte) error {
	temp := struct {
		UniqueExternalTx []byte
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	ethtx.uniqueExternalTx = temp.UniqueExternalTx
	return nil
}

type PortalExternalTxStateObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version               int
	portalExternalTxHash  common.Hash
	portalExternalTxState *PortalExternalTxState
	objectType            int
	deleted               bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPortalExternalTxObject(db *StateDB, hash common.Hash) *PortalExternalTxStateObject {
	return &PortalExternalTxStateObject{
		version:               defaultVersion,
		db:                    db,
		portalExternalTxHash:  hash,
		portalExternalTxState: NewPortalExternalTxState(),
		objectType:            PortalExternalTxObjectType,
		deleted:               false,
	}
}

func newPortalExternalTxObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*PortalExternalTxStateObject, error) {
	var newPortalEthTxState = NewPortalExternalTxState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPortalEthTxState)
		if err != nil {
			return nil, err
		}
	} else {
		newPortalEthTxState, ok = data.(*PortalExternalTxState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPortalExternalTxStateType, reflect.TypeOf(data))
		}
	}
	return &PortalExternalTxStateObject{
		version:               defaultVersion,
		portalExternalTxHash:  key,
		portalExternalTxState: newPortalEthTxState,
		db:                    db,
		objectType:            PortalExternalTxObjectType,
		deleted:               false,
	}, nil
}

func GeneratePortalExternalTxObjectKey(uniqueExtTx []byte) common.Hash {
	prefixHash := GetPortalExternalTxPrefix()
	valueHash := common.HashH(uniqueExtTx)
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (extTx PortalExternalTxStateObject) GetVersion() int {
	return extTx.version
}

// setError remembers the first non-nil error it is called with.
func (extTx *PortalExternalTxStateObject) SetError(err error) {
	if extTx.dbErr == nil {
		extTx.dbErr = err
	}
}

func (extTx PortalExternalTxStateObject) GetTrie(db DatabaseAccessWarper) Trie {
	return extTx.trie
}

func (extTx *PortalExternalTxStateObject) SetValue(data interface{}) error {
	var newPortalEthTxState = NewPortalExternalTxState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPortalEthTxState)
		if err != nil {
			return err
		}
	} else {
		newPortalEthTxState, ok = data.(*PortalExternalTxState)
		if !ok {
			return fmt.Errorf("%+v, got type %+v", ErrInvalidPortalExternalTxStateType, reflect.TypeOf(data))
		}
	}
	extTx.portalExternalTxState = newPortalEthTxState
	return nil
}

func (extTx PortalExternalTxStateObject) GetValue() interface{} {
	return extTx.portalExternalTxState
}

func (extTx PortalExternalTxStateObject) GetValueBytes() []byte {
	data := extTx.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal portal eth tx state")
	}
	return []byte(value)
}

func (extTx PortalExternalTxStateObject) GetHash() common.Hash {
	return extTx.portalExternalTxHash
}

func (extTx PortalExternalTxStateObject) GetType() int {
	return extTx.objectType
}

// MarkDelete will delete an object in trie
func (extTx *PortalExternalTxStateObject) MarkDelete() {
	extTx.deleted = true
}

func (extTx *PortalExternalTxStateObject) Reset() bool {
	extTx.portalExternalTxState = NewPortalExternalTxState()
	return true
}

func (extTx PortalExternalTxStateObject) IsDeleted() bool {
	return extTx.deleted
}

// value is either default or nil
func (extTx PortalExternalTxStateObject) IsEmpty() bool {
	temp := NewPortalExternalTxState()
	return reflect.DeepEqual(temp, extTx.portalExternalTxState) || extTx.portalExternalTxState == nil
}
