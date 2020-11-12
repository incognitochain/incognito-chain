package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type OutputCoinState struct {
	tokenID    common.Hash
	shardID    byte
	publicKey  []byte
	outputCoin []byte
}

func (o *OutputCoinState) OutputCoin() []byte {
	return o.outputCoin
}

func (o *OutputCoinState) SetOutputCoin(outputCoin []byte) {
	o.outputCoin = outputCoin
}

func (o OutputCoinState) TokenID() common.Hash {
	return o.tokenID
}

func (o *OutputCoinState) SetTokenID(tokenID common.Hash) {
	o.tokenID = tokenID
}

func (o OutputCoinState) ShardID() byte {
	return o.shardID
}

func (o *OutputCoinState) SetShardID(shardID byte) {
	o.shardID = shardID
}

func (o OutputCoinState) PublicKey() []byte {
	return o.publicKey
}

func (o *OutputCoinState) SetPublicKey(publicKey []byte) {
	o.publicKey = publicKey
}

func (o OutputCoinState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		TokenID    common.Hash
		ShardID    byte
		PublicKey  []byte
		OutputCoin []byte
	}{
		TokenID:    o.tokenID,
		ShardID:    o.shardID,
		PublicKey:  o.publicKey,
		OutputCoin: o.outputCoin,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (o *OutputCoinState) UnmarshalJSON(data []byte) error {
	temp := struct {
		TokenID    common.Hash
		ShardID    byte
		PublicKey  []byte
		OutputCoin []byte
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	o.tokenID = temp.TokenID
	o.shardID = temp.ShardID
	o.publicKey = temp.PublicKey
	o.outputCoin = temp.OutputCoin
	return nil
}

func NewOutputCoinStateWithValue(tokenID common.Hash, shardID byte, publicKey []byte, outputCoin []byte) *OutputCoinState {
	return &OutputCoinState{tokenID: tokenID, shardID: shardID, publicKey: publicKey, outputCoin: outputCoin}
}

func NewOutputCoinState() *OutputCoinState {
	return &OutputCoinState{}
}

type OutputCoinObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version         int
	outputCoinHash  common.Hash
	outputCoinState *OutputCoinState
	objectType      int
	deleted         bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newOutputCoinObject(db *StateDB, hash common.Hash) *OutputCoinObject {
	return &OutputCoinObject{
		version:         defaultVersion,
		db:              db,
		outputCoinHash:  hash,
		outputCoinState: NewOutputCoinState(),
		objectType:      OutputCoinObjectType,
		deleted:         false,
	}
}

func newOutputCoinObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*OutputCoinObject, error) {
	var newOutputCoinState = NewOutputCoinState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newOutputCoinState)
		if err != nil {
			return nil, err
		}
	} else {
		newOutputCoinState, ok = data.(*OutputCoinState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidOutputCoinStateType, reflect.TypeOf(data))
		}
	}
	return &OutputCoinObject{
		version:         defaultVersion,
		outputCoinHash:  key,
		outputCoinState: newOutputCoinState,
		db:              db,
		objectType:      OutputCoinObjectType,
		deleted:         false,
	}, nil
}

func GenerateOutputCoinObjectKey(tokenID common.Hash, shardID byte, publicKey []byte, outputCoin []byte) common.Hash {
	prefixHash := GetOutputCoinPrefix(tokenID, shardID, publicKey)
	valueHash := common.HashH(outputCoin)
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func GenerateReindexedOutputCoinObjectKey(tokenID common.Hash, shardID byte, publicKey []byte, outputCoin []byte) common.Hash {
	prefixHash := GetReindexedOutputCoinPrefix(tokenID, shardID, publicKey)
	valueHash := common.HashH(outputCoin)
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func GenerateReindexedOTAKeyObjectKey(theKey []byte) common.Hash {
	prefixHash := GetReindexedKeysPrefix()
	valueHash := common.HashH(theKey)
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (s OutputCoinObject) GetVersion() int {
	return s.version
}

// setError remembers the first non-nil error it is called with.
func (s *OutputCoinObject) SetError(err error) {
	if s.dbErr == nil {
		s.dbErr = err
	}
}

func (s OutputCoinObject) GetTrie(db DatabaseAccessWarper) Trie {
	return s.trie
}

func (s *OutputCoinObject) SetValue(data interface{}) error {
	newOutputCoinState, ok := data.(*OutputCoinState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidOutputCoinStateType, reflect.TypeOf(data))
	}
	s.outputCoinState = newOutputCoinState
	return nil
}

func (s OutputCoinObject) GetValue() interface{} {
	return s.outputCoinState
}

func (s OutputCoinObject) GetValueBytes() []byte {
	outputCoinState, ok := s.GetValue().(*OutputCoinState)
	if !ok {
		panic("wrong expected value type")
	}
	outputCoinStateBytes, err := json.Marshal(outputCoinState)
	if err != nil {
		panic(err.Error())
	}
	return outputCoinStateBytes
}

func (s OutputCoinObject) GetHash() common.Hash {
	return s.outputCoinHash
}

func (s OutputCoinObject) GetType() int {
	return s.objectType
}

// MarkDelete will delete an object in trie
func (s *OutputCoinObject) MarkDelete() {
	s.deleted = true
}

// Reset serial number into default
func (s *OutputCoinObject) Reset() bool {
	s.outputCoinState = NewOutputCoinState()
	return true
}

func (s OutputCoinObject) IsDeleted() bool {
	return s.deleted
}

// empty value or not
func (s OutputCoinObject) IsEmpty() bool {
	temp := NewOutputCoinState()
	return reflect.DeepEqual(temp, s.outputCoinState) || s.outputCoinState == nil
}
