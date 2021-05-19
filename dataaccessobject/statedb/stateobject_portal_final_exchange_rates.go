package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

type FinalExchangeRatesDetail struct {
	Amount uint64
}

type FinalExchangeRatesState struct {
	rates map[string]FinalExchangeRatesDetail
}

func (f *FinalExchangeRatesState) Rates() map[string]FinalExchangeRatesDetail {
	return f.rates
}

func (f *FinalExchangeRatesState) SetRates(rates map[string]FinalExchangeRatesDetail) {
	f.rates = rates
}

func NewFinalExchangeRatesState() *FinalExchangeRatesState {
	return &FinalExchangeRatesState{}
}

func NewFinalExchangeRatesStateWithValue(rates map[string]FinalExchangeRatesDetail) *FinalExchangeRatesState {
	return &FinalExchangeRatesState{rates: rates}
}

func GeneratePortalFinalExchangeRatesStateObjectKey() common.Hash {
	suffix := "exchangerates"
	prefixHash := GetFinalExchangeRatesStatePrefix()
	valueHash := common.HashH([]byte(suffix))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (f *FinalExchangeRatesState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Rates map[string]FinalExchangeRatesDetail
	}{
		Rates: f.rates,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (f *FinalExchangeRatesState) UnmarshalJSON(data []byte) error {
	temp := struct {
		Rates map[string]FinalExchangeRatesDetail
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	f.rates = temp.Rates
	return nil
}

type FinalExchangeRatesStateObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version                     int
	finalExchangeRatesStateHash common.Hash
	finalExchangeRatesState     *FinalExchangeRatesState
	objectType                  int
	deleted                     bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newFinalExchangeRatesStateObjectWithValue(db *StateDB, finalExchangeRatesStateHash common.Hash, data interface{}) (*FinalExchangeRatesStateObject, error) {
	var newFinalExchangeRatesState = NewFinalExchangeRatesState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newFinalExchangeRatesState)
		if err != nil {
			return nil, err
		}
	} else {
		newFinalExchangeRatesState, ok = data.(*FinalExchangeRatesState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidFinalExchangeRatesStateType, reflect.TypeOf(data))
		}
	}
	return &FinalExchangeRatesStateObject{
		db:                          db,
		version:                     defaultVersion,
		finalExchangeRatesStateHash: finalExchangeRatesStateHash,
		finalExchangeRatesState:     newFinalExchangeRatesState,
		objectType:                  PortalFinalExchangeRatesStateObjectType,
		deleted:                     false,
	}, nil
}

func newFinalExchangeRatesStateObject(db *StateDB, finalExchangeRatesStateHash common.Hash) *FinalExchangeRatesStateObject {
	return &FinalExchangeRatesStateObject{
		db:                          db,
		version:                     defaultVersion,
		finalExchangeRatesStateHash: finalExchangeRatesStateHash,
		finalExchangeRatesState:     NewFinalExchangeRatesState(),
		objectType:                  PortalFinalExchangeRatesStateObjectType,
		deleted:                     false,
	}
}

func (f FinalExchangeRatesStateObject) GetVersion() int {
	return f.version
}

// setError remembers the first non-nil error it is called with.
func (f *FinalExchangeRatesStateObject) SetError(err error) {
	if f.dbErr == nil {
		f.dbErr = err
	}
}

func (f FinalExchangeRatesStateObject) GetTrie(db DatabaseAccessWarper) Trie {
	return f.trie
}

func (f *FinalExchangeRatesStateObject) SetValue(data interface{}) error {
	finalExchangeRatesState, ok := data.(*FinalExchangeRatesState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidFinalExchangeRatesStateType, reflect.TypeOf(data))
	}
	f.finalExchangeRatesState = finalExchangeRatesState
	return nil
}

func (f FinalExchangeRatesStateObject) GetValue() interface{} {
	return f.finalExchangeRatesState
}

func (f FinalExchangeRatesStateObject) GetValueBytes() []byte {
	finalExchangeRatesState, ok := f.GetValue().(*FinalExchangeRatesState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(finalExchangeRatesState)
	if err != nil {
		panic("failed to marshal FinalExchangeRatesState")
	}
	return value
}

func (f FinalExchangeRatesStateObject) GetHash() common.Hash {
	return f.finalExchangeRatesStateHash
}

func (f FinalExchangeRatesStateObject) GetType() int {
	return f.objectType
}

// MarkDelete will delete an object in trie
func (f *FinalExchangeRatesStateObject) MarkDelete() {
	f.deleted = true
}

// reset all shard committee value into default value
func (f *FinalExchangeRatesStateObject) Reset() bool {
	f.finalExchangeRatesState = NewFinalExchangeRatesState()
	return true
}

func (f FinalExchangeRatesStateObject) IsDeleted() bool {
	return f.deleted
}

// value is either default or nil
func (f FinalExchangeRatesStateObject) IsEmpty() bool {
	temp := NewFinalExchangeRatesState()
	return reflect.DeepEqual(temp, f.finalExchangeRatesState) || f.finalExchangeRatesState == nil
}
