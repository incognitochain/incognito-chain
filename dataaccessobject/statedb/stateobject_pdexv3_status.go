package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type Pdexv3StatusState struct {
	statusType    []byte
	statusSuffix  []byte
	statusContent []byte
}

func (s Pdexv3StatusState) StatusType() []byte {
	return s.statusType
}

func (s *Pdexv3StatusState) SetStatusType(statusType []byte) {
	s.statusType = statusType
}

func (s Pdexv3StatusState) StatusSuffix() []byte {
	return s.statusSuffix
}

func (s *Pdexv3StatusState) SetStatusSuffix(statusSuffix []byte) {
	s.statusSuffix = statusSuffix
}

func (s Pdexv3StatusState) StatusContent() []byte {
	return s.statusContent
}

func (s *Pdexv3StatusState) SetStatusContent(statusContent []byte) {
	s.statusContent = statusContent
}

func (s Pdexv3StatusState) MarshalJSON() ([]byte, error) {
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

func (s *Pdexv3StatusState) UnmarshalJSON(data []byte) error {
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

func (s *Pdexv3StatusState) ToString() string {
	return "{" +
		" \"StatusType\": " + string(s.statusType) +
		" \"StatusSuffix\": " + string(s.statusSuffix) +
		" \"StatusContent\": " + string(s.statusContent) +
		"}"
}

func NewPdexv3StatusState() *Pdexv3StatusState {
	return &Pdexv3StatusState{}
}

func NewPdexv3StatusStateWithValue(statusType []byte, statusSuffix []byte, statusContent []byte) *Pdexv3StatusState {
	return &Pdexv3StatusState{statusType: statusType, statusSuffix: statusSuffix, statusContent: statusContent}
}

type Pdexv3StatusObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version          int
	Pdexv3StatusHash common.Hash
	Pdexv3Status     *Pdexv3StatusState
	objectType       int
	deleted          bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPdexv3StatusObject(db *StateDB, hash common.Hash) *Pdexv3StatusObject {
	return &Pdexv3StatusObject{
		version:          defaultVersion,
		db:               db,
		Pdexv3StatusHash: hash,
		Pdexv3Status:     NewPdexv3StatusState(),
		objectType:       Pdexv3StatusObjectType,
		deleted:          false,
	}
}

func newPdexv3StatusObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*Pdexv3StatusObject, error) {
	var newPortalStatus = NewPdexv3StatusState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPortalStatus)
		if err != nil {
			return nil, err
		}
	} else {
		newPortalStatus, ok = data.(*Pdexv3StatusState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3StatusStateType, reflect.TypeOf(data))
		}
	}
	return &Pdexv3StatusObject{
		version:          defaultVersion,
		Pdexv3StatusHash: key,
		Pdexv3Status:     newPortalStatus,
		db:               db,
		objectType:       Pdexv3StatusObjectType,
		deleted:          false,
	}, nil
}

func GeneratePdexv3StatusObjectKey(statusType []byte, statusSuffix []byte) common.Hash {
	prefixHash := GetPdexv3StatusPrefix(statusType)
	valueHash := common.HashH(append(statusType, statusSuffix...))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (t Pdexv3StatusObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *Pdexv3StatusObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t Pdexv3StatusObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *Pdexv3StatusObject) SetValue(data interface{}) error {
	newPortalStatus, ok := data.(*Pdexv3StatusState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3StatusStateType, reflect.TypeOf(data))
	}
	t.Pdexv3Status = newPortalStatus
	return nil
}

func (t Pdexv3StatusObject) GetValue() interface{} {
	return t.Pdexv3Status
}

func (t Pdexv3StatusObject) GetValueBytes() []byte {
	Pdexv3StatusState, ok := t.GetValue().(*Pdexv3StatusState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(Pdexv3StatusState)
	if err != nil {
		panic("failed to marshal Pdexv3StatusState")
	}
	return value
}

func (t Pdexv3StatusObject) GetHash() common.Hash {
	return t.Pdexv3StatusHash
}

func (t Pdexv3StatusObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *Pdexv3StatusObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *Pdexv3StatusObject) Reset() bool {
	t.Pdexv3Status = NewPdexv3StatusState()
	return true
}

func (t Pdexv3StatusObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t Pdexv3StatusObject) IsEmpty() bool {
	temp := NewPdexv3StatusState()
	return reflect.DeepEqual(temp, t.Pdexv3Status) || t.Pdexv3Status == nil
}
