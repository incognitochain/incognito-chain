package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

type RedeemRequest struct {
	uniqueRedeemID        string
	tokenID               string
	redeemerAddress       string
	redeemerRemoteAddress string
	redeemAmount          uint64
	custodians            []*MatchingRedeemCustodianDetail
	redeemFee             uint64
	beaconHeight          uint64
	txReqID               common.Hash

	shardHeight uint64
	shardID byte
}

type MatchingRedeemCustodianDetail struct {
	incAddress    string
	remoteAddress string
	amount        uint64
}

func (rq RedeemRequest) GetUniqueRedeemID() string {
	return rq.uniqueRedeemID
}

func (rq *RedeemRequest) SetUniqueRedeemID(uniqueRedeemID string) {
	rq.uniqueRedeemID = uniqueRedeemID
}

func (rq RedeemRequest) GetTokenID() string {
	return rq.tokenID
}

func (rq *RedeemRequest) SetTokenID(tokenID string) {
	rq.tokenID = tokenID
}

func (rq RedeemRequest) GetRedeemerAddress() string {
	return rq.redeemerAddress
}

func (rq *RedeemRequest) SetRedeemerAddress(redeemerAddress string) {
	rq.redeemerAddress = redeemerAddress
}

func (rq RedeemRequest) GetRedeemerRemoteAddress() string {
	return rq.redeemerRemoteAddress
}

func (rq *RedeemRequest) SetRedeemerRemoteAddress(redeemerRemoteAddress string) {
	rq.redeemerRemoteAddress = redeemerRemoteAddress
}

func (rq RedeemRequest) GetRedeemAmount() uint64 {
	return rq.redeemAmount
}

func (rq *RedeemRequest) SetRedeemAmount(redeemAmount uint64) {
	rq.redeemAmount = redeemAmount
}

func (rq RedeemRequest) GetCustodians() []*MatchingRedeemCustodianDetail {
	return rq.custodians
}

func (rq *RedeemRequest) SetCustodians(custodians []*MatchingRedeemCustodianDetail) {
	rq.custodians = custodians
}

func (rq RedeemRequest) GetRedeemFee() uint64 {
	return rq.redeemFee
}

func (rq *RedeemRequest) SetRedeemFee(redeemFee uint64) {
	rq.redeemFee = redeemFee
}

func (rq RedeemRequest) GetBeaconHeight() uint64 {
	return rq.beaconHeight
}

func (rq *RedeemRequest) SetBeaconHeight(beaconHeight uint64) {
	rq.beaconHeight = beaconHeight
}

func (rq RedeemRequest) GetTxReqID() common.Hash {
	return rq.txReqID
}

func (rq *RedeemRequest) SetTxReqID(txReqID common.Hash) {
	rq.txReqID = txReqID
}

func (rq RedeemRequest) MarshalJSON() ([]byte, error) {
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

func (rq *RedeemRequest) UnmarshalJSON(data []byte) error {
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

func NewRedeemRequest() *RedeemRequest {
	return &RedeemRequest{}
}

func NewRedeemRequestWithValue(
	uniqueRedeemID string,
	tokenID string,
	redeemerAddress string,
	redeemerRemoteAddress string,
	redeemAmount uint64,
	custodians []*MatchingRedeemCustodianDetail,
	redeemFee uint64,
	beaconHeight uint64,
	txReqID common.Hash) *RedeemRequest {

	return &RedeemRequest{
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

type RedeemRequestObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version                  int
	waitingRedeemRequestHash common.Hash
	waitingRedeemRequest     *RedeemRequest
	objectType               int
	deleted                  bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newRedeemRequestObject(db *StateDB, hash common.Hash) *RedeemRequestObject {
	return &RedeemRequestObject{
		version:                  defaultVersion,
		db:                       db,
		waitingRedeemRequestHash: hash,
		waitingRedeemRequest:     NewRedeemRequest(),
		objectType:               WaitingRedeemRequestObjectType,
		deleted:                  false,
	}
}

func newRedeemRequestObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*RedeemRequestObject, error) {
	var redeemRequest = NewRedeemRequest()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, redeemRequest)
		if err != nil {
			return nil, err
		}
	} else {
		redeemRequest, ok = data.(*RedeemRequest)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPortalWaitingRedeemRequestType, reflect.TypeOf(data))
		}
	}
	return &RedeemRequestObject{
		version:                  defaultVersion,
		waitingRedeemRequestHash: key,
		waitingRedeemRequest:     redeemRequest,
		db:                       db,
		objectType:               WaitingRedeemRequestObjectType,
		deleted:                  false,
	}, nil
}

func GenerateWaitingRedeemRequestObjectKey(redeemID string) common.Hash {
	prefixHash := GetWaitingRedeemRequestPrefix()
	valueHash := common.HashH([]byte(redeemID))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func GenerateMatchedRedeemRequestObjectKey(redeemID string) common.Hash {
	prefixHash := GetMatchedRedeemRequestPrefix()
	valueHash := common.HashH([]byte(redeemID))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (t RedeemRequestObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *RedeemRequestObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t RedeemRequestObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *RedeemRequestObject) SetValue(data interface{}) error {
	redeemRequest, ok := data.(*RedeemRequest)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPortalWaitingRedeemRequestType, reflect.TypeOf(data))
	}
	t.waitingRedeemRequest = redeemRequest
	return nil
}

func (t RedeemRequestObject) GetValue() interface{} {
	return t.waitingRedeemRequest
}

func (t RedeemRequestObject) GetValueBytes() []byte {
	redeemRequest, ok := t.GetValue().(*RedeemRequest)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(redeemRequest)
	if err != nil {
		panic("failed to marshal redeem request")
	}
	return value
}

func (t RedeemRequestObject) GetHash() common.Hash {
	return t.waitingRedeemRequestHash
}

func (t RedeemRequestObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *RedeemRequestObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *RedeemRequestObject) Reset() bool {
	t.waitingRedeemRequest = NewRedeemRequest()
	return true
}

func (t RedeemRequestObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t RedeemRequestObject) IsEmpty() bool {
	temp := NewRedeemRequest()
	return reflect.DeepEqual(temp, t.waitingRedeemRequest) || t.waitingRedeemRequest == nil
}
