package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type PDexV3StatusState struct {
	statusType    []byte
	statusSuffix  []byte
	statusContent []byte
}

func (s PDexV3StatusState) StatusType() []byte {
	return s.statusType
}

func (s *PDexV3StatusState) SetStatusType(statusType []byte) {
	s.statusType = statusType
}

func (s PDexV3StatusState) StatusSuffix() []byte {
	return s.statusSuffix
}

func (s *PDexV3StatusState) SetStatusSuffix(statusSuffix []byte) {
	s.statusSuffix = statusSuffix
}

func (s PDexV3StatusState) StatusContent() []byte {
	return s.statusContent
}

func (s *PDexV3StatusState) SetStatusContent(statusContent []byte) {
	s.statusContent = statusContent
}

func (s PDexV3StatusState) MarshalJSON() ([]byte, error) {
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

func (s *PDexV3StatusState) UnmarshalJSON(data []byte) error {
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

func (s *PDexV3StatusState) ToString() string {
	return "{" +
		" \"StatusType\": " + string(s.statusType) +
		" \"StatusSuffix\": " + string(s.statusSuffix) +
		" \"StatusContent\": " + string(s.statusContent) +
		"}"
}

func NewPDexV3StatusState() *PDexV3StatusState {
	return &PDexV3StatusState{}
}

func NewPDexV3StatusStateWithValue(statusType []byte, statusSuffix []byte, statusContent []byte) *PDexV3StatusState {
	return &PDexV3StatusState{statusType: statusType, statusSuffix: statusSuffix, statusContent: statusContent}
}

type PDexV3StatusObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version          int
	PDexV3StatusHash common.Hash
	PDexV3Status     *PDexV3StatusState
	objectType       int
	deleted          bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPDexV3StatusObject(db *StateDB, hash common.Hash) *PDexV3StatusObject {
	return &PDexV3StatusObject{
		version:          defaultVersion,
		db:               db,
		PDexV3StatusHash: hash,
		PDexV3Status:     NewPDexV3StatusState(),
		objectType:       PDexV3StatusObjectType,
		deleted:          false,
	}
}

func newPDexV3StatusObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*PDexV3StatusObject, error) {
	var newPortalStatus = NewPDexV3StatusState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPortalStatus)
		if err != nil {
			return nil, err
		}
	} else {
		newPortalStatus, ok = data.(*PDexV3StatusState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPDexV3StatusStateType, reflect.TypeOf(data))
		}
	}
	return &PDexV3StatusObject{
		version:          defaultVersion,
		PDexV3StatusHash: key,
		PDexV3Status:     newPortalStatus,
		db:               db,
		objectType:       PDexV3StatusObjectType,
		deleted:          false,
	}, nil
}

func GeneratePDexV3StatusObjectKey(statusType []byte, statusSuffix []byte) common.Hash {
	prefixHash := GetPDexV3StatusPrefix(statusType)
	valueHash := common.HashH(append(statusType, statusSuffix...))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (t PDexV3StatusObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *PDexV3StatusObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t PDexV3StatusObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *PDexV3StatusObject) SetValue(data interface{}) error {
	newPortalStatus, ok := data.(*PDexV3StatusState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPDexV3StatusStateType, reflect.TypeOf(data))
	}
	t.PDexV3Status = newPortalStatus
	return nil
}

func (t PDexV3StatusObject) GetValue() interface{} {
	return t.PDexV3Status
}

func (t PDexV3StatusObject) GetValueBytes() []byte {
	PDexV3StatusState, ok := t.GetValue().(*PDexV3StatusState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(PDexV3StatusState)
	if err != nil {
		panic("failed to marshal PDexV3StatusState")
	}
	return value
}

func (t PDexV3StatusObject) GetHash() common.Hash {
	return t.PDexV3StatusHash
}

func (t PDexV3StatusObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *PDexV3StatusObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *PDexV3StatusObject) Reset() bool {
	t.PDexV3Status = NewPDexV3StatusState()
	return true
}

func (t PDexV3StatusObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t PDexV3StatusObject) IsEmpty() bool {
	temp := NewPDexV3StatusState()
	return reflect.DeepEqual(temp, t.PDexV3Status) || t.PDexV3Status == nil
}
