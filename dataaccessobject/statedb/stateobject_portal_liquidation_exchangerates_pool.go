package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

type LiquidationPoolDetail struct {
	CollateralAmount uint64
	PubTokenAmount   uint64
}

type LiquidationPool struct {
	rates map[string]LiquidationPoolDetail //ptoken | detail
}

func (l *LiquidationPool) Rates() map[string]LiquidationPoolDetail {
	return l.rates
}

func (l *LiquidationPool) SetRates(rates map[string]LiquidationPoolDetail) {
	l.rates = rates
}

func NewLiquidationPool() *LiquidationPool {
	return &LiquidationPool{}
}

func NewLiquidationPoolWithValue(rates map[string]LiquidationPoolDetail) *LiquidationPool {
	return &LiquidationPool{rates: rates}
}

func GeneratePortalLiquidationPoolObjectKey() common.Hash {
	suffix := "liquidation"
	prefixHash := GetPortalLiquidationPoolPrefix()
	valueHash := common.HashH([]byte(suffix))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (l *LiquidationPool) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Rates map[string]LiquidationPoolDetail
	}{
		Rates: l.rates,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (l *LiquidationPool) UnmarshalJSON(data []byte) error {
	temp := struct {
		Rates map[string]LiquidationPoolDetail
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	l.rates = temp.Rates
	return nil
}

type LiquidationPoolObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version     int
	keyObject   common.Hash
	valueObject *LiquidationPool
	objectType  int
	deleted     bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newLiquidationPoolObjectWithValue(db *StateDB, keyObject common.Hash, valueObject interface{}) (*LiquidationPoolObject, error) {
	var content = NewLiquidationPool()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = valueObject.([]byte); ok {
		err := json.Unmarshal(dataBytes, content)
		if err != nil {
			return nil, err
		}
	} else {
		content, ok = valueObject.(*LiquidationPool)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidLiquidationExchangeRatesType, reflect.TypeOf(valueObject))
		}
	}
	return &LiquidationPoolObject{
		db:          db,
		version:     defaultVersion,
		keyObject:   keyObject,
		valueObject: content,
		objectType:  PortalLiquidationPoolObjectType,
		deleted:     false,
	}, nil
}

func newLiquidationPoolObject(db *StateDB, keyObject common.Hash) *LiquidationPoolObject {
	return &LiquidationPoolObject{
		db:          db,
		version:     defaultVersion,
		keyObject:   keyObject,
		valueObject: NewLiquidationPool(),
		objectType:  PortalLiquidationPoolObjectType,
		deleted:     false,
	}
}

func (l LiquidationPoolObject) GetVersion() int {
	return l.version
}

// setError remembers the first non-nil error it is called with.
func (l *LiquidationPoolObject) SetError(err error) {
	if l.dbErr == nil {
		l.dbErr = err
	}
}

func (l LiquidationPoolObject) GetTrie(db DatabaseAccessWarper) Trie {
	return l.trie
}

func (l *LiquidationPoolObject) SetValue(data interface{}) error {
	valueObject, ok := data.(*LiquidationPool)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidLiquidationExchangeRatesType, reflect.TypeOf(data))
	}
	l.valueObject = valueObject
	return nil
}

func (l LiquidationPoolObject) GetValue() interface{} {
	return l.valueObject
}

func (l LiquidationPoolObject) GetValueBytes() []byte {
	valueObject, ok := l.GetValue().(*LiquidationPool)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(valueObject)
	if err != nil {
		panic("failed to marshal LiquidationPool")
	}
	return value
}

func (l LiquidationPoolObject) GetHash() common.Hash {
	return l.keyObject
}

func (l LiquidationPoolObject) GetType() int {
	return l.objectType
}

// MarkDelete will delete an object in trie
func (l *LiquidationPoolObject) MarkDelete() {
	l.deleted = true
}

// reset all shard committee value into default value
func (l *LiquidationPoolObject) Reset() bool {
	l.valueObject = NewLiquidationPool()
	return true
}

func (l LiquidationPoolObject) IsDeleted() bool {
	return l.deleted
}

// value is either default or nil
func (l LiquidationPoolObject) IsEmpty() bool {
	temp := NewLiquidationPool()
	return reflect.DeepEqual(temp, l.valueObject) || l.valueObject == nil
}
