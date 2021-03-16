package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type SlashingCommitteeState struct {
	shardID    byte
	epoch      uint64
	committees []string
}

func NewSlashingCommitteeState() *SlashingCommitteeState {
	return &SlashingCommitteeState{}
}

func NewSlashingCommitteeStateWithValue(shardID byte, epoch uint64, committees []string) *SlashingCommitteeState {
	return &SlashingCommitteeState{shardID: shardID, epoch: epoch, committees: committees}
}

func (s *SlashingCommitteeState) Committees() []string {
	return s.committees
}

func (s *SlashingCommitteeState) SetCommittees(committees []string) {
	s.committees = committees
}

func (s *SlashingCommitteeState) ShardID() byte {
	return s.shardID
}

func (s *SlashingCommitteeState) SetShardID(shardID byte) {
	s.shardID = shardID
}

func (s *SlashingCommitteeState) Epoch() uint64 {
	return s.epoch
}

func (s *SlashingCommitteeState) SetEpoch(epoch uint64) {
	s.epoch = epoch
}

func (s SlashingCommitteeState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		ShardID    byte
		Epoch      uint64
		Committees []string
	}{
		ShardID:    s.shardID,
		Epoch:      s.epoch,
		Committees: s.committees,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (s *SlashingCommitteeState) UnmarshalJSON(data []byte) error {
	temp := struct {
		ShardID    byte
		Epoch      uint64
		Committees []string
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	s.shardID = temp.ShardID
	s.epoch = temp.Epoch
	s.committees = temp.Committees
	return nil
}

type SlashingCommitteeObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version                   int
	slashingCommitteeObjecKey common.Hash
	slashingCommitteeState    *SlashingCommitteeState
	objectType                int
	deleted                   bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newSlashingCommitteeObject(db *StateDB, hash common.Hash) *SlashingCommitteeObject {
	return &SlashingCommitteeObject{
		version:                   defaultVersion,
		db:                        db,
		slashingCommitteeObjecKey: hash,
		slashingCommitteeState:    NewSlashingCommitteeState(),
		objectType:                SlashingCommitteeObjectType,
		deleted:                   false,
	}
}

func newSlashingCommitteeObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*SlashingCommitteeObject, error) {
	var newSlashingCommitteeState = NewSlashingCommitteeState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newSlashingCommitteeState)
		if err != nil {
			return nil, err
		}
	} else {
		newSlashingCommitteeState, ok = data.(*SlashingCommitteeState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidSlasingCommitteeStateType, reflect.TypeOf(data))
		}
	}
	return &SlashingCommitteeObject{
		version:                   defaultVersion,
		slashingCommitteeObjecKey: key,
		slashingCommitteeState:    newSlashingCommitteeState,
		db:                        db,
		objectType:                SlashingCommitteeObjectType,
		deleted:                   false,
	}, nil
}

func GenerateSlashingCommitteeObjectKey(shardID byte, epoch uint64) common.Hash {
	prefixHash := GetSlashingCommitteePrefix(epoch)
	valueHash := common.HashH(append([]byte{shardID}))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (c SlashingCommitteeObject) GetVersion() int {
	return c.version
}

// setError remembers the first non-nil error it is called with.
func (c *SlashingCommitteeObject) SetError(err error) {
	if c.dbErr == nil {
		c.dbErr = err
	}
}

func (c SlashingCommitteeObject) GetTrie(db DatabaseAccessWarper) Trie {
	return c.trie
}

func (c *SlashingCommitteeObject) SetValue(data interface{}) error {
	newSlashingCommitteeState, ok := data.(*SlashingCommitteeState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidCommitteeStateType, reflect.TypeOf(data))
	}
	c.slashingCommitteeState = newSlashingCommitteeState
	return nil
}

func (c SlashingCommitteeObject) GetValue() interface{} {
	return c.slashingCommitteeState
}

func (c SlashingCommitteeObject) GetValueBytes() []byte {
	data := c.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal all shard committee")
	}
	return value
}

func (c SlashingCommitteeObject) GetHash() common.Hash {
	return c.slashingCommitteeObjecKey
}

func (c SlashingCommitteeObject) GetType() int {
	return c.objectType
}

// MarkDelete will delete an object in trie
func (c *SlashingCommitteeObject) MarkDelete() {
	c.deleted = true
}

// reset all shard committee value into default value
func (c *SlashingCommitteeObject) Reset() bool {
	c.slashingCommitteeState = NewSlashingCommitteeState()
	return true
}

func (c SlashingCommitteeObject) IsDeleted() bool {
	return c.deleted
}

// value is either default or nil
func (c SlashingCommitteeObject) IsEmpty() bool {
	temp := NewSlashingCommitteeState()
	return reflect.DeepEqual(temp, c.slashingCommitteeState) || c.slashingCommitteeState == nil
}
