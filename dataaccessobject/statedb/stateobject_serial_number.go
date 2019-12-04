package statedb

import (
	"github.com/incognitochain/incognito-chain/common"
)

type SerialNumber []byte
type SerialNumberObject struct {
	serialNumberHash common.Hash
	serialNumber     SerialNumber
	db               *StateDB
	objectType       int
	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access
}

func newSerialNumberObject(db *StateDB, hash common.Hash) *SerialNumberObject {
	return &SerialNumberObject{
		db:               db,
		serialNumberHash: hash,
		serialNumber:     []byte{},
		objectType:       SerialNumberObjectType,
	}
}
func newSerialNumberObjectWithValue(db *StateDB, key common.Hash, data interface{}) *SerialNumberObject {
	newSerialNumber, ok := data.([]byte)
	if !ok {
		panic("Wrong expected value")
	}
	return &SerialNumberObject{
		serialNumberHash: key,
		serialNumber:     newSerialNumber,
		db:               db,
		objectType:       SerialNumberObjectType,
	}
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

func (s *SerialNumberObject) SetValue(data interface{}) {
	newSerialNumber, ok := data.([]byte)
	if !ok {
		panic("Wrong expected value")
	}
	s.serialNumber = newSerialNumber
}

func (s *SerialNumberObject) GetValue() interface{} {
	return s.serialNumber
}

func (s *SerialNumberObject) GetValueBytes() []byte {
	data := s.GetValue()
	return data.(SerialNumber)[:]
}

func (s *SerialNumberObject) GetHash() common.Hash {
	return s.serialNumberHash
}

func (s *SerialNumberObject) GetType() int {
	return s.objectType
}

//TODO: implement
func (s *SerialNumberObject) Delete() error {
	return nil
}

//TODO: implement
func (s *SerialNumberObject) Exist() bool {
	return false
}

//TODO: implement
func (s *SerialNumberObject) Reset() bool {
	return false
}

//TODO": implement
func (s *SerialNumberObject) IsDeleted() bool {
	return false
}
