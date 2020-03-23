package statedb

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

type TestObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	key        common.Hash
	value      []byte
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newTestObject(db *StateDB, hash common.Hash) *TestObject {
	return &TestObject{
		version:    defaultVersion,
		db:         db,
		key:        hash,
		value:      []byte{},
		objectType: TestObjectType,
		deleted:    false,
	}
}
func newTestObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*TestObject, error) {
	newSerialNumber, ok := data.([]byte)
	if !ok {
		return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidByteArrayType, reflect.TypeOf(data))
	}
	return &TestObject{
		version:    defaultVersion,
		key:        key,
		value:      newSerialNumber,
		db:         db,
		objectType: TestObjectType,
		deleted:    false,
	}, nil
}

func (s *TestObject) GetVersion() int {
	return s.version
}

// setError remembers the first non-nil error it is called with.
func (s *TestObject) SetError(err error) {
	if s.dbErr == nil {
		s.dbErr = err
	}
}

func (s *TestObject) GetTrie(db DatabaseAccessWarper) Trie {
	return s.trie
}

func (s *TestObject) SetValue(data interface{}) error {
	newSerialNumber, ok := data.([]byte)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidByteArrayType, reflect.TypeOf(data))
	}
	s.value = newSerialNumber
	return nil
}

func (s *TestObject) GetValue() interface{} {
	return s.value
}

func (s *TestObject) GetValueBytes() []byte {
	return s.GetValue().([]byte)
}

func (s *TestObject) GetHash() common.Hash {
	return s.key
}

func (s *TestObject) GetType() int {
	return s.objectType
}

// MarkDelete will delete an object in trie
func (s *TestObject) MarkDelete() {
	s.deleted = true
}

// Reset serial number into default
func (s *TestObject) Reset() bool {
	s.value = []byte{}
	return true
}

func (s *TestObject) IsDeleted() bool {
	return s.deleted
}

// empty value or not
func (s *TestObject) IsEmpty() bool {
	return len(s.value[:]) == 0
}
