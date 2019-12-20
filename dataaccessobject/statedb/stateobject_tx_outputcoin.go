package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

type OutputCoinState struct {
	tokenID     common.Hash
	shardID     byte
	publicKey   []byte
	outputCoins [][]byte
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

func (o OutputCoinState) OutputCoins() [][]byte {
	return o.outputCoins
}

func (o *OutputCoinState) SetOutputCoins(outputCoins [][]byte) {
	o.outputCoins = outputCoins
}

func (o OutputCoinState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		TokenID     common.Hash
		ShardID     byte
		PublicKey   []byte
		OutputCoins [][]byte
	}{
		TokenID:     o.tokenID,
		ShardID:     o.shardID,
		PublicKey:   o.publicKey,
		OutputCoins: o.outputCoins,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (o *OutputCoinState) UnmarshalJSON(data []byte) error {
	temp := struct {
		TokenID     common.Hash
		ShardID     byte
		PublicKey   []byte
		OutputCoins [][]byte
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	o.tokenID = temp.TokenID
	o.shardID = temp.ShardID
	o.publicKey = temp.PublicKey
	o.outputCoins = temp.OutputCoins
	return nil
}
func NewOutputCoinStateWithValue(tokenID common.Hash, shardID byte, publicKey []byte, outputCoins [][]byte) *OutputCoinState {
	return &OutputCoinState{tokenID: tokenID, shardID: shardID, publicKey: publicKey, outputCoins: outputCoins}
}

func NewOutputCoinState() *OutputCoinState {
	return &OutputCoinState{}
}

type OutputCoinObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version          int
	snDerivatorHash  common.Hash
	snDerivatorState *OutputCoinState
	objectType       int
	deleted          bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newOutputCoinObject(db *StateDB, hash common.Hash) *OutputCoinObject {
	return &OutputCoinObject{
		version:          defaultVersion,
		db:               db,
		snDerivatorHash:  hash,
		snDerivatorState: NewOutputCoinState(),
		objectType:       OutputCoinObjectType,
		deleted:          false,
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
		version:          defaultVersion,
		snDerivatorHash:  key,
		snDerivatorState: newOutputCoinState,
		db:               db,
		objectType:       OutputCoinObjectType,
		deleted:          false,
	}, nil
}

func GenerateOutputCoinObjectKey(tokenID common.Hash, shardID byte, publicKey []byte) common.Hash {
	prefixHash := GetOutputCoinPrefix(tokenID, shardID)
	valueHash := common.HashH(publicKey)
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
	s.snDerivatorState = newOutputCoinState
	return nil
}

func (s OutputCoinObject) GetValue() interface{} {
	return s.snDerivatorState
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
	return s.snDerivatorHash
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
	s.snDerivatorState = NewOutputCoinState()
	return true
}

func (s OutputCoinObject) IsDeleted() bool {
	return s.deleted
}

// empty value or not
func (s OutputCoinObject) IsEmpty() bool {
	temp := NewOutputCoinState()
	return reflect.DeepEqual(temp, s.snDerivatorState) || s.snDerivatorState == nil
}
