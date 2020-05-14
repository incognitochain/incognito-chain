package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

type LiquidateExchangeRatesDetail struct {
	HoldAmountFreeCollateral uint64
	HoldAmountPubToken       uint64
}

type LiquidateExchangeRatesPool struct {
	rates map[string]LiquidateExchangeRatesDetail //ptoken | detail
}

func (l *LiquidateExchangeRatesPool) Rates() map[string]LiquidateExchangeRatesDetail {
	return l.rates
}

func (l *LiquidateExchangeRatesPool) SetRates(rates map[string]LiquidateExchangeRatesDetail) {
	l.rates = rates
}

func NewLiquidateExchangeRatesPool() *LiquidateExchangeRatesPool {
	return &LiquidateExchangeRatesPool{}
}

func NewLiquidateExchangeRatesPoolWithValue(rates map[string]LiquidateExchangeRatesDetail) *LiquidateExchangeRatesPool {
	return &LiquidateExchangeRatesPool{rates: rates}
}

func GeneratePortalLiquidateExchangeRatesPoolObjectKey() common.Hash {
	suffix := "liquidation"
	prefixHash := GetPortalLiquidationExchangeRatesPoolPrefix()
	valueHash := common.HashH([]byte(suffix))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (l *LiquidateExchangeRatesPool) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Rates map[string]LiquidateExchangeRatesDetail
	}{
		Rates: l.rates,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (l *LiquidateExchangeRatesPool) UnmarshalJSON(data []byte) error {
	temp := struct {
		Rates map[string]LiquidateExchangeRatesDetail
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	l.rates = temp.Rates
	return nil
}

type LiquidateExchangeRatesPoolObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version     int
	keyObject   common.Hash
	valueObject *LiquidateExchangeRatesPool
	objectType  int
	deleted     bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newLiquidateExchangeRatesPoolObjectWithValue(db *StateDB, keyObject common.Hash, valueObject interface{}) (*LiquidateExchangeRatesPoolObject, error) {
	var content = NewLiquidateExchangeRatesPool()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = valueObject.([]byte); ok {
		err := json.Unmarshal(dataBytes, content)
		if err != nil {
			return nil, err
		}
	} else {
		content, ok = valueObject.(*LiquidateExchangeRatesPool)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidLiquidationExchangeRatesType, reflect.TypeOf(valueObject))
		}
	}
	return &LiquidateExchangeRatesPoolObject{
		db:          db,
		version:     defaultVersion,
		keyObject:   keyObject,
		valueObject: content,
		objectType:  PortalLiquidationExchangeRatesPoolObjectType,
		deleted:     false,
	}, nil
}

func newLiquidateExchangeRatesPoolObject(db *StateDB, keyObject common.Hash) *LiquidateExchangeRatesPoolObject {
	return &LiquidateExchangeRatesPoolObject{
		db:          db,
		version:     defaultVersion,
		keyObject:   keyObject,
		valueObject: NewLiquidateExchangeRatesPool(),
		objectType:  PortalLiquidationExchangeRatesPoolObjectType,
		deleted:     false,
	}
}

func (l LiquidateExchangeRatesPoolObject) GetVersion() int {
	return l.version
}

// setError remembers the first non-nil error it is called with.
func (l *LiquidateExchangeRatesPoolObject) SetError(err error) {
	if l.dbErr == nil {
		l.dbErr = err
	}
}

func (l LiquidateExchangeRatesPoolObject) GetTrie(db DatabaseAccessWarper) Trie {
	return l.trie
}

func (l *LiquidateExchangeRatesPoolObject) SetValue(data interface{}) error {
	valueObject, ok := data.(*LiquidateExchangeRatesPool)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidLiquidationExchangeRatesType, reflect.TypeOf(data))
	}
	l.valueObject = valueObject
	return nil
}

func (l LiquidateExchangeRatesPoolObject) GetValue() interface{} {
	return l.valueObject
}

func (l LiquidateExchangeRatesPoolObject) GetValueBytes() []byte {
	valueObject, ok := l.GetValue().(*LiquidateExchangeRatesPool)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(valueObject)
	if err != nil {
		panic("failed to marshal LiquidateExchangeRatesPool")
	}
	return value
}

func (l LiquidateExchangeRatesPoolObject) GetHash() common.Hash {
	return l.keyObject
}

func (l LiquidateExchangeRatesPoolObject) GetType() int {
	return l.objectType
}

// MarkDelete will delete an object in trie
func (l *LiquidateExchangeRatesPoolObject) MarkDelete() {
	l.deleted = true
}

// reset all shard committee value into default value
func (l *LiquidateExchangeRatesPoolObject) Reset() bool {
	l.valueObject = NewLiquidateExchangeRatesPool()
	return true
}

func (l LiquidateExchangeRatesPoolObject) IsDeleted() bool {
	return l.deleted
}

// value is either default or nil
func (l LiquidateExchangeRatesPoolObject) IsEmpty() bool {
	temp := NewLiquidateExchangeRatesPool()
	return reflect.DeepEqual(temp, l.valueObject) || l.valueObject == nil
}
