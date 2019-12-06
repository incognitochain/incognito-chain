package statedb

import (
	"github.com/incognitochain/incognito-chain/common"
)

type StateObject interface {
	GetValue() interface{}
	GetValueBytes() []byte
	GetHash() common.Hash
	GetType() int
	SetValue(interface{})
	GetTrie(DatabaseAccessWarper) Trie
	SetError(error)
	MarkDelete()
	Exist() bool
	Reset() bool
	IsDeleted() bool
	Empty() bool
}

func newStateObjectWithValue(db *StateDB, objectType int, hash common.Hash, value interface{}) StateObject {
	switch objectType {
	case SerialNumberObjectType:
		return newSerialNumberObjectWithValue(db, hash, value)
	default:
		panic("state object type not exist")
	}
}

func newStateObject(db *StateDB, objectType int, hash common.Hash) StateObject {
	switch objectType {
	case SerialNumberObjectType:
		return newSerialNumberObject(db, hash)
	default:
		panic("state object type not exist")
	}
}
