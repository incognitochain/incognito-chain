package statedb

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

type PDEStatusState struct {
	statusType int
	status     int
	refundData []byte
}

func (s PDEStatusState) StatusType() int {
	return s.statusType
}

func (s *PDEStatusState) SetStatusType(statusType int) {
	s.statusType = statusType
}

func (s PDEStatusState) Status() int {
	return s.status
}

func (s *PDEStatusState) SetStatus(status int) {
	s.status = status
}

func (s PDEStatusState) RefundData() []byte {
	return s.refundData
}

func (s *PDEStatusState) SetRefundData(refundData []byte) {
	s.refundData = refundData
}

func (s PDEStatusState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		StatusType int
		Status     int
		RefundData []byte
	}{
		StatusType: s.statusType,
		Status:     s.status,
		RefundData: s.refundData,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (s *PDEStatusState) UnmarshalJSON(data []byte) error {
	temp := struct {
		StatusType int
		Status     int
		RefundData []byte
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	s.statusType = temp.StatusType
	s.status = temp.Status
	s.refundData = temp.RefundData
	return nil
}

func NewPDEStatusState() *PDEStatusState {
	return &PDEStatusState{}
}

func NewPDEStatusStateWithValue(statusType int, status int, refundData []byte) *PDEStatusState {
	return &PDEStatusState{statusType: statusType, status: status, refundData: refundData}
}

type PDEStatusObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version                     int
	waitingPDEContributionHash  common.Hash
	waitingPDEContributionState *PDEStatusState
	objectType                  int
	deleted                     bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPDEStatusObject(db *StateDB, hash common.Hash) *PDEStatusObject {
	return &PDEStatusObject{
		version:                     defaultVersion,
		db:                          db,
		waitingPDEContributionHash:  hash,
		waitingPDEContributionState: NewPDEStatusState(),
		objectType:                  PDEStatusObjectType,
		deleted:                     false,
	}
}
func newPDEStatusObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*PDEStatusObject, error) {
	var newPDEStatus = NewPDEStatusState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPDEStatus)
		if err != nil {
			return nil, err
		}
	} else {
		newPDEStatus, ok = data.(*PDEStatusState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPDEStatusStateType, reflect.TypeOf(data))
		}
	}
	return &PDEStatusObject{
		version:                     defaultVersion,
		waitingPDEContributionHash:  key,
		waitingPDEContributionState: newPDEStatus,
		db:                          db,
		objectType:                  PDEStatusObjectType,
		deleted:                     false,
	}, nil
}

func GeneratePDEStatusObjectKey(statusType int, key []byte) common.Hash {
	prefixHash := GetPDEStatusPrefix()
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(statusType))
	valueHash := common.HashH(append(buf, key...))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (t PDEStatusObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *PDEStatusObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t PDEStatusObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *PDEStatusObject) SetValue(data interface{}) error {
	newPDEStatus, ok := data.(*PDEStatusState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPDEStatusStateType, reflect.TypeOf(data))
	}
	t.waitingPDEContributionState = newPDEStatus
	return nil
}

func (t PDEStatusObject) GetValue() interface{} {
	return t.waitingPDEContributionState
}

func (t PDEStatusObject) GetValueBytes() []byte {
	waitingPDEcontributionState, ok := t.GetValue().(*PDEStatusState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(waitingPDEcontributionState)
	if err != nil {
		panic("failed to marshal token state")
	}
	return value
}

func (t PDEStatusObject) GetHash() common.Hash {
	return t.waitingPDEContributionHash
}

func (t PDEStatusObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *PDEStatusObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *PDEStatusObject) Reset() bool {
	t.waitingPDEContributionState = NewPDEStatusState()
	return true
}

func (t PDEStatusObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t PDEStatusObject) IsEmpty() bool {
	temp := NewPDEStatusState()
	return reflect.DeepEqual(temp, t.waitingPDEContributionState) || t.waitingPDEContributionState == nil
}
