package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

type TokenTransactionState struct {
	txHash common.Hash
}

func (t *TokenTransactionState) TxHash() common.Hash {
	return t.txHash
}

func (t *TokenTransactionState) SetTxHash(txHash common.Hash) {
	t.txHash = txHash
}

func (t TokenTransactionState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		TxHash common.Hash
	}{
		TxHash: t.txHash,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (t *TokenTransactionState) UnmarshalJSON(data []byte) error {
	temp := struct {
		TxHash common.Hash
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	t.txHash = temp.TxHash
	return nil
}

func NewTokenTransactionState() *TokenTransactionState {
	return &TokenTransactionState{}
}

func NewTokenTransactionStateWithValue(txHash common.Hash) *TokenTransactionState {
	return &TokenTransactionState{txHash: txHash}
}

type TokenTransactionObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version               int
	tokenHash             common.Hash
	TokenTransactionState *TokenTransactionState
	objectType            int
	deleted               bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newTokenTransactionObject(db *StateDB, hash common.Hash) *TokenTransactionObject {
	return &TokenTransactionObject{
		version:               defaultVersion,
		db:                    db,
		tokenHash:             hash,
		TokenTransactionState: NewTokenTransactionState(),
		objectType:            TokenTransactionObjectType,
		deleted:               false,
	}
}

func newTokenTransactionObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*TokenTransactionObject, error) {
	var newTokenTransactionState = NewTokenTransactionState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newTokenTransactionState)
		if err != nil {
			return nil, err
		}
	} else {
		newTokenTransactionState, ok = data.(*TokenTransactionState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidTokenTransactionStateType, reflect.TypeOf(data))
		}
	}
	return &TokenTransactionObject{
		version:               defaultVersion,
		tokenHash:             key,
		TokenTransactionState: newTokenTransactionState,
		db:                    db,
		objectType:            TokenTransactionObjectType,
		deleted:               false,
	}, nil
}

func GenerateTokenTransactionObjectKey(tokenID, txHash common.Hash) common.Hash {
	prefixHash := GetTokenTransactionPrefix(tokenID)
	valueHash := txHash[:]
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (t TokenTransactionObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *TokenTransactionObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t TokenTransactionObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *TokenTransactionObject) SetValue(data interface{}) error {
	newTokenTransactionState, ok := data.(*TokenTransactionState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidTokenTransactionStateType, reflect.TypeOf(data))
	}
	t.TokenTransactionState = newTokenTransactionState
	return nil
}

func (t TokenTransactionObject) GetValue() interface{} {
	return t.TokenTransactionState
}

func (t TokenTransactionObject) GetValueBytes() []byte {
	TokenTransactionState, ok := t.GetValue().(*TokenTransactionState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(TokenTransactionState)
	if err != nil {
		panic("failed to marshal token state")
	}
	return value
}

func (t TokenTransactionObject) GetHash() common.Hash {
	return t.tokenHash
}

func (t TokenTransactionObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *TokenTransactionObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *TokenTransactionObject) Reset() bool {
	t.TokenTransactionState = NewTokenTransactionState()
	return true
}

func (t TokenTransactionObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t TokenTransactionObject) IsEmpty() bool {
	temp := NewTokenTransactionState()
	return reflect.DeepEqual(temp, t.TokenTransactionState) || t.TokenTransactionState == nil
}
