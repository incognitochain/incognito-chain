package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type PortalV4StatusState struct {
	statusType    []byte
	statusSuffix  []byte
	statusContent []byte
}

func (s PortalV4StatusState) StatusType() []byte {
	return s.statusType
}

func (s *PortalV4StatusState) SetStatusType(statusType []byte) {
	s.statusType = statusType
}

func (s PortalV4StatusState) StatusSuffix() []byte {
	return s.statusSuffix
}

func (s *PortalV4StatusState) SetStatusSuffix(statusSuffix []byte) {
	s.statusSuffix = statusSuffix
}

func (s PortalV4StatusState) StatusContent() []byte {
	return s.statusContent
}

func (s *PortalV4StatusState) SetStatusContent(statusContent []byte) {
	s.statusContent = statusContent
}

func (s PortalV4StatusState) MarshalJSON() ([]byte, error) {
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

func (s *PortalV4StatusState) UnmarshalJSON(data []byte) error {
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

func (s *PortalV4StatusState) ToString() string {
	return "{" +
		" \"StatusType\": " + string(s.statusType) +
		" \"StatusSuffix\": " + string(s.statusSuffix) +
		" \"StatusContent\": " + string(s.statusContent) +
		"}"
}

func NewPortalV4StatusState() *PortalV4StatusState {
	return &PortalV4StatusState{}
}

func NewPortalV4StatusStateWithValue(statusType []byte, statusSuffix []byte, statusContent []byte) *PortalV4StatusState {
	return &PortalV4StatusState{statusType: statusType, statusSuffix: statusSuffix, statusContent: statusContent}
}

type PortalV4StatusObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version            int
	portalV4StatusHash common.Hash
	portalV4Status     *PortalV4StatusState
	objectType         int
	deleted            bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPortalV4StatusObject(db *StateDB, hash common.Hash) *PortalV4StatusObject {
	return &PortalV4StatusObject{
		version:            defaultVersion,
		db:                 db,
		portalV4StatusHash: hash,
		portalV4Status:     NewPortalV4StatusState(),
		objectType:         PortalV4StatusObjectType,
		deleted:            false,
	}
}

func newPortalV4StatusObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*PortalV4StatusObject, error) {
	var newPortalStatus = NewPortalV4StatusState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPortalStatus)
		if err != nil {
			return nil, err
		}
	} else {
		newPortalStatus, ok = data.(*PortalV4StatusState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPortalV4StatusStateType, reflect.TypeOf(data))
		}
	}
	return &PortalV4StatusObject{
		version:            defaultVersion,
		portalV4StatusHash: key,
		portalV4Status:     newPortalStatus,
		db:                 db,
		objectType:         PortalV4StatusObjectType,
		deleted:            false,
	}, nil
}

func GeneratePortalV4StatusObjectKey(statusType []byte, statusSuffix []byte) common.Hash {
	prefixHash := GetPortalV4StatusPrefix()
	valueHash := common.HashH(append(statusType, statusSuffix...))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (t PortalV4StatusObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *PortalV4StatusObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t PortalV4StatusObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *PortalV4StatusObject) SetValue(data interface{}) error {
	newPortalStatus, ok := data.(*PortalV4StatusState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPortalV4StatusStateType, reflect.TypeOf(data))
	}
	t.portalV4Status = newPortalStatus
	return nil
}

func (t PortalV4StatusObject) GetValue() interface{} {
	return t.portalV4Status
}

func (t PortalV4StatusObject) GetValueBytes() []byte {
	PortalV4StatusState, ok := t.GetValue().(*PortalV4StatusState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(PortalV4StatusState)
	if err != nil {
		panic("failed to marshal PortalV4StatusState")
	}
	return value
}

func (t PortalV4StatusObject) GetHash() common.Hash {
	return t.portalV4StatusHash
}

func (t PortalV4StatusObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *PortalV4StatusObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *PortalV4StatusObject) Reset() bool {
	t.portalV4Status = NewPortalV4StatusState()
	return true
}

func (t PortalV4StatusObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t PortalV4StatusObject) IsEmpty() bool {
	temp := NewPortalV4StatusState()
	return reflect.DeepEqual(temp, t.portalV4Status) || t.portalV4Status == nil
}
