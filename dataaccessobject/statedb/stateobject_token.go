package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

type TokenState struct {
	tokenID        common.Hash
	propertyName   string
	propertySymbol string
	tokenType      int    // action type
	mintable       bool   // default false
	amount         uint64 // init amount
	info           []byte
	initTx         common.Hash
	txs            []common.Hash
}

func (t *TokenState) Info() []byte {
	return t.info
}

func (t *TokenState) SetInfo(info []byte) {
	t.info = info
}

func (t TokenState) TokenID() common.Hash {
	return t.tokenID
}

func (t *TokenState) SetTokenID(tokenID common.Hash) {
	t.tokenID = tokenID
}

func (t TokenState) PropertyName() string {
	return t.propertyName
}

func (t *TokenState) SetPropertyName(propertyName string) {
	t.propertyName = propertyName
}

func (t TokenState) PropertySymbol() string {
	return t.propertySymbol
}

func (t *TokenState) SetPropertySymbol(propertySymbol string) {
	t.propertySymbol = propertySymbol
}

func (t TokenState) TokenType() int {
	return t.tokenType
}

func (t *TokenState) SetTokenType(tokenType int) {
	t.tokenType = tokenType
}

func (t TokenState) Mintable() bool {
	return t.mintable
}

func (t *TokenState) SetMintable(mintable bool) {
	t.mintable = mintable
}

func (t TokenState) Amount() uint64 {
	return t.amount
}

func (t *TokenState) SetAmount(amount uint64) {
	t.amount = amount
}

func (t TokenState) InitTx() common.Hash {
	return t.initTx
}

func (t *TokenState) SetInitTx(initTx common.Hash) {
	t.initTx = initTx
}

func (t *TokenState) AddTxs(txs []common.Hash) {
	t.txs = append(t.txs, txs...)
}

func (t TokenState) Txs() []common.Hash {
	return t.txs
}

func (t *TokenState) SetTxs(txs []common.Hash) {
	t.txs = txs
}

func (t TokenState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		TokenID        common.Hash
		PropertyName   string
		PropertySymbol string
		TokenType      int
		Mintable       bool
		Amount         uint64
		Info           []byte
		InitTx         common.Hash
		Txs            []common.Hash
	}{
		TokenID:        t.tokenID,
		PropertyName:   t.propertyName,
		PropertySymbol: t.propertySymbol,
		TokenType:      t.tokenType,
		Mintable:       t.mintable,
		Amount:         t.amount,
		Info:           t.info,
		InitTx:         t.initTx,
		Txs:            t.txs,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (t *TokenState) UnmarshalJSON(data []byte) error {
	temp := struct {
		TokenID        common.Hash
		PropertyName   string
		PropertySymbol string
		TokenType      int
		Mintable       bool
		Amount         uint64
		Info           []byte
		InitTx         common.Hash
		Txs            []common.Hash
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	t.tokenID = temp.TokenID
	t.propertyName = temp.PropertyName
	t.propertySymbol = temp.PropertySymbol
	t.tokenType = temp.TokenType
	t.mintable = temp.Mintable
	t.amount = temp.Amount
	t.info = temp.Info
	t.initTx = temp.InitTx
	t.txs = temp.Txs
	return nil
}

func NewTokenState() *TokenState {
	return &TokenState{}
}

func NewTokenStateForInitToken(tokenID common.Hash, initTx common.Hash) *TokenState {
	return &TokenState{tokenID: tokenID, initTx: initTx}
}

func NewTokenStateWithValue(tokenID common.Hash, propertyName string, propertySymbol string, tokenType int, mintable bool, amount uint64, info []byte, initTx common.Hash, txs []common.Hash) *TokenState {
	return &TokenState{tokenID: tokenID, propertyName: propertyName, propertySymbol: propertySymbol, tokenType: tokenType, mintable: mintable, amount: amount, info: info, initTx: initTx, txs: txs}
}

type TokenObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	tokenHash  common.Hash
	tokenState *TokenState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newTokenObject(db *StateDB, hash common.Hash) *TokenObject {
	return &TokenObject{
		version:    defaultVersion,
		db:         db,
		tokenHash:  hash,
		tokenState: NewTokenState(),
		objectType: TokenObjectType,
		deleted:    false,
	}
}
func newTokenObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*TokenObject, error) {
	var newTokenState = NewTokenState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newTokenState)
		if err != nil {
			return nil, err
		}
	} else {
		newTokenState, ok = data.(*TokenState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidTokenStateType, reflect.TypeOf(data))
		}
	}
	return &TokenObject{
		version:    defaultVersion,
		tokenHash:  key,
		tokenState: newTokenState,
		db:         db,
		objectType: TokenObjectType,
		deleted:    false,
	}, nil
}

func GenerateTokenObjectKey(tokenID common.Hash) common.Hash {
	prefixHash := GetTokenPrefix()
	valueHash := common.HashH(tokenID[:])
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (t TokenObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *TokenObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t TokenObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *TokenObject) SetValue(data interface{}) error {
	newTokenState, ok := data.(*TokenState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidTokenStateType, reflect.TypeOf(data))
	}
	t.tokenState = newTokenState
	return nil
}

func (t TokenObject) GetValue() interface{} {
	return t.tokenState
}

func (t TokenObject) GetValueBytes() []byte {
	tokenState, ok := t.GetValue().(*TokenState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(tokenState)
	if err != nil {
		panic("failed to marshal token state")
	}
	return value
}

func (t TokenObject) GetHash() common.Hash {
	return t.tokenHash
}

func (t TokenObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *TokenObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *TokenObject) Reset() bool {
	t.tokenState = NewTokenState()
	return true
}

func (t TokenObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t TokenObject) IsEmpty() bool {
	temp := NewTokenState()
	return reflect.DeepEqual(temp, t.tokenState) || t.tokenState == nil
}
