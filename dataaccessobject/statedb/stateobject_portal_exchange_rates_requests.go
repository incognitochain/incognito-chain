package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

type ExchangeRateInfo struct {
	PTokenID string
	Rate     uint64
}

type ExchangeRatesRequest struct {
	senderAddress string
	rates         []*ExchangeRateInfo
}

func NewExchangeRatesRequestWithValue(senderAddress string, rates []*ExchangeRateInfo) *ExchangeRatesRequest {
	return &ExchangeRatesRequest{senderAddress: senderAddress, rates: rates}
}

func NewExchangeRatesRequest() *ExchangeRatesRequest {
	return &ExchangeRatesRequest{}
}


func (e *ExchangeRatesRequest) Rates() []*ExchangeRateInfo {
	return e.rates
}

func (e *ExchangeRatesRequest) SetRates(rates []*ExchangeRateInfo) {
	e.rates = rates
}

func (e *ExchangeRatesRequest) SenderAddress() string {
	return e.senderAddress
}

func (e *ExchangeRatesRequest) SetSenderAddress(senderAddress string) {
	e.senderAddress = senderAddress
}

func (e *ExchangeRatesRequest) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		SenderAddress string
		Rates         []*ExchangeRateInfo
	}{
		SenderAddress: e.senderAddress,
		Rates:         e.rates,
	})
	if err != nil {
		return []byte{}, err
	}

	return data, nil
}

func (e *ExchangeRatesRequest) UnmarshalJSON(data[]byte) error {
	temp := struct {
		SenderAddress string
		Rates         []*ExchangeRateInfo
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	e.senderAddress= temp.SenderAddress
	e.rates = temp.Rates
	return nil
}

type ExchangeRatesRequestObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version                     int
	exchangeRatesRequestHash  common.Hash
	exchangeRatesRequest *ExchangeRatesRequest
	objectType                  int
	deleted                     bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func (e ExchangeRatesRequestObject) GetVersion() int {
	return e.version
}

// setError remembers the first non-nil error it is called with.
func (e *ExchangeRatesRequestObject) SetError(err error) {
	if e.dbErr == nil {
		e.dbErr = err
	}
}

func (e ExchangeRatesRequestObject) GetTrie(db DatabaseAccessWarper) Trie {
	return e.trie
}

func (e *ExchangeRatesRequestObject) SetValue(data interface{}) error {
	exchangeRatesRequests, ok := data.(*ExchangeRatesRequest)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidExchangeRatesRequestType, reflect.TypeOf(data))
	}
	e.exchangeRatesRequest = exchangeRatesRequests
	return nil
}

func (e ExchangeRatesRequestObject) GetValue() interface{} {
	return e.exchangeRatesRequest
}

func (e ExchangeRatesRequestObject) GetValueBytes() []byte {
	exchangeRatesRequests, ok := e.GetValue().(*ExchangeRatesRequest)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(exchangeRatesRequests)
	if err != nil {
		panic("failed to marshal ExchangeRatesRequest")
	}
	return value
}

func (e ExchangeRatesRequestObject) GetHash() common.Hash {
	return e.exchangeRatesRequestHash
}

func (e ExchangeRatesRequestObject) GetType() int {
	return e.objectType
}

// MarkDelete will delete an object in trie
func (e *ExchangeRatesRequestObject) MarkDelete() {
	e.deleted = true
}

// reset all shard committee value into default value
func (e *ExchangeRatesRequestObject) Reset() bool {
	e.exchangeRatesRequest = NewExchangeRatesRequest()
	return true
}

func (e ExchangeRatesRequestObject) IsDeleted() bool {
	return e.deleted
}

// value is either default or nil
func (e ExchangeRatesRequestObject) IsEmpty() bool {
	temp := NewExchangeRatesRequest()
	return reflect.DeepEqual(temp, e.exchangeRatesRequest) || e.exchangeRatesRequest == nil
}

func NewExchangeRatesRequestObjectWithValue(db *StateDB, exchangeRatesRequestsHash common.Hash, data interface{}) (*ExchangeRatesRequestObject, error) {
	var exchangeRatesRequests = NewExchangeRatesRequest()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, exchangeRatesRequests)
		if err != nil {
			return nil, err
		}
	} else {
		exchangeRatesRequests, ok = data.(*ExchangeRatesRequest)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidExchangeRatesRequestType, reflect.TypeOf(data))
		}
	}
	return &ExchangeRatesRequestObject{
		db:                         db,
		version:                    defaultVersion,
		exchangeRatesRequestHash: 	exchangeRatesRequestsHash,
		exchangeRatesRequest:     	exchangeRatesRequests,
		objectType:                 ExchangeRatesStateObjectType,
		deleted:                    false,
	}, nil
}

func NewExchangeRatesRequestObject(db *StateDB, exchangeRatesRequestsHash common.Hash) *ExchangeRatesRequestObject {
	return &ExchangeRatesRequestObject{
		db:                         db,
		version:                    defaultVersion,
		exchangeRatesRequestHash: 	exchangeRatesRequestsHash,
		exchangeRatesRequest:     	NewExchangeRatesRequest(),
		objectType:                 ExchangeRatesStateObjectType,
		deleted:                    false,
	}
}

func GenerateExchangeRatesRequestObjectKey(txId string) common.Hash {
	prefixHash := GetExchangeRatesRequestPrefix()
	valueHash := common.HashH([]byte(txId))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}


