package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type PTokenInitState struct {
	tokenID     string
	tokenName   string
	tokenSymbol string
	amount      uint64
}

func (s PTokenInitState) TokenID() string {
	return s.tokenID
}

func (s *PTokenInitState) SetTokenID(tokenID string) {
	s.tokenID = tokenID
}

func (s PTokenInitState) TokenName() string {
	return s.tokenName
}

func (s *PTokenInitState) SetTokenName(tokenName string) {
	s.tokenName = tokenName
}

func (s PTokenInitState) TokenSymbol() string {
	return s.tokenSymbol
}

func (s *PTokenInitState) SetTokenSymbol(tokenSymbol string) {
	s.tokenSymbol = tokenSymbol
}

func (s PTokenInitState) Amount() uint64 {
	return s.amount
}

func (s *PTokenInitState) SetAmount(amount uint64) {
	s.amount = amount
}

func (s PTokenInitState) MarshalJSON() ([]byte, error) {
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

func (s *PTokenInitState) UnmarshalJSON(data []byte) error {
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

func NewPTokenInitState() *PTokenInitState {
	return &PTokenInitState{}
}

func NewPTokenInitStateWithValue(
	tokenID string,
	tokenName string,
	tokenSymbol string,
	amount uint64,
) *PTokenInitState {
	return &PTokenInitState{
		tokenID:     tokenID,
		tokenName:   tokenName,
		tokenSymbol: tokenSymbol,
		amount:      amount,
	}
}

type PTokenInitObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version         int
	pTokenInitHash  common.Hash
	pTokenInitState *PTokenInitState
	objectType      int
	deleted         bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPTokenInitObject(db *StateDB, hash common.Hash) *PTokenInitObject {
	return &PTokenInitObject{
		version:         defaultVersion,
		db:              db,
		pTokenInitHash:  hash,
		pTokenInitState: NewPTokenInitState(),
		objectType:      PTokenInitObjectType,
		deleted:         false,
	}
}

func newPTokenInitObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*PTokenInitObject, error) {
	var newPTokenInitState = NewPTokenInitState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPTokenInitState)
		if err != nil {
			return nil, err
		}
	} else {
		newPTokenInitState, ok = data.(*PTokenInitState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPTokenInitStateType, reflect.TypeOf(data))
		}
	}
	return &PTokenInitObject{
		version:         defaultVersion,
		pTokenInitHash:  key,
		pTokenInitState: newPTokenInitState,
		db:              db,
		objectType:      PTokenInitObjectType,
		deleted:         false,
	}, nil
}

func GeneratePTokenInitObjecKey(tokenID string) common.Hash {
	prefixHash := GetPTokenInitPrefix()
	valueHash := common.HashH([]byte(tokenID))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (t PTokenInitObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *PTokenInitObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t PTokenInitObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *PTokenInitObject) SetValue(data interface{}) error {
	var newPTokenInitState = NewPTokenInitState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPTokenInitState)
		if err != nil {
			return err
		}
	} else {
		newPTokenInitState, ok = data.(*PTokenInitState)
		if !ok {
			return fmt.Errorf("%+v, got type %+v", ErrInvalidPTokenInitStateType, reflect.TypeOf(data))
		}
	}
	t.pTokenInitState = newPTokenInitState
	return nil
}

func (t PTokenInitObject) GetValue() interface{} {
	return t.pTokenInitState
}

func (t PTokenInitObject) GetValueBytes() []byte {
	pTokenInitState, ok := t.GetValue().(*PTokenInitState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(pTokenInitState)
	if err != nil {
		panic("failed to marshal PTokenInitState")
	}
	return value
}

func (t PTokenInitObject) GetHash() common.Hash {
	return t.pTokenInitHash
}

func (t PTokenInitObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *PTokenInitObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *PTokenInitObject) Reset() bool {
	t.pTokenInitState = NewPTokenInitState()
	return true
}

func (t PTokenInitObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t PTokenInitObject) IsEmpty() bool {
	temp := NewPTokenInitState()
	return reflect.DeepEqual(temp, t.pTokenInitState) || t.pTokenInitState == nil
}
