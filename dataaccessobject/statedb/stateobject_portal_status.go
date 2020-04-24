package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

type PortalStatusState struct {
	statusType    []byte
	statusSuffix  []byte
	statusContent []byte
}

func (s PortalStatusState) StatusType() []byte {
	return s.statusType
}

func (s *PortalStatusState) SetStatusType(statusType []byte) {
	s.statusType = statusType
}

func (s PortalStatusState) StatusSuffix() []byte {
	return s.statusSuffix
}

func (s *PortalStatusState) SetStatusSuffix(statusSuffix []byte) {
	s.statusSuffix = statusSuffix
}

func (s PortalStatusState) StatusContent() []byte {
	return s.statusContent
}

func (s *PortalStatusState) SetStatusContent(statusContent []byte) {
	s.statusContent = statusContent
}

func (s PortalStatusState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		StatusType    []byte
		StatusSuffix  []byte
		StatusContent []byte
	}{
		StatusType:    s.statusType,
		StatusSuffix:  s.statusSuffix,
		StatusContent: s.statusContent,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (s *PortalStatusState) UnmarshalJSON(data []byte) error {
	temp := struct {
		StatusType    []byte
		StatusSuffix  []byte
		StatusContent []byte
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	s.statusType = temp.StatusType
	s.statusContent = temp.StatusContent
	return nil
}

func (s *PortalStatusState) ToString() string {
	return "{" +
		" \"StatusType\": " + string(s.statusType) +
		" \"StatusSuffix\": " + string(s.statusSuffix) +
		" \"StatusContent\": " + string(s.statusContent) +
		"}"
}

func NewPortalStatusState() *PortalStatusState {
	return &PortalStatusState{}
}

func NewPortalStatusStateWithValue(statusType []byte, statusSuffix []byte, statusContent []byte) *PortalStatusState {
	return &PortalStatusState{statusType: statusType, statusSuffix: statusSuffix, statusContent: statusContent}
}

type PortalStatusObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version          int
	portalStatusHash common.Hash
	portalStatus     *PortalStatusState
	objectType       int
	deleted          bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPortalStatusObject(db *StateDB, hash common.Hash) *PortalStatusObject {
	return &PortalStatusObject{
		version:          defaultVersion,
		db:               db,
		portalStatusHash: hash,
		portalStatus:     NewPortalStatusState(),
		objectType:       PortalStatusObjectType,
		deleted:          false,
	}
}

func newPortalStatusObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*PortalStatusObject, error) {
	var newPortalStatus = NewPortalStatusState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPortalStatus)
		if err != nil {
			return nil, err
		}
	} else {
		newPortalStatus, ok = data.(*PortalStatusState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPortalStatusStateType, reflect.TypeOf(data))
		}
	}
	return &PortalStatusObject{
		version:          defaultVersion,
		portalStatusHash: key,
		portalStatus:     newPortalStatus,
		db:               db,
		objectType:       PortalStatusObjectType,
		deleted:          false,
	}, nil
}

func GeneratePortalStatusObjectKey(statusType []byte, statusSuffix []byte) common.Hash {
	prefixHash := GetPortalStatusPrefix()
	valueHash := common.HashH(append(statusType, statusSuffix...))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (t PortalStatusObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *PortalStatusObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t PortalStatusObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *PortalStatusObject) SetValue(data interface{}) error {
	newPortalStatus, ok := data.(*PortalStatusState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPortalStatusStateType, reflect.TypeOf(data))
	}
	t.portalStatus = newPortalStatus
	return nil
}

func (t PortalStatusObject) GetValue() interface{} {
	return t.portalStatus
}

func (t PortalStatusObject) GetValueBytes() []byte {
	portalStatusState, ok := t.GetValue().(*PortalStatusState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(portalStatusState)
	if err != nil {
		panic("failed to marshal portalStatusState")
	}
	return value
}

func (t PortalStatusObject) GetHash() common.Hash {
	return t.portalStatusHash
}

func (t PortalStatusObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *PortalStatusObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *PortalStatusObject) Reset() bool {
	t.portalStatus = NewPortalStatusState()
	return true
}

func (t PortalStatusObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t PortalStatusObject) IsEmpty() bool {
	temp := NewPortalStatusState()
	return reflect.DeepEqual(temp, t.portalStatus) || t.portalStatus == nil
}
