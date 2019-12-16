package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type CommitteeState struct {
	ShardID            int
	CommitteePublicKey incognitokey.CommitteePublicKey
}

func NewCommitteeState() *CommitteeState {
	return &CommitteeState{}
}
func NewCommitteeStateWithValue(shardID int, committeePublicKey incognitokey.CommitteePublicKey) *CommitteeState {
	return &CommitteeState{ShardID: shardID, CommitteePublicKey: committeePublicKey}
}

func (c *CommitteeState) GetCommitteePublicKey() incognitokey.CommitteePublicKey {
	return c.CommitteePublicKey
}

func (c *CommitteeState) SetCommitteePublicKey(committeePublicKey incognitokey.CommitteePublicKey) {
	c.CommitteePublicKey = committeePublicKey
}
func (c *CommitteeState) GetShardID() int {
	return c.ShardID
}

func (c *CommitteeState) SetShardID(shardID int) {
	c.ShardID = shardID
}

type CommitteeObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	committeePublicKeyHash common.Hash
	committeeState         *CommitteeState
	objectType             int
	deleted                bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newCommitteeObject(db *StateDB, hash common.Hash) *CommitteeObject {
	return &CommitteeObject{
		db:                     db,
		committeePublicKeyHash: hash,
		committeeState:         NewCommitteeState(),
		objectType:             CommitteeObjectType,
		deleted:                false,
	}
}
func newCommitteeObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*CommitteeObject, error) {
	var newCommitteeState = NewCommitteeState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newCommitteeState)
		if err != nil {
			return nil, NewStatedbError(InvalidCommitteeStateTypeError, err)
		}
	} else {
		newCommitteeState, ok = data.(*CommitteeState)
		if !ok {
			return nil, NewStatedbError(InvalidCommitteeStateTypeError, fmt.Errorf("%+v", reflect.TypeOf(data)))
		}
	}
	return &CommitteeObject{
		committeePublicKeyHash: key,
		committeeState:         NewCommitteeStateWithValue(newCommitteeState.ShardID, newCommitteeState.CommitteePublicKey),
		db:                     db,
		objectType:             CommitteeObjectType,
		deleted:                false,
	}, nil
}

func GenerateCommitteeObjectKey(shardID int, committee incognitokey.CommitteePublicKey) (common.Hash, error) {
	committeeBytes, err := committee.Bytes()
	if err != nil {
		return common.Hash{}, NewStatedbError(InvalidCommitteeStateTypeError, err)
	}
	prefixHash := GetCommitteePrefixByShardID(shardID)
	valueHash := common.HashH(committeeBytes)
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...)), nil
}

// setError remembers the first non-nil error it is called with.
func (c *CommitteeObject) SetError(err error) {
	if c.dbErr == nil {
		c.dbErr = err
	}
}

func (c *CommitteeObject) GetTrie(db DatabaseAccessWarper) Trie {
	return c.trie
}

func (c *CommitteeObject) SetValue(data interface{}) error {
	newCommitteeState, ok := data.(*CommitteeState)
	if !ok {
		return NewStatedbError(InvalidCommitteeStateTypeError, fmt.Errorf("%+v", reflect.TypeOf(data)))
	}
	c.committeeState = newCommitteeState
	return nil
}

func (c *CommitteeObject) GetValue() interface{} {
	return c.committeeState
}

func (c *CommitteeObject) GetValueBytes() []byte {
	data := c.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal all shard committee")
	}
	return value
}

func (c *CommitteeObject) GetHash() common.Hash {
	return c.committeePublicKeyHash
}

func (c *CommitteeObject) GetType() int {
	return c.objectType
}

// MarkDelete will delete an object in trie
func (c *CommitteeObject) MarkDelete() {
	c.deleted = true
}

// reset all shard committee value into default value
func (c *CommitteeObject) Reset() bool {
	c.committeeState = NewCommitteeState()
	return true
}

func (c *CommitteeObject) IsDeleted() bool {
	return c.deleted
}

// value is either default or nil
func (c *CommitteeObject) IsEmpty() bool {
	temp := NewCommitteeState()
	return reflect.DeepEqual(temp, c.committeeState) || c.committeeState == nil
}
