package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BridgeHubTxState struct {
	uniqueBridgeHubTx []byte
}

func (bridgeHubTx BridgeHubTxState) UniqueBridgeHubTx() []byte {
	return bridgeHubTx.uniqueBridgeHubTx
}

func (bridgeHubTx *BridgeHubTxState) SetUniqueBridgeHubTx(uniqueBridgeHubTx []byte) {
	bridgeHubTx.uniqueBridgeHubTx = uniqueBridgeHubTx
}

func NewBridgeHubTxState() *BridgeHubTxState {
	return &BridgeHubTxState{}
}

func NewBridgeHubTxStateWithValue(uniqueBridgeHubTx []byte) *BridgeHubTxState {
	return &BridgeHubTxState{uniqueBridgeHubTx: uniqueBridgeHubTx}
}

func (bridgeHubTx BridgeHubTxState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		UniqueBridgeHubTx []byte
	}{
		UniqueBridgeHubTx: bridgeHubTx.uniqueBridgeHubTx,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (bridgeHubTx *BridgeHubTxState) UnmarshalJSON(data []byte) error {
	temp := struct {
		UniqueBridgeHubTx []byte
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	bridgeHubTx.uniqueBridgeHubTx = temp.UniqueBridgeHubTx
	return nil
}

type BridgeHubTxObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version          int
	bridgeFTMTxHash  common.Hash
	BridgeHubTxState *BridgeHubTxState
	objectType       int
	deleted          bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBridgeHubTxObject(db *StateDB, hash common.Hash) *BridgeHubTxObject {
	return &BridgeHubTxObject{
		version:          defaultVersion,
		db:               db,
		bridgeFTMTxHash:  hash,
		BridgeHubTxState: NewBridgeHubTxState(),
		objectType:       BridgeHubTxObjectType,
		deleted:          false,
	}
}

func newBridgeHubTxObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BridgeHubTxObject, error) {
	var newBridgeHubTxState = NewBridgeHubTxState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeHubTxState)
		if err != nil {
			return nil, err
		}
	} else {
		newBridgeHubTxState, ok = data.(*BridgeHubTxState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeHubTxStateType, reflect.TypeOf(data))
		}
	}
	return &BridgeHubTxObject{
		version:          defaultVersion,
		bridgeFTMTxHash:  key,
		BridgeHubTxState: newBridgeHubTxState,
		db:               db,
		objectType:       BridgeHubTxObjectType,
		deleted:          false,
	}, nil
}

func GenerateBridgeHubTxObjectKey(uniqueBridgeHubTx []byte) common.Hash {
	prefixHash := GetBridgeHubTxPrefix()
	valueHash := common.HashH(uniqueBridgeHubTx)
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (bridgeHubTx BridgeHubTxObject) GetVersion() int {
	return bridgeHubTx.version
}

// setError remembers the first non-nil error it is called with.
func (bridgeHubTx *BridgeHubTxObject) SetError(err error) {
	if bridgeHubTx.dbErr == nil {
		bridgeHubTx.dbErr = err
	}
}

func (bridgeHubTx BridgeHubTxObject) GetTrie(db DatabaseAccessWarper) Trie {
	return bridgeHubTx.trie
}

func (bridgeHubTx *BridgeHubTxObject) SetValue(data interface{}) error {
	var newBridgeHubTxState = NewBridgeHubTxState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeHubTxState)
		if err != nil {
			return err
		}
	} else {
		newBridgeHubTxState, ok = data.(*BridgeHubTxState)
		if !ok {
			return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeHubTxStateType, reflect.TypeOf(data))
		}
	}
	bridgeHubTx.BridgeHubTxState = newBridgeHubTxState
	return nil
}

func (bridgeHubTx BridgeHubTxObject) GetValue() interface{} {
	return bridgeHubTx.BridgeHubTxState
}

func (bridgeHubTx BridgeHubTxObject) GetValueBytes() []byte {
	data := bridgeHubTx.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal bridge BSC tx state")
	}
	return value
}

func (bridgeHubTx BridgeHubTxObject) GetHash() common.Hash {
	return bridgeHubTx.bridgeFTMTxHash
}

func (bridgeHubTx BridgeHubTxObject) GetType() int {
	return bridgeHubTx.objectType
}

// MarkDelete will delete an object in trie
func (bridgeHubTx *BridgeHubTxObject) MarkDelete() {
	bridgeHubTx.deleted = true
}

func (bridgeHubTx *BridgeHubTxObject) Reset() bool {
	bridgeHubTx.BridgeHubTxState = NewBridgeHubTxState()
	return true
}

func (bridgeHubTx BridgeHubTxObject) IsDeleted() bool {
	return bridgeHubTx.deleted
}

// value is either default or nil
func (bridgeHubTx BridgeHubTxObject) IsEmpty() bool {
	temp := NewBridgeHubTxState()
	return reflect.DeepEqual(temp, bridgeHubTx.BridgeHubTxState) || bridgeHubTx.BridgeHubTxState == nil
}
