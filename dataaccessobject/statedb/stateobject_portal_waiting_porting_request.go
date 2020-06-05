package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

type MatchingPortingCustodianDetail struct {
	IncAddress             string
	RemoteAddress          string
	Amount                 uint64
	LockedAmountCollateral uint64
}

type WaitingPortingRequest struct {
	uniquePortingID string
	tokenID         string
	porterAddress   string
	amount          uint64
	custodians      []*MatchingPortingCustodianDetail
	portingFee      uint64
	beaconHeight    uint64
	txReqID         common.Hash
}

func (w *WaitingPortingRequest) BeaconHeight() uint64 {
	return w.beaconHeight
}

func (w *WaitingPortingRequest) SetBeaconHeight(beaconHeight uint64) {
	w.beaconHeight = beaconHeight
}

func (w *WaitingPortingRequest) PortingFee() uint64 {
	return w.portingFee
}

func (w *WaitingPortingRequest) SetPortingFee(portingFee uint64) {
	w.portingFee = portingFee
}

func (w *WaitingPortingRequest) Custodians() []*MatchingPortingCustodianDetail {
	return w.custodians
}

func (w *WaitingPortingRequest) SetCustodians(custodians []*MatchingPortingCustodianDetail) {
	w.custodians = custodians
}

func (w *WaitingPortingRequest) Amount() uint64 {
	return w.amount
}

func (w *WaitingPortingRequest) SetAmount(amount uint64) {
	w.amount = amount
}

func (w *WaitingPortingRequest) PorterAddress() string {
	return w.porterAddress
}

func (w *WaitingPortingRequest) SetPorterAddress(porterAddress string) {
	w.porterAddress = porterAddress
}

func (w *WaitingPortingRequest) TokenID() string {
	return w.tokenID
}

func (w *WaitingPortingRequest) SetTokenID(tokenID string) {
	w.tokenID = tokenID
}

func (w *WaitingPortingRequest) TxReqID() common.Hash {
	return w.txReqID
}

func (w *WaitingPortingRequest) SetTxReqID(txReqID common.Hash) {
	w.txReqID = txReqID
}

func (w *WaitingPortingRequest) UniquePortingID() string {
	return w.uniquePortingID
}

func (w *WaitingPortingRequest) SetUniquePortingID(uniquePortingID string) {
	w.uniquePortingID = uniquePortingID
}

func NewWaitingPortingRequest() *WaitingPortingRequest {
	return &WaitingPortingRequest{}
}

func NewWaitingPortingRequestWithValue(
	uniquePortingID string,
	txReqID common.Hash,
	tokenID string,
	porterAddress string,
	amount uint64,
	custodians []*MatchingPortingCustodianDetail,
	portingFee uint64,
	beaconHeight uint64) *WaitingPortingRequest {
	return &WaitingPortingRequest{uniquePortingID: uniquePortingID, txReqID: txReqID, tokenID: tokenID, porterAddress: porterAddress, amount: amount, custodians: custodians, portingFee: portingFee, beaconHeight: beaconHeight}
}

func GeneratePortalWaitingPortingRequestObjectKey(portingRequestId string) common.Hash {
	prefixHash := GetPortalWaitingPortingRequestPrefix()
	valueHash := common.HashH([]byte(portingRequestId))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (w *WaitingPortingRequest) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		UniquePortingID string
		TxReqID         common.Hash
		TokenID         string
		PorterAddress   string
		Amount          uint64
		Custodians      []*MatchingPortingCustodianDetail
		PortingFee      uint64
		BeaconHeight    uint64
	}{
		UniquePortingID: w.uniquePortingID,
		TxReqID:         w.txReqID,
		TokenID:         w.tokenID,
		PorterAddress:   w.porterAddress,
		Amount:          w.amount,
		Custodians:      w.custodians,
		PortingFee:      w.portingFee,
		BeaconHeight:    w.beaconHeight,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (w *WaitingPortingRequest) UnmarshalJSON(data []byte) error {
	temp := struct {
		UniquePortingID string
		TxReqID         common.Hash
		TokenID         string
		PorterAddress   string
		Amount          uint64
		Custodians      []*MatchingPortingCustodianDetail
		PortingFee      uint64
		BeaconHeight    uint64
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}

	w.uniquePortingID = temp.UniquePortingID
	w.txReqID = temp.TxReqID
	w.tokenID = temp.TokenID
	w.porterAddress = temp.PorterAddress
	w.amount = temp.Amount
	w.custodians = temp.Custodians
	w.portingFee = temp.PortingFee
	w.beaconHeight = temp.BeaconHeight

	return nil
}

type WaitingPortingRequestObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version     int
	keyObject   common.Hash
	valueObject *WaitingPortingRequest
	objectType  int
	deleted     bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newWaitingPortingRequestObjectWithValue(db *StateDB, keyObject common.Hash, valueObject interface{}) (*WaitingPortingRequestObject, error) {
	var content = NewWaitingPortingRequest()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = valueObject.([]byte); ok {
		err := json.Unmarshal(dataBytes, content)
		if err != nil {
			return nil, err
		}
	} else {
		content, ok = valueObject.(*WaitingPortingRequest)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidWaitingPortingRequestType, reflect.TypeOf(valueObject))
		}
	}
	return &WaitingPortingRequestObject{
		db:          db,
		version:     defaultVersion,
		keyObject:   keyObject,
		valueObject: content,
		objectType:  PortalWaitingPortingRequestObjectType,
		deleted:     false,
	}, nil
}

func newWaitingPortingRequestObject(db *StateDB, keyObject common.Hash) *WaitingPortingRequestObject {
	return &WaitingPortingRequestObject{
		db:          db,
		version:     defaultVersion,
		keyObject:   keyObject,
		valueObject: NewWaitingPortingRequest(),
		objectType:  PortalWaitingPortingRequestObjectType,
		deleted:     false,
	}
}

func (l WaitingPortingRequestObject) GetVersion() int {
	return l.version
}

// setError remembers the first non-nil error it is called with.
func (l *WaitingPortingRequestObject) SetError(err error) {
	if l.dbErr == nil {
		l.dbErr = err
	}
}

func (l WaitingPortingRequestObject) GetTrie(db DatabaseAccessWarper) Trie {
	return l.trie
}

func (l *WaitingPortingRequestObject) SetValue(data interface{}) error {
	valueObject, ok := data.(*WaitingPortingRequest)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidWaitingPortingRequestType, reflect.TypeOf(data))
	}
	l.valueObject = valueObject
	return nil
}

func (l WaitingPortingRequestObject) GetValue() interface{} {
	return l.valueObject
}

func (l WaitingPortingRequestObject) GetValueBytes() []byte {
	valueObject, ok := l.GetValue().(*WaitingPortingRequest)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(valueObject)
	if err != nil {
		panic("failed to marshal WaitingPortingRequest")
	}
	return value
}

func (l WaitingPortingRequestObject) GetHash() common.Hash {
	return l.keyObject
}

func (l WaitingPortingRequestObject) GetType() int {
	return l.objectType
}

// MarkDelete will delete an object in trie
func (l *WaitingPortingRequestObject) MarkDelete() {
	l.deleted = true
}

// reset all shard committee value into default value
func (l *WaitingPortingRequestObject) Reset() bool {
	l.valueObject = NewWaitingPortingRequest()
	return true
}

func (l WaitingPortingRequestObject) IsDeleted() bool {
	return l.deleted
}

// value is either default or nil
func (l WaitingPortingRequestObject) IsEmpty() bool {
	temp := NewWaitingPortingRequest()
	return reflect.DeepEqual(temp, l.valueObject) || l.valueObject == nil
}
