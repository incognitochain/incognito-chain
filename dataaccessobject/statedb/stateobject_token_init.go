package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

type TokenInitState struct {
	tokenID     string
	tokenName   string
	tokenSymbol string
	amount      uint64
}

func (s TokenInitState) TokenID() string {
	return s.tokenID
}
func (s TokenInitState) TokenName() string {
	return s.tokenName
}
func (s TokenInitState) TokenSymbol() string {
	return s.tokenSymbol
}
func (s TokenInitState) Amount() uint64 {
	return s.amount
}

func (s *TokenInitState) SetTokenID(tokenID string) {
	s.tokenID = tokenID
}
func (s *TokenInitState) SetTokenName(tokenName string) {
	s.tokenName = tokenName
}
func (s *TokenInitState) SetTokenSymbol(tokenSymbol string) {
	s.tokenSymbol = tokenSymbol
}
func (s *TokenInitState) SetAmount(amount uint64) {
	s.amount = amount
}

func (s TokenInitState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		TokenID string
		Amount  uint64
	}{
		TokenID: s.tokenID,
		Amount:  s.amount,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}
func (s *TokenInitState) UnmarshalJSON(data []byte) error {
	temp := struct {
		TokenID string
		Amount  uint64
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	s.tokenID = temp.TokenID
	s.amount = temp.Amount
	return nil
}

func NewTokenInitState() *TokenInitState {
	return &TokenInitState{}
}
func NewTokenInitStateWithValue(
	tokenID string,
	tokenName string,
	tokenSymbol string,
	amount uint64,
) *TokenInitState {
	return &TokenInitState{
		tokenID:     tokenID,
		tokenName:   tokenName,
		tokenSymbol: tokenSymbol,
		amount:      amount,
	}
}


type TokenInitObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version        int
	tokenInitHash  common.Hash
	tokenInitState *TokenInitState
	objectType     int
	deleted        bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newTokenInitObject(db *StateDB, hash common.Hash) *TokenInitObject {
	return &TokenInitObject{
		version:        defaultVersion,
		db:             db,
		tokenInitHash:  hash,
		tokenInitState: NewTokenInitState(),
		objectType:     TokenInitObjectType,
		deleted:        false,
	}
}

func newTokenInitObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*TokenInitObject, error) {
	var newTokenInitState = NewTokenInitState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newTokenInitState)
		if err != nil {
			return nil, err
		}
	} else {
		newTokenInitState, ok = data.(*TokenInitState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidTokenInitStateType, reflect.TypeOf(data))
		}
	}
	return &TokenInitObject{
		version:        defaultVersion,
		tokenInitHash:  key,
		tokenInitState: newTokenInitState,
		db:             db,
		objectType:     TokenInitObjectType,
		deleted:        false,
	}, nil
}

func GenerateTokenInitObjectKey(tokenID string) common.Hash {
	prefixHash := GetTokenInitPrefix()
	valueHash := common.HashH([]byte(tokenID))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (t TokenInitObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *TokenInitObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t TokenInitObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *TokenInitObject) SetValue(data interface{}) error {
	var newTokenInitState = NewTokenInitState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newTokenInitState)
		if err != nil {
			return err
		}
	} else {
		newTokenInitState, ok = data.(*TokenInitState)
		if !ok {
			return fmt.Errorf("%+v, got type %+v", ErrInvalidTokenInitStateType, reflect.TypeOf(data))
		}
	}
	t.tokenInitState = newTokenInitState
	return nil
}

func (t TokenInitObject) GetValue() interface{} {
	return t.tokenInitState
}

func (t TokenInitObject) GetValueBytes() []byte {
	tokenInitState, ok := t.GetValue().(*TokenInitState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(tokenInitState)
	if err != nil {
		panic("failed to marshal TokenInitState")
	}
	return value
}

func (t TokenInitObject) GetHash() common.Hash {
	return t.tokenInitHash
}

func (t TokenInitObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *TokenInitObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *TokenInitObject) Reset() bool {
	t.tokenInitState = NewTokenInitState()
	return true
}

func (t TokenInitObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t TokenInitObject) IsEmpty() bool {
	temp := NewTokenInitState()
	return reflect.DeepEqual(temp, t.tokenInitState) || t.tokenInitState == nil
}