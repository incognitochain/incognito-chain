package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

type WaitingRedeemRequest struct {
	uniqueRedeemID        string
	tokenID               string
	redeemerAddress       string
	redeemerRemoteAddress string
	redeemAmount          uint64
	custodians            []*MatchingRedeemCustodianDetail
	redeemFee             uint64
	beaconHeight          uint64
	txReqID               common.Hash
}

type MatchingRedeemCustodianDetail struct {
	IncAddress    string
	RemoteAddress string
	Amount        uint64
}

func (rq WaitingRedeemRequest) GetUniqueRedeemID() string {
	return rq.uniqueRedeemID
}

func (rq *WaitingRedeemRequest) SetUniqueRedeemID(uniqueRedeemID string) {
	rq.uniqueRedeemID = uniqueRedeemID
}

func (rq WaitingRedeemRequest) GetTokenID() string {
	return rq.tokenID
}

func (rq *WaitingRedeemRequest) SetTokenID(tokenID string) {
	rq.tokenID = tokenID
}

func (rq WaitingRedeemRequest) GetRedeemerAddress() string {
	return rq.redeemerAddress
}

func (rq *WaitingRedeemRequest) SetRedeemerAddress(redeemerAddress string) {
	rq.redeemerAddress = redeemerAddress
}

func (rq WaitingRedeemRequest) GetRedeemerRemoteAddress() string {
	return rq.redeemerRemoteAddress
}

func (rq *WaitingRedeemRequest) SetRedeemerRemoteAddress(redeemerRemoteAddress string) {
	rq.redeemerRemoteAddress = redeemerRemoteAddress
}

func (rq WaitingRedeemRequest) GetRedeemAmount() uint64 {
	return rq.redeemAmount
}

func (rq *WaitingRedeemRequest) SetRedeemAmount(redeemAmount uint64) {
	rq.redeemAmount = redeemAmount
}

func (rq WaitingRedeemRequest) GetCustodians() []*MatchingRedeemCustodianDetail {
	return rq.custodians
}

func (rq *WaitingRedeemRequest) SetCustodians(custodians []*MatchingRedeemCustodianDetail) {
	rq.custodians = custodians
}

func (rq WaitingRedeemRequest) GetRedeemFee() uint64 {
	return rq.redeemFee
}

func (rq *WaitingRedeemRequest) SetRedeemFee(redeemFee uint64) {
	rq.redeemFee = redeemFee
}

func (rq WaitingRedeemRequest) GetBeaconHeight() uint64 {
	return rq.beaconHeight
}

func (rq *WaitingRedeemRequest) SetBeaconHeight(beaconHeight uint64) {
	rq.beaconHeight = beaconHeight
}

func (rq WaitingRedeemRequest) GetTxReqID() common.Hash {
	return rq.txReqID
}

func (rq *WaitingRedeemRequest) SetTxReqID(txReqID common.Hash) {
	rq.txReqID = txReqID
}

func (rq WaitingRedeemRequest) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		UniqueRedeemID        string
		TokenID               string
		RedeemerAddress       string
		RedeemerRemoteAddress string
		RedeemAmount          uint64
		Custodians            []*MatchingRedeemCustodianDetail
		RedeemFee             uint64
		BeaconHeight          uint64
		TxReqID               common.Hash
	}{
		UniqueRedeemID:        rq.uniqueRedeemID,
		TokenID:               rq.tokenID,
		RedeemerAddress:       rq.redeemerAddress,
		RedeemerRemoteAddress: rq.redeemerRemoteAddress,
		RedeemAmount:          rq.redeemAmount,
		Custodians:            rq.custodians,
		RedeemFee:             rq.redeemFee,
		BeaconHeight:          rq.beaconHeight,
		TxReqID:               rq.txReqID,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (rq *WaitingRedeemRequest) UnmarshalJSON(data []byte) error {
	temp := struct {
		UniqueRedeemID        string
		TokenID               string
		RedeemerAddress       string
		RedeemerRemoteAddress string
		RedeemAmount          uint64
		Custodians            []*MatchingRedeemCustodianDetail
		RedeemFee             uint64
		BeaconHeight          uint64
		TxReqID               common.Hash
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	rq.uniqueRedeemID = temp.UniqueRedeemID
	rq.tokenID = temp.TokenID
	rq.redeemerAddress = temp.RedeemerAddress
	rq.redeemerRemoteAddress = temp.RedeemerRemoteAddress
	rq.redeemAmount = temp.RedeemAmount
	rq.custodians = temp.Custodians
	rq.redeemFee = temp.RedeemFee
	rq.beaconHeight = temp.BeaconHeight
	rq.txReqID = temp.TxReqID
	return nil
}

func NewWaitingRedeemRequest() *WaitingRedeemRequest {
	return &WaitingRedeemRequest{}
}

func NewWaitingRedeemRequestWithValue(
	uniqueRedeemID string,
	tokenID string,
	redeemerAddress string,
	redeemerRemoteAddress string,
	redeemAmount uint64,
	custodians []*MatchingRedeemCustodianDetail,
	redeemFee uint64,
	beaconHeight uint64,
	txReqID common.Hash) *WaitingRedeemRequest {

	return &WaitingRedeemRequest{
		uniqueRedeemID:        uniqueRedeemID,
		tokenID:               tokenID,
		redeemerAddress:       redeemerAddress,
		redeemerRemoteAddress: redeemerRemoteAddress,
		redeemAmount:          redeemAmount,
		custodians:            custodians,
		redeemFee:             redeemFee,
		beaconHeight:          beaconHeight,
		txReqID:               txReqID,
	}
}

type WaitingRedeemRequestObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version                  int
	waitingRedeemRequestHash common.Hash
	waitingRedeemRequest     *WaitingRedeemRequest
	objectType               int
	deleted                  bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newWaitingRedeemRequestObject(db *StateDB, hash common.Hash) *WaitingRedeemRequestObject {
	return &WaitingRedeemRequestObject{
		version:                  defaultVersion,
		db:                       db,
		waitingRedeemRequestHash: hash,
		waitingRedeemRequest:     NewWaitingRedeemRequest(),
		objectType:               RedeemRequestObjectType,
		deleted:                  false,
	}
}

func newWaitingRedeemRequestObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*WaitingRedeemRequestObject, error) {
	var redeemRequest = NewWaitingRedeemRequest()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, redeemRequest)
		if err != nil {
			return nil, err
		}
	} else {
		redeemRequest, ok = data.(*WaitingRedeemRequest)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPortalWaitingRedeemRequestType, reflect.TypeOf(data))
		}
	}
	return &WaitingRedeemRequestObject{
		version:                  defaultVersion,
		waitingRedeemRequestHash: key,
		waitingRedeemRequest:     redeemRequest,
		db:                       db,
		objectType:               CustodianStateObjectType,
		deleted:                  false,
	}, nil
}

func GenerateWaitingRedeemRequestObjectKey(beaconHeight uint64, redeemID string) common.Hash {
	prefixHash := GetWaitingRedeemRequestPrefix()
	valueHash := common.HashH(append([]byte(fmt.Sprintf("%d-", beaconHeight)), []byte(redeemID)...))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (t WaitingRedeemRequestObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *WaitingRedeemRequestObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t WaitingRedeemRequestObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *WaitingRedeemRequestObject) SetValue(data interface{}) error {
	redeemRequest, ok := data.(*WaitingRedeemRequest)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPortalWaitingRedeemRequestType, reflect.TypeOf(data))
	}
	t.waitingRedeemRequest = redeemRequest
	return nil
}

func (t WaitingRedeemRequestObject) GetValue() interface{} {
	return t.waitingRedeemRequest
}

func (t WaitingRedeemRequestObject) GetValueBytes() []byte {
	redeemRequest, ok := t.GetValue().(*WaitingRedeemRequest)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(redeemRequest)
	if err != nil {
		panic("failed to marshal redeem request")
	}
	return value
}

func (t WaitingRedeemRequestObject) GetHash() common.Hash {
	return t.waitingRedeemRequestHash
}

func (t WaitingRedeemRequestObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *WaitingRedeemRequestObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *WaitingRedeemRequestObject) Reset() bool {
	t.waitingRedeemRequest = NewWaitingRedeemRequest()
	return true
}

func (t WaitingRedeemRequestObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t WaitingRedeemRequestObject) IsEmpty() bool {
	temp := NewCustodianState()
	return reflect.DeepEqual(temp, t.waitingRedeemRequest) || t.waitingRedeemRequest == nil
}
