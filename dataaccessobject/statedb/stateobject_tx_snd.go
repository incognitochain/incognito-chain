package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

type SNDerivatorState struct {
	tokenID common.Hash
	snd     []byte
}

func (s SNDerivatorState) Snd() []byte {
	return s.snd
}

func (s *SNDerivatorState) SetSnd(snd []byte) {
	s.snd = snd
}

func (s SNDerivatorState) TokenID() common.Hash {
	return s.tokenID
}

func (s *SNDerivatorState) SetTokenID(tokenID common.Hash) {
	s.tokenID = tokenID
}

func (s SNDerivatorState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		TokenID common.Hash
		Snd     []byte
	}{
		TokenID: s.tokenID,
		Snd:     s.snd,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (s *SNDerivatorState) UnmarshalJSON(data []byte) error {
	temp := struct {
		TokenID common.Hash
		Snd     []byte
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	s.tokenID = temp.TokenID
	s.snd = temp.Snd
	return nil
}

func NewSNDerivatorState() *SNDerivatorState {
	return &SNDerivatorState{}
}

func NewSNDerivatorStateWithValue(tokenID common.Hash, snd []byte) *SNDerivatorState {
	return &SNDerivatorState{tokenID: tokenID, snd: snd}
}

type SNDerivatorObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version          int
	snDerivatorHash  common.Hash
	snDerivatorState *SNDerivatorState
	objectType       int
	deleted          bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newSNDerivatorObject(db *StateDB, hash common.Hash) *SNDerivatorObject {
	return &SNDerivatorObject{
		version:          defaultVersion,
		db:               db,
		snDerivatorHash:  hash,
		snDerivatorState: NewSNDerivatorState(),
		objectType:       SNDerivatorObjectType,
		deleted:          false,
	}
}

func newSNDerivatorObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*SNDerivatorObject, error) {
	var newSNDerivatorState = NewSNDerivatorState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newSNDerivatorState)
		if err != nil {
			return nil, err
		}
	} else {
		newSNDerivatorState, ok = data.(*SNDerivatorState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidSNDerivatorStateType, reflect.TypeOf(data))
		}
	}
	return &SNDerivatorObject{
		version:          defaultVersion,
		snDerivatorHash:  key,
		snDerivatorState: newSNDerivatorState,
		db:               db,
		objectType:       SNDerivatorObjectType,
		deleted:          false,
	}, nil
}

func GenerateSNDerivatorObjectKey(tokenID common.Hash, snd []byte) common.Hash {
	prefixHash := GetSNDerivatorPrefix(tokenID)
	valueHash := common.HashH(snd)
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (s SNDerivatorObject) GetVersion() int {
	return s.version
}

// setError remembers the first non-nil error it is called with.
func (s *SNDerivatorObject) SetError(err error) {
	if s.dbErr == nil {
		s.dbErr = err
	}
}

func (s SNDerivatorObject) GetTrie(db DatabaseAccessWarper) Trie {
	return s.trie
}

func (s *SNDerivatorObject) SetValue(data interface{}) error {
	newSNDerivatorState, ok := data.(*SNDerivatorState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidSNDerivatorStateType, reflect.TypeOf(data))
	}
	s.snDerivatorState = newSNDerivatorState
	return nil
}

func (s SNDerivatorObject) GetValue() interface{} {
	return s.snDerivatorState
}

func (s SNDerivatorObject) GetValueBytes() []byte {
	sndState, ok := s.GetValue().(*SNDerivatorState)
	if !ok {
		panic("wrong expected value type")
	}
	sndStateBytes, err := json.Marshal(sndState)
	if err != nil {
		panic(err.Error())
	}
	return sndStateBytes
}

func (s SNDerivatorObject) GetHash() common.Hash {
	return s.snDerivatorHash
}

func (s SNDerivatorObject) GetType() int {
	return s.objectType
}

// MarkDelete will delete an object in trie
func (s *SNDerivatorObject) MarkDelete() {
	s.deleted = true
}

// Reset serial number into default
func (s *SNDerivatorObject) Reset() bool {
	s.snDerivatorState = NewSNDerivatorState()
	return true
}

func (s SNDerivatorObject) IsDeleted() bool {
	return s.deleted
}

// empty value or not
func (s SNDerivatorObject) IsEmpty() bool {
	temp := NewSNDerivatorState()
	return reflect.DeepEqual(temp, s.snDerivatorState) || s.snDerivatorState == nil
}
