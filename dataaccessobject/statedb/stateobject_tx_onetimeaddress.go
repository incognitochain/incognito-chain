package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

type OnetimeAddressState struct {
	tokenID     common.Hash
	shardID     byte
	height		[]byte
	outputCoin	[]byte
}

func (ota OnetimeAddressState) OutputCoin() []byte {
	return ota.outputCoin
}

func (ota *OnetimeAddressState) SetOutputCoin(outputCoin []byte) {
	ota.outputCoin = outputCoin
}

func (ota OnetimeAddressState) TokenID() common.Hash {
	return ota.tokenID
}

func (ota *OnetimeAddressState) SetTokenID(tokenID common.Hash) {
	ota.tokenID = tokenID
}

func (ota OnetimeAddressState) ShardID() byte {
	return ota.shardID
}

func (ota *OnetimeAddressState) SetShardID(shardID byte) {
	ota.shardID = shardID
}

func (ota OnetimeAddressState) Height() []byte {
	return ota.height
}

func (ota *OnetimeAddressState) SetHeight(height []byte) {
	ota.height = height
}

func (ota *OnetimeAddressState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		TokenID    common.Hash
		ShardID    byte
		Height     []byte
		OutputCoin []byte
	}{
		TokenID:    ota.tokenID,
		ShardID:    ota.shardID,
		Height:  ota.height,
		OutputCoin: ota.outputCoin,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (ota *OnetimeAddressState) UnmarshalJSON(data []byte) error {
	temp := struct {
		TokenID    common.Hash
		ShardID    byte
		Height     []byte
		OutputCoin []byte
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	ota.tokenID = temp.TokenID
	ota.shardID = temp.ShardID
	ota.height = temp.Height
	ota.outputCoin = temp.OutputCoin
	return nil
}

func NewOnetimeAddressStateWithValue(tokenID common.Hash, shardID byte, height []byte, outputCoin []byte) *OnetimeAddressState {
	return &OnetimeAddressState{tokenID: tokenID, shardID: shardID, height: height, outputCoin: outputCoin}
}

func NewOnetimeAddressState() *OnetimeAddressState {
	return &OnetimeAddressState{}
}

type OnetimeAddressObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version          int
	onetimeAddressHash  common.Hash
	onetimeAddressState *OnetimeAddressState
	objectType       int
	deleted          bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newOnetimeAddressObject(db *StateDB, hash common.Hash) *OnetimeAddressObject {
	return &OnetimeAddressObject{
		version:          defaultVersion,
		db:               db,
		onetimeAddressHash:  hash,
		onetimeAddressState: NewOnetimeAddressState(),
		objectType:       OnetimeAddressObjectType,
		deleted:          false,
	}
}

func newOnetimeAddressObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*OnetimeAddressObject, error) {
	var newOnetimeAddressState = NewOnetimeAddressState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newOnetimeAddressState)
		if err != nil {
			return nil, err
		}
	} else {
		newOnetimeAddressState, ok = data.(*OnetimeAddressState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidOnetimeAddressStateType, reflect.TypeOf(data))
		}
	}
	return &OnetimeAddressObject{
		version:          defaultVersion,
		onetimeAddressHash:  key,
		onetimeAddressState: newOnetimeAddressState,
		db:               db,
		objectType:       OnetimeAddressObjectType,
		deleted:          false,
	}, nil
}

func GenerateOnetimeAddressObjectKey(tokenID common.Hash, shardID byte, height []byte, outputCoin []byte) common.Hash {
	prefixHash := GetOnetimeAddressPrefix(tokenID, shardID, height)
	valueHash := common.HashH(outputCoin)
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (s OnetimeAddressObject) GetVersion() int {
	return s.version
}

// setError remembers the first non-nil error it is called with.
func (s *OnetimeAddressObject) SetError(err error) {
	if s.dbErr == nil {
		s.dbErr = err
	}
}

func (s OnetimeAddressObject) GetTrie(db DatabaseAccessWarper) Trie {
	return s.trie
}

func (s *OnetimeAddressObject) SetValue(data interface{}) error {
	newOnetimeAddressState, ok := data.(*OnetimeAddressState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidOnetimeAddressStateType, reflect.TypeOf(data))
	}
	s.onetimeAddressState = newOnetimeAddressState
	return nil
}

func (s OnetimeAddressObject) GetValue() interface{} {
	return s.onetimeAddressState
}

func (s OnetimeAddressObject) GetValueBytes() []byte {
	onetimeAddressState, ok := s.GetValue().(*OnetimeAddressState)
	if !ok {
		panic("wrong expected value type")
	}
	onetimeAddressStateBytes, err := json.Marshal(onetimeAddressState)
	if err != nil {
		panic(err.Error())
	}
	return onetimeAddressStateBytes
}

func (s OnetimeAddressObject) GetHash() common.Hash {
	return s.onetimeAddressHash
}

func (s OnetimeAddressObject) GetType() int {
	return s.objectType
}

// MarkDelete will delete an object in trie
func (s *OnetimeAddressObject) MarkDelete() {
	s.deleted = true
}

// Reset serial number into default
func (s *OnetimeAddressObject) Reset() bool {
	s.onetimeAddressState = NewOnetimeAddressState()
	return true
}

func (s OnetimeAddressObject) IsDeleted() bool {
	return s.deleted
}

// empty value or not
func (s OnetimeAddressObject) IsEmpty() bool {
	temp := NewOnetimeAddressState()
	return reflect.DeepEqual(temp, s.onetimeAddressState) || s.onetimeAddressState == nil
}
