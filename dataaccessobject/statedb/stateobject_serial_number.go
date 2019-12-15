package statedb

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

type SerialNumberObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	serialNumberHash common.Hash
	serialNumber     []byte
	objectType       int
	deleted          bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newSerialNumberObject(db *StateDB, hash common.Hash) *SerialNumberObject {
	return &SerialNumberObject{
		db:               db,
		serialNumberHash: hash,
		serialNumber:     []byte{},
		objectType:       SerialNumberObjectType,
		deleted:          false,
	}
}
func newSerialNumberObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*SerialNumberObject, error) {
	newSerialNumber, ok := data.([]byte)
	if !ok {
		return nil, NewStatedbError(InvalidByteArrayTypeError, fmt.Errorf("%+v", reflect.TypeOf(data)))
	}
	return &SerialNumberObject{
		serialNumberHash: key,
		serialNumber:     newSerialNumber,
		db:               db,
		objectType:       SerialNumberObjectType,
		deleted:          false,
	}, nil
}

// setError remembers the first non-nil error it is called with.
func (s *SerialNumberObject) SetError(err error) {
	if s.dbErr == nil {
		s.dbErr = err
	}
}

func (s *SerialNumberObject) GetTrie(db DatabaseAccessWarper) Trie {
	return s.trie
}

func (s *SerialNumberObject) SetValue(data interface{}) error {
	newSerialNumber, ok := data.([]byte)
	if !ok {
		return NewStatedbError(InvalidByteArrayTypeError, fmt.Errorf("%+v", reflect.TypeOf(data)))
	}
	s.serialNumber = newSerialNumber
	return nil
}

func (s *SerialNumberObject) GetValue() interface{} {
	return s.serialNumber
}

func (s *SerialNumberObject) GetValueBytes() []byte {
	return s.GetValue().([]byte)
}

func (s *SerialNumberObject) GetHash() common.Hash {
	return s.serialNumberHash
}

func (s *SerialNumberObject) GetType() int {
	return s.objectType
}

// MarkDelete will delete an object in trie
func (s *SerialNumberObject) MarkDelete() {
	s.deleted = true
}

// Reset serial number into default
func (s *SerialNumberObject) Reset() bool {
	s.serialNumber = []byte{}
	return true
}

func (s *SerialNumberObject) IsDeleted() bool {
	return s.deleted
}

// empty value or not
func (s *SerialNumberObject) IsEmpty() bool {
	return len(s.serialNumber[:]) == 0
}
