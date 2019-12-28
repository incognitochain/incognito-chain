package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BridgeStatusState struct {
	txReqID common.Hash
	status  byte
}

func (s BridgeStatusState) TxReqID() common.Hash {
	return s.txReqID
}

func (s *BridgeStatusState) SetTxReqID(txReqID common.Hash) {
	s.txReqID = txReqID
}

func (s BridgeStatusState) Status() byte {
	return s.status
}

func (s *BridgeStatusState) SetStatus(status byte) {
	s.status = status
}

func (s BridgeStatusState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		TxReqID common.Hash
		Status  byte
	}{
		TxReqID: s.txReqID,
		Status:  s.status,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (s *BridgeStatusState) UnmarshalJSON(data []byte) error {
	temp := struct {
		TxReqID common.Hash
		Status  byte
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	s.txReqID = temp.TxReqID
	s.status = temp.Status
	return nil
}

func NewBridgeStatusState() *BridgeStatusState {
	return &BridgeStatusState{}
}

func NewBridgeStatusStateWithValue(txReqID common.Hash, status byte) *BridgeStatusState {
	return &BridgeStatusState{txReqID: txReqID, status: status}
}

type BridgeStatusObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version           int
	bridgeStatusHash  common.Hash
	bridgeStatusState *BridgeStatusState
	objectType        int
	deleted           bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBridgeStatusObject(db *StateDB, hash common.Hash) *BridgeStatusObject {
	return &BridgeStatusObject{
		version:           defaultVersion,
		db:                db,
		bridgeStatusHash:  hash,
		bridgeStatusState: NewBridgeStatusState(),
		objectType:        BridgeStatusObjectType,
		deleted:           false,
	}
}
func newBridgeStatusObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BridgeStatusObject, error) {
	var newBridgeStatusState = NewBridgeStatusState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeStatusState)
		if err != nil {
			return nil, err
		}
	} else {
		newBridgeStatusState, ok = data.(*BridgeStatusState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeStatusStateType, reflect.TypeOf(data))
		}
	}
	return &BridgeStatusObject{
		version:           defaultVersion,
		bridgeStatusHash:  key,
		bridgeStatusState: newBridgeStatusState,
		db:                db,
		objectType:        BridgeStatusObjectType,
		deleted:           false,
	}, nil
}

func GenerateBridgeStatusObjectKey(txReqID common.Hash) common.Hash {
	prefixHash := GetBridgeStatusPrefix()
	valueHash := common.HashH(txReqID[:])
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (s BridgeStatusObject) GetVersion() int {
	return s.version
}

// setError remembers the first non-nil error it is called with.
func (s *BridgeStatusObject) SetError(err error) {
	if s.dbErr == nil {
		s.dbErr = err
	}
}

func (s BridgeStatusObject) GetTrie(db DatabaseAccessWarper) Trie {
	return s.trie
}

func (s *BridgeStatusObject) SetValue(data interface{}) error {
	var newBridgeStatusState = NewBridgeStatusState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeStatusState)
		if err != nil {
			return err
		}
	} else {
		newBridgeStatusState, ok = data.(*BridgeStatusState)
		if !ok {
			return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeStatusStateType, reflect.TypeOf(data))
		}
	}
	s.bridgeStatusState = newBridgeStatusState
	return nil
}

func (s BridgeStatusObject) GetValue() interface{} {
	return s.bridgeStatusState
}

func (s BridgeStatusObject) GetValueBytes() []byte {
	data := s.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal bridge status state")
	}
	return []byte(value)
}

func (s BridgeStatusObject) GetHash() common.Hash {
	return s.bridgeStatusHash
}

func (s BridgeStatusObject) GetType() int {
	return s.objectType
}

// MarkDelete will delete an object in trie
func (s *BridgeStatusObject) MarkDelete() {
	s.deleted = true
}

func (s *BridgeStatusObject) Reset() bool {
	s.bridgeStatusState = NewBridgeStatusState()
	return true
}

func (s BridgeStatusObject) IsDeleted() bool {
	return s.deleted
}

// value is either default or nil
func (s BridgeStatusObject) IsEmpty() bool {
	temp := NewBridgeStatusState()
	return reflect.DeepEqual(temp, s.bridgeStatusState) || s.bridgeStatusState == nil
}
