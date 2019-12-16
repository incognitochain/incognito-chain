package statedb

import (
	"github.com/incognitochain/incognito-chain/common"
)

type StateObject interface {
	GetValue() interface{}
	GetValueBytes() []byte
	GetHash() common.Hash
	GetType() int
	SetValue(interface{}) error
	GetTrie(DatabaseAccessWarper) Trie
	SetError(error)
	MarkDelete()
	IsDeleted() bool
	IsEmpty() bool
	Reset() bool
}

func newStateObjectWithValue(db *StateDB, objectType int, hash common.Hash, value interface{}) (StateObject, error) {
	switch objectType {
	case TestObjectType:
		return newTestObjectWithValue(db, hash, value)
	case SerialNumberObjectType:
		return newSerialNumberObjectWithValue(db, hash, value)
	case CommitteeObjectType:
		return newCommitteeObjectWithValue(db, hash, value)
	case RewardReceiverObjectType:
		return newRewardReceiverObjectWithValue(db, hash, value)
	default:
		panic("state object type not exist")
	}
}

func newStateObject(db *StateDB, objectType int, hash common.Hash) StateObject {
	switch objectType {
	case TestObjectType:
		return newTestObject(db, hash)
	case SerialNumberObjectType:
		return newSerialNumberObject(db, hash)
	case CommitteeObjectType:
		return newCommitteeObject(db, hash)
	case RewardReceiverObjectType:
		return newRewardReceiverObject(db, hash)
	default:
		panic("state object type not exist")
	}
}
