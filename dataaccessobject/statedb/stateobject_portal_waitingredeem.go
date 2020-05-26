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
	incAddress    string
	remoteAddress string
	amount        uint64
}

func (wrq WaitingRedeemRequest) GetUniqueRedeemID() string {
	return wrq.uniqueRedeemID
}

func (wrq *WaitingRedeemRequest) SetUniqueRedeemID(uniqueRedeemID string) {
	wrq.uniqueRedeemID = uniqueRedeemID
}

func (wrq WaitingRedeemRequest) GetTokenID() string {
	return wrq.tokenID
}

func (wrq *WaitingRedeemRequest) SetTokenID(tokenID string) {
	wrq.tokenID = tokenID
}

func (wrq WaitingRedeemRequest) GetRedeemerAddress() string {
	return wrq.redeemerAddress
}

func (wrq *WaitingRedeemRequest) SetRedeemerAddress(redeemerAddress string) {
	wrq.redeemerAddress = redeemerAddress
}

func (wrq WaitingRedeemRequest) GetRedeemerRemoteAddress() string {
	return wrq.redeemerRemoteAddress
}

func (wrq *WaitingRedeemRequest) SetRedeemerRemoteAddress(redeemerRemoteAddress string) {
	wrq.redeemerRemoteAddress = redeemerRemoteAddress
}

func (wrq WaitingRedeemRequest) GetRedeemAmount() uint64 {
	return wrq.redeemAmount
}

func (wrq *WaitingRedeemRequest) SetRedeemAmount(redeemAmount uint64) {
	wrq.redeemAmount = redeemAmount
}

func (wrq WaitingRedeemRequest) GetCustodians() []*MatchingRedeemCustodianDetail {
	return wrq.custodians
}

func (wrq *WaitingRedeemRequest) SetCustodians(custodians []*MatchingRedeemCustodianDetail) {
	wrq.custodians = custodians
}

func (wrq WaitingRedeemRequest) GetRedeemFee() uint64 {
	return wrq.redeemFee
}

func (wrq *WaitingRedeemRequest) SetRedeemFee(redeemFee uint64) {
	wrq.redeemFee = redeemFee
}

func (wrq WaitingRedeemRequest) GetBeaconHeight() uint64 {
	return wrq.beaconHeight
}

func (wrq *WaitingRedeemRequest) SetBeaconHeight(beaconHeight uint64) {
	wrq.beaconHeight = beaconHeight
}

func (wrq WaitingRedeemRequest) GetTxReqID() common.Hash {
	return wrq.txReqID
}

func (wrq *WaitingRedeemRequest) SetTxReqID(txReqID common.Hash) {
	wrq.txReqID = txReqID
}

func (wrq WaitingRedeemRequest) MarshalJSON() ([]byte, error) {
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
		UniqueRedeemID:        wrq.uniqueRedeemID,
		TokenID:               wrq.tokenID,
		RedeemerAddress:       wrq.redeemerAddress,
		RedeemerRemoteAddress: wrq.redeemerRemoteAddress,
		RedeemAmount:          wrq.redeemAmount,
		Custodians:            wrq.custodians,
		RedeemFee:             wrq.redeemFee,
		BeaconHeight:          wrq.beaconHeight,
		TxReqID:               wrq.txReqID,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (wrq *WaitingRedeemRequest) UnmarshalJSON(data []byte) error {
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
	wrq.uniqueRedeemID = temp.UniqueRedeemID
	wrq.tokenID = temp.TokenID
	wrq.redeemerAddress = temp.RedeemerAddress
	wrq.redeemerRemoteAddress = temp.RedeemerRemoteAddress
	wrq.redeemAmount = temp.RedeemAmount
	wrq.custodians = temp.Custodians
	wrq.redeemFee = temp.RedeemFee
	wrq.beaconHeight = temp.BeaconHeight
	wrq.txReqID = temp.TxReqID
	return nil
}

func (mc MatchingRedeemCustodianDetail) GetIncognitoAddress() string {
	return mc.incAddress
}

func (mc *MatchingRedeemCustodianDetail) SetIncognitoAddress(incognitoAddress string) {
	mc.incAddress = incognitoAddress
}

func (mc MatchingRedeemCustodianDetail) GetRemoteAddress() string {
	return mc.remoteAddress
}

func (mc *MatchingRedeemCustodianDetail) SetRemoteAddress(remoteAddress string) {
	mc.remoteAddress = remoteAddress
}

func (mc MatchingRedeemCustodianDetail) GetAmount() uint64 {
	return mc.amount
}

func (mc *MatchingRedeemCustodianDetail) SetAmount(amount uint64) {
	mc.amount = amount
}

func (mc MatchingRedeemCustodianDetail) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		IncAddress    string
		RemoteAddress string
		Amount        uint64
	}{
		IncAddress:    mc.incAddress,
		RemoteAddress: mc.remoteAddress,
		Amount:        mc.amount,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (mc *MatchingRedeemCustodianDetail) UnmarshalJSON(data []byte) error {
	temp := struct {
		IncAddress    string
		RemoteAddress string
		Amount        uint64
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	mc.incAddress = temp.IncAddress
	mc.remoteAddress = temp.RemoteAddress
	mc.amount = temp.Amount
	return nil
}

func NewMatchingRedeemCustodianDetailWithValue(
	incAddress string,
	remoteAddress string,
	amount uint64) *MatchingRedeemCustodianDetail {

	return &MatchingRedeemCustodianDetail{
		incAddress:    incAddress,
		remoteAddress: remoteAddress,
		amount:        amount,
	}
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
		objectType:               WaitingRedeemRequestObjectType,
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

func GenerateWaitingRedeemRequestObjectKey(redeemID string) common.Hash {
	prefixHash := GetWaitingRedeemRequestPrefix()
	valueHash := common.HashH([]byte(redeemID))
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
