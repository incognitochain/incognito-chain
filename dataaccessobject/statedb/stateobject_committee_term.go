package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type CommitteeTerm struct {
	value uint64
}

func NewCommitteeTerm() *CommitteeTerm {
	return &CommitteeTerm{}
}

func NewCommitteeTermWithValue(
	value uint64,
) *CommitteeTerm {
	return &CommitteeTerm{
		value: value,
	}
}

func (ct CommitteeTerm) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Value uint64
	}{
		Value: ct.value,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (ct *CommitteeTerm) UnmarshalJSON(data []byte) error {
	temp := struct {
		Value uint64
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	ct.value = temp.Value
	return nil
}

func (ct *CommitteeTerm) SetValue(value uint64) {
	ct.value = value
}

func (ct CommitteeTerm) Value() uint64 {
	return ct.Value()
}

type CommitteeTermObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version       int
	publicKeyHash common.Hash
	committeeTerm *CommitteeTerm
	objectType    int
	deleted       bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newCommitteeTermObject(db *StateDB, publicKeyHash common.Hash) *CommitteeTermObject {
	return &CommitteeTermObject{
		version:       defaultVersion,
		db:            db,
		publicKeyHash: publicKeyHash,
		committeeTerm: &CommitteeTerm{},
		objectType:    CommitteeTermObjectType,
		deleted:       false,
	}
}

func newCommitteeTermObjectWithValue(db *StateDB, publicKeyHash common.Hash, data interface{}) (*CommitteeTermObject, error) {
	var newCommitteeTerm = NewCommitteeTerm()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newCommitteeTerm)
		if err != nil {
			return nil, err
		}
	} else {
		newCommitteeTerm, ok = data.(*CommitteeTerm)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidCommitteeTermType, reflect.TypeOf(data))
		}
	}
	return &CommitteeTermObject{
		version:       defaultVersion,
		publicKeyHash: publicKeyHash,
		committeeTerm: newCommitteeTerm,
		db:            db,
		objectType:    CommitteeObjectType,
		deleted:       false,
	}, nil
}

func (ct CommitteeTermObject) GetVersion() int {
	return ct.version
}

// setError remembers the first non-nil error it is called with.
func (ct *CommitteeTermObject) SetError(err error) {
	if ct.dbErr == nil {
		ct.dbErr = err
	}
}

func (ct CommitteeTermObject) GetTrie(db DatabaseAccessWarper) Trie {
	return ct.trie
}

func (ct *CommitteeTermObject) SetValue(data interface{}) error {
	newCommitteeTerm, ok := data.(*CommitteeTerm)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidCommitteeTermType, reflect.TypeOf(data))
	}
	ct.committeeTerm = newCommitteeTerm
	return nil
}

func (ct CommitteeTermObject) GetValue() interface{} {
	return ct.committeeTerm
}

func (ct CommitteeTermObject) GetValueBytes() []byte {
	data := ct.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal committee term object")
	}
	return value
}

func (ct CommitteeTermObject) GetPublicKeyHash() common.Hash {
	return ct.publicKeyHash
}

func (ct CommitteeTermObject) GetType() int {
	return ct.objectType
}

// MarkDelete will delete an object in trie
func (ct *CommitteeTermObject) MarkDelete() {
	ct.deleted = true
}

// reset all shard committee value into default value
func (ct *CommitteeTermObject) Reset() bool {
	ct.committeeTerm = NewCommitteeTerm()
	return true
}

func (ct CommitteeTermObject) IsDeleted() bool {
	return ct.deleted
}

// value is either default or nil
func (ct CommitteeTermObject) IsEmpty() bool {
	temp := NewCommitteeTerm()
	return reflect.DeepEqual(temp, ct.committeeTerm) || ct.committeeTerm == nil
}
