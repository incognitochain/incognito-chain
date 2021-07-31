package statedb

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type CommitmentState struct {
	tokenID    common.Hash
	shardID    byte
	commitment []byte
	index      *big.Int
}

func (c CommitmentState) Index() *big.Int {
	return c.index
}

func (c *CommitmentState) SetIndex(index *big.Int) {
	c.index = index
}

func (c CommitmentState) Commitment() []byte {
	return c.commitment
}

func (c *CommitmentState) SetCommitment(commitment []byte) {
	c.commitment = commitment
}

func (c CommitmentState) ShardID() byte {
	return c.shardID
}

func (c *CommitmentState) SetShardID(shardID byte) {
	c.shardID = shardID
}

func (c CommitmentState) TokenID() common.Hash {
	return c.tokenID
}

func (c *CommitmentState) SetTokenID(tokenID common.Hash) {
	c.tokenID = tokenID
}

func NewCommitmentState() *CommitmentState {
	return &CommitmentState{}
}

func NewCommitmentStateWithValue(tokenID common.Hash, shardID byte, commitment []byte, index *big.Int) *CommitmentState {
	return &CommitmentState{
		tokenID:    tokenID,
		shardID:    shardID,
		commitment: commitment,
		index:      index,
	}
}

func (c CommitmentState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		TokenID    common.Hash
		ShardID    byte
		Commitment []byte
		PublicKey  []byte
		Additional []byte
		Index      *big.Int
	}{
		TokenID:    c.tokenID,
		ShardID:    c.shardID,
		Commitment: c.commitment,
		Index:      c.index,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (c *CommitmentState) UnmarshalJSON(data []byte) error {
	temp := struct {
		TokenID    common.Hash
		ShardID    byte
		Commitment []byte
		PublicKey  []byte
		Additional []byte
		Index      *big.Int
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	c.tokenID = temp.TokenID
	c.shardID = temp.ShardID
	c.commitment = temp.Commitment
	c.index = temp.Index
	return nil
}

type CommitmentObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version         int
	commitmentHash  common.Hash
	commitmentState *CommitmentState
	objectType      int
	deleted         bool
	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newCommitmentObject(db *StateDB, hash common.Hash) *CommitmentObject {
	return &CommitmentObject{
		version:         defaultVersion,
		db:              db,
		commitmentHash:  hash,
		commitmentState: NewCommitmentState(),
		objectType:      CommitmentObjectType,
		deleted:         false,
	}
}

func newCommitmentObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*CommitmentObject, error) {
	var newCommitmentState = NewCommitmentState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newCommitmentState)
		if err != nil {
			return nil, err
		}
	} else {
		newCommitmentState, ok = data.(*CommitmentState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidCommitmentStateType, reflect.TypeOf(data))
		}
	}
	return &CommitmentObject{
		version:         defaultVersion,
		commitmentHash:  key,
		commitmentState: newCommitmentState,
		db:              db,
		objectType:      CommitmentObjectType,
		deleted:         false,
	}, nil
}

func GenerateCommitmentObjectKey(tokenID common.Hash, shardID byte, commitment []byte) common.Hash {
	prefixHash := GetCommitmentPrefix(tokenID, shardID)
	valueHash := common.HashH(commitment)
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (s CommitmentObject) GetVersion() int {
	return s.version
}

// setError remembers the first non-nil error it is called with.
func (s *CommitmentObject) SetError(err error) {
	if s.dbErr == nil {
		s.dbErr = err
	}
}

func (s CommitmentObject) GetTrie(db DatabaseAccessWarper) Trie {
	return s.trie
}

func (s *CommitmentObject) SetValue(data interface{}) error {
	newCommitmentState, ok := data.(*CommitmentState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidCommitmentStateType, reflect.TypeOf(data))
	}
	s.commitmentState = newCommitmentState
	return nil
}

func (s CommitmentObject) GetValue() interface{} {
	return s.commitmentState
}

func (s CommitmentObject) GetValueBytes() []byte {
	commitmentState, ok := s.GetValue().(*CommitmentState)
	if !ok {
		panic("wrong expected value type")
	}
	commitmentStateBytes, err := json.Marshal(commitmentState)
	if err != nil {
		panic(err.Error())
	}
	return commitmentStateBytes
}

func (s CommitmentObject) GetHash() common.Hash {
	return s.commitmentHash
}

func (s CommitmentObject) GetType() int {
	return s.objectType
}

// MarkDelete will delete an object in trie
func (s *CommitmentObject) MarkDelete() {
	s.deleted = true
}

// Reset serial number into default
func (s *CommitmentObject) Reset() bool {
	s.commitmentState = NewCommitmentState()
	return true
}

func (s CommitmentObject) IsDeleted() bool {
	return s.deleted
}

// empty value or not
func (s CommitmentObject) IsEmpty() bool {
	temp := NewCommitmentState()
	return reflect.DeepEqual(temp, s.commitmentState) || s.commitmentState == nil
}

//========================================================== INDEX
type CommitmentIndexObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version             int
	commitmentIndexHash common.Hash
	commitmentHash      common.Hash
	objectType          int
	deleted             bool
	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newCommitmentIndexObject(db *StateDB, hash common.Hash) *CommitmentIndexObject {
	return &CommitmentIndexObject{
		version:             defaultVersion,
		db:                  db,
		commitmentIndexHash: hash,
		commitmentHash:      common.Hash{},
		objectType:          CommitmentIndexObjectType,
		deleted:             false,
	}
}

func newCommitmentIndexObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*CommitmentIndexObject, error) {
	var newCommitmentIndexState = common.Hash{}
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := newCommitmentIndexState.SetBytes(dataBytes)
		if err != nil {
			return nil, err
		}
	} else {
		newCommitmentIndexState, ok = data.(common.Hash)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidHashType, reflect.TypeOf(data))
		}
	}
	return &CommitmentIndexObject{
		version:             defaultVersion,
		commitmentIndexHash: key,
		commitmentHash:      newCommitmentIndexState,
		db:                  db,
		objectType:          CommitmentIndexObjectType,
		deleted:             false,
	}, nil
}

func GenerateCommitmentIndexObjectKey(tokenID common.Hash, shardID byte, index *big.Int) common.Hash {
	prefixHash := GetCommitmentIndexPrefix(tokenID, shardID)
	valueHash := common.HashH(index.Bytes())
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (s CommitmentIndexObject) GetVersion() int {
	return s.version
}

// setError remembers the first non-nil error it is called with.
func (s *CommitmentIndexObject) SetError(err error) {
	if s.dbErr == nil {
		s.dbErr = err
	}
}

func (s CommitmentIndexObject) GetTrie(db DatabaseAccessWarper) Trie {
	return s.trie
}

func (s *CommitmentIndexObject) SetValue(data interface{}) error {
	newCommitmentIndexState, ok := data.(common.Hash)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidHashType, reflect.TypeOf(data))
	}
	s.commitmentHash = newCommitmentIndexState
	return nil
}

func (s CommitmentIndexObject) GetValue() interface{} {
	return s.commitmentHash
}

func (s CommitmentIndexObject) GetValueBytes() []byte {
	temp := s.GetValue().(common.Hash)
	return temp.Bytes()
}

func (s CommitmentIndexObject) GetHash() common.Hash {
	return s.commitmentIndexHash
}

func (s CommitmentIndexObject) GetType() int {
	return s.objectType
}

// MarkDelete will delete an object in trie
func (s *CommitmentIndexObject) MarkDelete() {
	s.deleted = true
}

// Reset serial number into default
func (s *CommitmentIndexObject) Reset() bool {
	s.commitmentHash = common.Hash{}
	return true
}

func (s CommitmentIndexObject) IsDeleted() bool {
	return s.deleted
}

// empty value or not
func (s CommitmentIndexObject) IsEmpty() bool {
	temp := common.Hash{}
	return reflect.DeepEqual(temp, s.commitmentHash)
}

//========================================================== Length
type CommitmentLengthObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version              int
	commitmentLengthHash common.Hash
	commitmentLength     *big.Int
	objectType           int
	deleted              bool
	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newCommitmentLengthObject(db *StateDB, hash common.Hash) *CommitmentLengthObject {
	return &CommitmentLengthObject{
		version:              defaultVersion,
		db:                   db,
		commitmentLengthHash: hash,
		commitmentLength:     new(big.Int).SetUint64(0),
		objectType:           CommitmentLengthObjectType,
		deleted:              false,
	}
}

func newCommitmentLengthObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*CommitmentLengthObject, error) {
	var newCommitmentLengthValue = new(big.Int)
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		// if bytes.Compare(dataBytes, zeroBigInt) == 0 {
		// 	newCommitmentLengthValue.SetUint64(0)
		// } else {
		newCommitmentLengthValue.SetBytes(dataBytes)
		// }
	} else {
		newCommitmentLengthValue, ok = data.(*big.Int)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBigIntType, reflect.TypeOf(data))
		}
	}
	return &CommitmentLengthObject{
		version:              defaultVersion,
		commitmentLengthHash: key,
		commitmentLength:     newCommitmentLengthValue,
		db:                   db,
		objectType:           CommitmentLengthObjectType,
		deleted:              false,
	}, nil
}

func GenerateCommitmentLengthObjectKey(tokenID common.Hash, shardID byte) common.Hash {
	prefixHash := GetCommitmentLengthPrefix()
	valueHash := common.HashH(append(tokenID[:], shardID))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (s CommitmentLengthObject) GetVersion() int {
	return s.version
}

// setError remembers the first non-nil error it is called with.
func (s *CommitmentLengthObject) SetError(err error) {
	if s.dbErr == nil {
		s.dbErr = err
	}
}

func (s CommitmentLengthObject) GetTrie(db DatabaseAccessWarper) Trie {
	return s.trie
}

func (s *CommitmentLengthObject) SetValue(data interface{}) error {
	newCommitmentLengthValue, ok := data.(*big.Int)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidBigIntType, reflect.TypeOf(data))
	}
	s.commitmentLength = newCommitmentLengthValue
	return nil
}

func (s CommitmentLengthObject) GetValue() interface{} {
	return s.commitmentLength
}

func (s CommitmentLengthObject) GetValueBytes() []byte {
	if s.GetValue().(*big.Int).Uint64() == 0 {
		return []byte{0}
	}
	return s.GetValue().(*big.Int).Bytes()
}

func (s CommitmentLengthObject) GetHash() common.Hash {
	return s.commitmentLengthHash
}

func (s CommitmentLengthObject) GetType() int {
	return s.objectType
}

// MarkDelete will delete an object in trie
func (s *CommitmentLengthObject) MarkDelete() {
	s.deleted = true
}

// Reset serial number into default
func (s *CommitmentLengthObject) Reset() bool {
	s.commitmentLength = new(big.Int).SetUint64(0)
	return true
}

func (s CommitmentLengthObject) IsDeleted() bool {
	return s.deleted
}

// empty value or not
func (s CommitmentLengthObject) IsEmpty() bool {
	return s.commitmentLength == nil
}
