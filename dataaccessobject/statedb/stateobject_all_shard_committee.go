package statedb

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"reflect"
)

type AllShardCommitteeObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	allShardCommitteeHash common.Hash
	temporaryID           []byte
	allShardCommittee     map[byte][]incognitokey.CommitteePublicKey
	objectType            int
	deleted               bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newAllShardCommitteeObject(db *StateDB, hash common.Hash) *AllShardCommitteeObject {
	return &AllShardCommitteeObject{
		db:                    db,
		allShardCommitteeHash: hash,
		allShardCommittee:     make(map[byte][]incognitokey.CommitteePublicKey),
		objectType:            AllShardCommitteeObjectType,
		deleted:               false,
	}
}
func newAllShardCommitteeObjectWithValue(db *StateDB, key common.Hash, data interface{}) *AllShardCommitteeObject {
	newAllShardCommittee, ok := data.(map[byte][]incognitokey.CommitteePublicKey)
	if !ok {
		panic("Wrong expected value")
	}
	return &AllShardCommitteeObject{
		allShardCommitteeHash: key,
		allShardCommittee:     newAllShardCommittee,
		db:                    db,
		objectType:            AllShardCommitteeObjectType,
		deleted:               false,
	}
}

// setError remembers the first non-nil error it is called with.
func (a *AllShardCommitteeObject) SetError(err error) {
	if a.dbErr == nil {
		a.dbErr = err
	}
}

func (a *AllShardCommitteeObject) GetTrie(db DatabaseAccessWarper) Trie {
	return a.trie
}

func (a *AllShardCommitteeObject) SetValue(data interface{}) {
	newAllShardCommittee, ok := data.(map[byte][]incognitokey.CommitteePublicKey)
	if !ok {
		panic("Wrong expected value")
	}
	a.allShardCommittee = newAllShardCommittee
}

func (a *AllShardCommitteeObject) GetValue() interface{} {
	return a.allShardCommittee
}

func (a *AllShardCommitteeObject) GetValueBytes() []byte {
	data := a.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal all shard committee")
	}
	return value
}

func (a *AllShardCommitteeObject) GetHash() common.Hash {
	return a.allShardCommitteeHash
}

func (a *AllShardCommitteeObject) GetType() int {
	return a.objectType
}

// MarkDelete will delete an object in trie
func (a *AllShardCommitteeObject) MarkDelete() {
	a.deleted = true
}

// reset all shard committee value into default value
func (a *AllShardCommitteeObject) Reset() bool {
	a.allShardCommittee = make(map[byte][]incognitokey.CommitteePublicKey)
	return true
}

func (a *AllShardCommitteeObject) IsDeleted() bool {
	return a.deleted
}

// value is either default or nil
func (a *AllShardCommitteeObject) IsEmpty() bool {
	temp := make(map[byte][]incognitokey.CommitteePublicKey)
	return reflect.DeepEqual(temp, a.allShardCommittee) || a.allShardCommittee == nil
}
